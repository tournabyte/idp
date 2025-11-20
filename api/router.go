/*
 * package api defines the server net/http server instance used for processing requests for idp service
 */
package api

import (
	"context"
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
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type OperationResult struct {
}

type TournabyteIdentityProviderService struct {
	db  *mongo.Client
	mux *http.ServeMux
}

var listenOn int

type TournabyteIdentityProviderOperation func(ctx context.Context) OperationResult

func NewIdentityProviderServer(opts model.CommandOpts) (*TournabyteIdentityProviderService, error) {
	var tbyteService TournabyteIdentityProviderService
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if conn_err := tbyteService.connectDatabase(opts.Dbhosts, opts.Dbname, opts.Dbuser, opts.Dbpass); conn_err != nil {
		return nil, fmt.Errorf("Could not connect to database: %w", conn_err)
	}

	if ping_err := tbyteService.pingDatabase(ctx); ping_err != nil {
		return nil, fmt.Errorf("Database unreachable: %w", ping_err)
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
	provider.mux.HandleFunc(CREATE_ACCOUNT_ENDPOINT, func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Handler for `%s` invoked", CREATE_ACCOUNT_ENDPOINT)
		time.Sleep(3 * time.Second)
		w.Write([]byte(fmt.Sprintf("Response received from `%s` handler", CREATE_ACCOUNT_ENDPOINT)))
	})
	provider.mux.HandleFunc(LOOKUP_ACCOUNT_ENDPOINT, func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Handler for `%s` invoked", LOOKUP_ACCOUNT_ENDPOINT)
		time.Sleep(3 * time.Second)
		w.Write([]byte(fmt.Sprintf("Response received from `%s` handler", LOOKUP_ACCOUNT_ENDPOINT)))
	})
}

/*
	func (provider *TournabyteIdentityProviderService) getAccount(ctx context.Context) OperationResult {
		return OperationResult{}
	}

	func (provider *TournabyteIdentityProviderService) makeAccount(ctx context.Context) OperationResult {
		return OperationResult{}
	}
*/

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
		log.Fatalf("Server exited ungracefully: %v", err)
	}
	log.Println("Server exited gracefully")
}
