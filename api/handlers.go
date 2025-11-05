/*
 * package api defines the server net/http server instance used for processing requests for idp service
 */
package api

import (
	"log"
	"net/http"
)

func createAccount(w http.ResponseWriter, r *http.Request) {
	log.Println("Create account handler invoked!")
	w.Write([]byte("Account creation handled\n"))
}

func getAccount(w http.ResponseWriter, r *http.Request) {
	log.Println("Get account handler invoked!")
	w.Write([]byte("Account lookup handled\n"))
}
