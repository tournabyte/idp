/*
 * package api defines the server net/http server instance used for processing requests for idp service
 */
package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/tournabyte/idp/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type TournabyteIdentityProviderService struct {
	db  *mongo.Client
	mux *http.ServeMux
}

var listenOn int

func NewIdentityProviderServer(opts model.CommandOpts) (*TournabyteIdentityProviderService, error) {
	var tbyteService TournabyteIdentityProviderService
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if connErr := tbyteService.connectDatabase(opts.Dbhosts, opts.Dbname, opts.Dbuser, opts.Dbpass); connErr != nil {
		return nil, fmt.Errorf("Could not connect to database: %w", connErr)
	}

	if pingErr := tbyteService.pingDatabase(ctx); pingErr != nil {
		return nil, fmt.Errorf("Database unreachable: %w", pingErr)
	}

	tbyteService.configureHandlers()
	listenOn = opts.Port

	return &tbyteService, nil
}

func (provider *TournabyteIdentityProviderService) connectDatabase(
	hosts []string,
	database string,
	username string,
	password string,
) error {
	var uri string
	var connectOptions options.ClientOptions

	if len(hosts) == 0 {
		return fmt.Errorf("at least one host must be provided")
	}

	hostList := strings.Join(hosts, ",")
	credentials := fmt.Sprintf(
		"%s:%s",
		url.QueryEscape(username),
		url.QueryEscape(password),
	)

	if database != "" {
		uri = fmt.Sprintf("mongodb://%s@%s/%s", credentials, hostList, database)
	} else {
		uri = fmt.Sprintf("mongodb://%s@%s", credentials, hostList)
	}

	connectOptions.ApplyURI(uri)
	connectOptions.Auth.AuthSource = "admin"
	log.Printf("Using `%s` as the database", connectOptions.GetURI())
	conn, conn_err := mongo.Connect(&connectOptions)
	if conn_err != nil {
		return fmt.Errorf("failed to initialize the mongo client: %w", conn_err)
	}

	provider.db = conn
	return nil
}

func (provider *TournabyteIdentityProviderService) pingDatabase(ctx context.Context) error {
	if provider.db == nil {
		return fmt.Errorf("cannot ping a null deployment")
	}

	return provider.db.Ping(ctx, readpref.Primary())
}

func (provider *TournabyteIdentityProviderService) configureHandlers() {
	provider.mux = http.NewServeMux()
	provider.mux.HandleFunc(
		CREATE_ACCOUNT_ENDPOINT,
		SetRequestTimeout(ReadRequestBodyAsJSON[model.CreateAccountRequest](provider.createAccount), 30),
	)

	provider.mux.HandleFunc(
		LOOKUP_ACCOUNT_ENDPOINT,
		SetRequestTimeout(ExtractPathParameters(provider.findAccountById, "id"), 30),
	)

}

func (provider *TournabyteIdentityProviderService) Run() {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", listenOn),
		Handler: provider.mux,
	}

	go func() {
		log.Printf("Starting server on port %d", listenOn)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start service: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		server.Close()
		log.Fatalf("Could not shutdown the server, forcing it anyway %v", err)
	}
	log.Println("Server exited gracefully")
}

func (provider *TournabyteIdentityProviderService) findAccountById(w http.ResponseWriter, r *http.Request) {
	if idHex, ok := r.Context().Value(PATH_VALUE_MAPPING).(map[string]string)["id"]; ok {
		accountsCollectionHandle := model.NewTournabyteAccountRepository(provider.db.Database("idp"))
		account, findErr := accountsCollectionHandle.FindById(r.Context(), idHex)

		switch {
		case errors.Is(findErr, mongo.ErrNoDocuments):
			r = r.WithContext(
				context.WithValue(r.Context(), HANDLER_STATUS_CODE, http.StatusNotFound),
			)
			r = r.WithContext(
				context.WithValue(
					r.Context(),
					HANDLER_RESPONSE_BODY,
					model.ErrorResponse{Reason: "NO_MATCHING_RESOURCE", Message: "No resource found for the given object ID"},
				))
			defer RecoverResponse(w, r)
			panic("Resource not found")

		case errors.Is(findErr, bson.ErrInvalidHex):
			r = r.WithContext(
				context.WithValue(r.Context(), HANDLER_STATUS_CODE, http.StatusBadRequest),
			)
			r = r.WithContext(
				context.WithValue(
					r.Context(),
					HANDLER_RESPONSE_BODY,
					model.ErrorResponse{Reason: "PATH_PARAMETER_MALFORMED", Message: "Given hex is not a valid object ID"},
				))
			defer RecoverResponse(w, r)
			panic("ID parameter invalid")

		default:
			r = r.WithContext(
				context.WithValue(r.Context(), HANDLER_STATUS_CODE, http.StatusOK),
			)
			r = r.WithContext(
				context.WithValue(
					r.Context(),
					HANDLER_RESPONSE_BODY,
					*account,
				))
			EmitResponseAsJSON[model.Account](w, r)
		}

	} else {
		r = r.WithContext(
			context.WithValue(r.Context(), HANDLER_STATUS_CODE, http.StatusNotFound),
		)
		r = r.WithContext(
			context.WithValue(
				r.Context(),
				HANDLER_RESPONSE_BODY,
				model.ErrorResponse{Reason: "PATH_PARAMETER_NOT_PRESENT", Message: "Required dynamic path part not present"},
			))
		defer RecoverResponse(w, r)
		panic("Required dynamic path part not present")

	}
}

func (provider *TournabyteIdentityProviderService) createAccount(w http.ResponseWriter, r *http.Request) {
	if deadline, ok := r.Context().Deadline(); ok {
		log.Printf("Time remaining: %d", time.Until(deadline))
	} else {
		log.Printf("Deadline already exceeded")
	}
	if newAccountDetails, ok := r.Context().Value(DECODED_JSON_BODY).(model.CreateAccountRequest); ok {
		accountsCollectionHandle := model.NewTournabyteAccountRepository(provider.db.Database("idp"))
		newAccountRecord := model.Account{Email: newAccountDetails.NewAccountEmail}
		if createErr := accountsCollectionHandle.Create(r.Context(), &newAccountRecord); createErr != nil {
			log.Printf("Did not create the account: %v", createErr)
			r = r.WithContext(
				context.WithValue(r.Context(), HANDLER_STATUS_CODE, http.StatusInternalServerError),
			)
			r = r.WithContext(
				context.WithValue(
					r.Context(),
					HANDLER_RESPONSE_BODY,
					model.ErrorResponse{Reason: "ACCOUNT_NOT_CREATED", Message: "Did not create the requested account"},
				))
			defer RecoverResponse(w, r)
			panic("Account creation failed")

		}
		log.Printf("Created the account: %v", newAccountRecord)
		r = r.WithContext(
			context.WithValue(r.Context(), HANDLER_STATUS_CODE, http.StatusCreated),
		)
		r = r.WithContext(
			context.WithValue(
				r.Context(),
				HANDLER_RESPONSE_BODY,
				newAccountRecord,
			))
		EmitResponseAsJSON[model.Account](w, r)

	} else {
		r = r.WithContext(
			context.WithValue(r.Context(), HANDLER_STATUS_CODE, http.StatusBadRequest),
		)
		r = r.WithContext(
			context.WithValue(
				r.Context(),
				HANDLER_RESPONSE_BODY,
				model.ErrorResponse{Reason: "INVALID_JSON_BODY", Message: "Required body is not present or incorrectly structured"},
			))
		defer RecoverResponse(w, r)
		panic("Required dynamic path part not present")

	}
}
