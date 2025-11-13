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
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := model.ErrorResponse{
			Reason:  "PROCESSING_REQUEST_FAILED",
			Message: "Request body could not be read",
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorRespnse := model.ErrorResponse{
			Reason:  "BODY_DECODE_FAILED",
			Message: "Request body did not contain valid JSON",
		}
		json.NewEncoder(w).Encode(errorRespnse)
		return
	}

	val := r.Context().Value("CONN")
	var accounts *model.TournabyteAccountRepository
	if conn, ok := val.(*mongo.Database); ok {
		accounts = model.NewTournabyteAccountRepository(conn)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := model.ErrorResponse{
			Reason:  "PROCESSING_REQUEST_FAILED",
			Message: "Service encountered an unexpected database error",
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	newAccount := model.Account{
		Email: payload.NewAccountEmail,
	}
	if create_err := accounts.Create(ctx, &newAccount); create_err != nil {
		w.WriteHeader(http.StatusForbidden)
		errorResponse := model.ErrorResponse{
			Reason:  "REQUEST_NOT_PERMITTED",
			Message: create_err.Error(),
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newAccount)
}

func getAccount(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idPathParameter := r.PathValue("id")
	if idPathParameter == "" {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := model.ErrorResponse{
			Reason:  "PROCESSING_REQUEST_FAILED",
			Message: "No parameter provided",
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	val := r.Context().Value("CONN")
	var accounts *model.TournabyteAccountRepository
	if conn, ok := val.(*mongo.Database); ok {
		accounts = model.NewTournabyteAccountRepository(conn)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := model.ErrorResponse{
			Reason:  "PROCESSING_REQUEST_FAILED",
			Message: "Service encountered an unexpected database error",
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	account, findErr := accounts.FindById(ctx, idPathParameter)
	if findErr != nil {
		w.WriteHeader(http.StatusNotFound)
		errorResponse := model.ErrorResponse{
			Reason:  "REQUEST_NOT_PERMITTED",
			Message: findErr.Error(),
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)

}
