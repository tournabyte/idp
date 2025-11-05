/*
* package api defines the server net/http server instance used for processing requests for idp service
 */
package api

import (
	"net/http"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type TournabyteIdentityProviderService struct {
	db  *mongo.Client
	mux *http.ServeMux
}

func NewIdentityProviderServer(addr string) *TournabyteIdentityProviderService {
	return &TournabyteIdentityProviderService{
		db:  nil,
		mux: http.NewServeMux(),
	}
}

func (server *TournabyteIdentityProviderService) AddHandler(route string, handler http.HandlerFunc) {
	server.mux.HandleFunc(route, handler)
}

func (server *TournabyteIdentityProviderService) RunServer() error {
	listener := &http.Server{
		Addr:    ":8080",
		Handler: server.mux,
	}

	return listener.ListenAndServe()
}

func (server *TournabyteIdentityProviderService) ConfigureServer() *TournabyteIdentityProviderService {
	server.AddHandler("POST /accounts", createAccount)
	server.AddHandler("GET /accounts", getAccount)
	return server
}
