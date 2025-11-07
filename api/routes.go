/*
* package api defines the server net/http server instance used for processing requests for idp service
 */
package api

import (
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type TournabyteIdentityProviderService struct {
	db  *mongo.Client
	mux *http.ServeMux
}

func NewIdentityProviderServer(dbhost, dbuser, dbpass string) (*TournabyteIdentityProviderService, error) {

	uri := fmt.Sprintf("mongodb://%s:%s@%s", dbuser, dbpass, dbhost)
	opts := options.Client().ApplyURI(uri)

	conn, conn_err := mongo.Connect(opts)

	if conn_err != nil {
		return nil, conn_err
	}

	log.Printf("Using %s as the database", uri)

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
