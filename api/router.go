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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/tournabyte/idp/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type TournabyteIdentityProviderService struct {
	db                 *mongo.Client
	mux                *http.ServeMux
	env                *model.ApplicationOptions
	sessionTokenSigner jose.Signer
}

func NewIdentityProviderServer(opts *model.ApplicationOptions) (*TournabyteIdentityProviderService, error) {
	var tbyteService TournabyteIdentityProviderService
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tbyteService.env = opts
	tbyteService.configureHandlers()

	if connErr := tbyteService.connectDatabase(); connErr != nil {
		return nil, fmt.Errorf("Could not connect to database: %w", connErr)
	}

	if pingErr := tbyteService.pingDatabase(ctx); pingErr != nil {
		return nil, fmt.Errorf("Database unreachable: %w", pingErr)
	}

	if signErr := tbyteService.initializeTokenSigner(); signErr != nil {
		return nil, fmt.Errorf("Failed to create token signer: %w", signErr)
	}

	return &tbyteService, nil
}

func (provider *TournabyteIdentityProviderService) initializeTokenSigner() error {
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.HS256, Key: []byte(provider.env.Serve.WebToken.Key)},
		nil,
	)
	if err != nil {
		return err
	}

	provider.sessionTokenSigner = signer
	return nil
}

func (provider *TournabyteIdentityProviderService) connectDatabase() error {
	var connectOptions options.ClientOptions

	connectOptions.SetHosts(provider.env.Datastore.Hosts)
	connectOptions.SetAuth(options.Credential{
		Username:    provider.env.Datastore.Username,
		Password:    provider.env.Datastore.Password,
		PasswordSet: true,
	})

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

	provider.mux.HandleFunc(
		AUTHORIZE_LOGIN,
		SetRequestTimeout(ReadRequestBodyAsJSON[model.LoginAttempt](provider.authorizeAccount), 30),
	)

}

func (provider *TournabyteIdentityProviderService) Run() {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", provider.env.Serve.Port),
		Handler: provider.mux,
	}

	go func() {
		log.Printf("Starting server on port %d", provider.env.Serve.Port)
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
		accountsCollectionHandle := model.NewTournabyteAccountRepository(
			provider.db.Database("idp").Collection("accounts"),
		)
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
					account.BasicInfo(),
				))
			EmitResponseAsJSON[model.BasicAccountInfoResponse](w, r)
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
	if newAccountDetails, ok := r.Context().Value(DECODED_JSON_BODY).(model.CreateAccountRequest); ok {
		accountsCollectionHandle := model.NewTournabyteAccountRepository(
			provider.db.Database("idp").Collection("accounts"),
		)
		newAccountRecord := model.Account{
			Email:    newAccountDetails.NewAccountEmail,
			LoginKey: provider.mustHashPassword(newAccountDetails.NewAccountPassword),
		}
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

func (provider *TournabyteIdentityProviderService) authorizeAccount(w http.ResponseWriter, r *http.Request) {
	if loginAttempt, ok := r.Context().Value(DECODED_JSON_BODY).(model.LoginAttempt); ok {
		accountsCollectionHandle := model.NewTournabyteAccountRepository(
			provider.db.Database("idp").Collection("accounts"),
		)
		if acc, err := accountsCollectionHandle.FindByEmail(r.Context(), loginAttempt.LoginId); err != nil {
			log.Printf("No account found with email: %s", loginAttempt.LoginId)
			r = r.WithContext(
				context.WithValue(r.Context(), HANDLER_STATUS_CODE, http.StatusForbidden),
			)
			r = r.WithContext(
				context.WithValue(
					r.Context(),
					HANDLER_RESPONSE_BODY,
					model.ErrorResponse{Reason: "NO_MATCHING_RESOURCE", Message: "Invalid email or password"},
				))
			defer RecoverResponse(w, r)
			panic("Invalid log in attempt")

		} else {
			if match, err := argon2id.ComparePasswordAndHash(loginAttempt.LoginSecret, acc.LoginKey); err != nil {
				log.Printf("Error during password comparison: %v", err)
				r = r.WithContext(
					context.WithValue(r.Context(), HANDLER_STATUS_CODE, http.StatusUnauthorized),
				)
				r = r.WithContext(
					context.WithValue(
						r.Context(),
						HANDLER_RESPONSE_BODY,
						model.ErrorResponse{Reason: "NO_MATCHING_RESOURCE", Message: "Invalid email or password"},
					))
				defer RecoverResponse(w, r)
				panic("Invalid log in attempt")
			} else if !match {
				log.Printf("Comparison succeeded but no match found")
				r = r.WithContext(
					context.WithValue(r.Context(), HANDLER_STATUS_CODE, http.StatusUnauthorized),
				)
				r = r.WithContext(
					context.WithValue(
						r.Context(),
						HANDLER_RESPONSE_BODY,
						model.ErrorResponse{Reason: "NO_MATCHING_RESOURCE", Message: "Invalid email or password"},
					))
				defer RecoverResponse(w, r)
				panic("Invalid log in attempt")
			} else {
				log.Printf("Comparison succeeded and match detected")
				r = r.WithContext(
					context.WithValue(r.Context(), HANDLER_STATUS_CODE, http.StatusCreated),
				)
				r = r.WithContext(
					context.WithValue(
						r.Context(),
						HANDLER_RESPONSE_BODY,
						model.SuccessfulAuthenticationResponse{Token: provider.makeSessionToken(acc.Id.String())},
					))
				EmitResponseAsJSON[model.SuccessfulAuthenticationResponse](w, r)

			}
		}
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
		panic("Invalid log in attempt")

	}
}

func (provider *TournabyteIdentityProviderService) mustHashPassword(passwd string) string {
	hash, err := argon2id.CreateHash(passwd, argon2id.DefaultParams)
	if err != nil {
		panic(fmt.Sprintf("Hashing failed: %v", err))
	}
	return hash
}

func (provider *TournabyteIdentityProviderService) makeSessionToken(userId string) string {
	cl := jwt.Claims{
		Issuer:   "example.com",
		Audience: jwt.Audience{"example-audience"},
		Expiry:   jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt: jwt.NewNumericDate(time.Now()),
		ID:       userId,
	}

	raw, err := jwt.Signed(provider.sessionTokenSigner).Claims(cl).Serialize()
	if err != nil {
		panic(fmt.Sprintf("JWT creation failed: %v", err))
	}
	return raw
}
