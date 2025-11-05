/*
 * package api defines the server net/http server instance used for processing requests for idp service
 */
package api

import (
	"fmt"
	"net/http"
)

func createAccount(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create account handler invoked!")
}

func getAccount(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get account handler invoked!")
}
