/*
 * package api defines the server net/http server instance used for processing requests for idp service
 */
package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/tournabyte/idp/model"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func createAccount(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var payload model.CreateAccountRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Body did not contain valid JSON", http.StatusBadRequest)
		return
	}

	val := r.Context().Value("CONN")
	var accounts *model.TournabyteAccountRepository
	if conn, ok := val.(*mongo.Database); ok {
		accounts = model.NewTournabyteAccountRepository(conn)
	} else {
		http.Error(w, "Database Error", http.StatusBadGateway)
		return
	}

	newAccount := model.Account{
		Email: payload.NewAccountEmail,
	}
	if create_err := accounts.Create(ctx, &newAccount); create_err != nil {
		http.Error(w, create_err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Successfully created the account"))
}

func getAccount(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idPathParameter := r.PathValue("id")
	if idPathParameter == "" {
		http.Error(w, "No parameter provided", http.StatusBadRequest)
		return
	}

	val := r.Context().Value("CONN")
	var accounts *model.TournabyteAccountRepository
	if conn, ok := val.(*mongo.Database); ok {
		accounts = model.NewTournabyteAccountRepository(conn)
	} else {
		http.Error(w, "Database Error", http.StatusBadGateway)
		return
	}

	account, findErr := accounts.FindById(ctx, idPathParameter)
	if findErr != nil {
		http.Error(w, findErr.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)

}
