/*
* package api defines the server net/http server instance used for processing requests for idp service
 */
package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type TournabyteIdentityProviderService struct {
	db  *mongo.Client
	mux *http.ServeMux
}

func NewIdentityProviderServer() (*TournabyteIdentityProviderService, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Client()
	conn, conn_err := mongo.Connect(opts)

	if conn_err != nil {
		return nil, conn_err
	}

	if ping_err := conn.Ping(ctx, readpref.Primary()); ping_err != nil {
		return nil, ping_err
	}

	log.Printf("Using `%s` as the database", opts.GetURI())

	return &TournabyteIdentityProviderService{
		db:  conn,
		mux: http.NewServeMux(),
	}, nil
}

func (server *TournabyteIdentityProviderService) AddHandler(route string, handler http.HandlerFunc) {
	server.mux.HandleFunc(route, handler)
}

func (server *TournabyteIdentityProviderService) RunServer(port int) error {
	listener := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: server.mux,
	}

	return listener.ListenAndServe()
}

func (server *TournabyteIdentityProviderService) ConfigureServer() *TournabyteIdentityProviderService {
	server.AddHandler("POST /accounts", createAccount)
	server.AddHandler("GET /accounts", getAccount)
	return server
}
