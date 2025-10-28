/*
* package api defines the server net/http server instance used for processing requests for idp service
 */
package api

import (
	"fmt"
	"net/http"
)

func NewIdentityProviderServer(addr string) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /accounts", createAccount)
	mux.HandleFunc("GET /accounts", getAccount)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}

func createAccount(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create account handler invoked!")
}

func getAccount(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get account handler invoked!")
}
