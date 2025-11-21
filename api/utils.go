/*
 * package api defines the server net/http server instance used for processing requests for idp service
 */
package api

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/tournabyte/idp/model"
)

const (
	CREATE_ACCOUNT_ENDPOINT = "POST /accounts"
	LOOKUP_ACCOUNT_ENDPOINT = "GET /accounts/{id}"
)

type RequestContextKey string

const (
	DECODED_JSON_BODY     = "DECODED_BODY_VALUE"
	PATH_VALUE_MAPPING    = "PATH_PARAMETERS"
	QUERY_VALUE_MAPPING   = "QUERY_PARAMETERS"
	HANDLER_RESPONSE_BODY = "RESPONSE_BODY"
	HANDLER_STATUS_CODE   = "RESPONSE_STATUS"
)

type HandlerFuncProcessingStep func(http.HandlerFunc) http.HandlerFunc

func RecoverResponse(w http.ResponseWriter, r *http.Request) {
	if err := recover(); err != nil {
		log.Printf("Recovering from a fatal error from the http.HandlerFunc sequence: %v", err)
		emitter := json.NewEncoder(w)
		ctx := r.Context()

		errStatusCode, okGotErrorStatusCode := ctx.Value(HANDLER_STATUS_CODE).(int)
		if !okGotErrorStatusCode {
			log.Printf("Expected an integer http status code to report, got %T", errStatusCode)
			w.WriteHeader(http.StatusInternalServerError)
			emitter.Encode(map[string]string{"Bad context": "Internal Server Error"})
			return
		}

		errorResponse, okGotErrorResponse := ctx.Value(HANDLER_RESPONSE_BODY).(model.ErrorResponse)
		if !okGotErrorResponse {
			log.Printf("Expected an error response struct to emit, got %v", errorResponse)
			w.WriteHeader(http.StatusInternalServerError)
			emitter.Encode(map[string]string{"Bad context": "Internal Server Error"})
			return
		}

		w.WriteHeader(errStatusCode)
		emitter.Encode(errorResponse)
	}
}

func SetRequestTimeout(next http.HandlerFunc, seconds int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Allowing up to %d seconds for processing this request", seconds)
		ctx, cancel := context.WithTimeout(r.Context(), time.Duration(seconds)*time.Second)
		defer cancel()

		go func() {

		}()
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func ReadRequestBodyAsJSON[T any](next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var decodedPayload T
		body, decodeErr := io.ReadAll(r.Body)
		log.Printf("Looking to decode request body to %T", decodedPayload)
		if decodeErr != nil {
			log.Printf("Failed to read request body contents, updating request context with an error response")
			r = r.WithContext(
				context.WithValue(r.Context(), HANDLER_STATUS_CODE, http.StatusBadRequest),
			)
			r = r.WithContext(
				context.WithValue(
					r.Context(),
					HANDLER_RESPONSE_BODY,
					model.ErrorResponse{Reason: "BODY_DECODE_FAILED", Message: "Request body could not be read"},
				),
			)
			defer RecoverResponse(w, r)
			panic("Failed to read request body contents")
		}
		defer r.Body.Close()

		if unmarshalErr := json.Unmarshal(body, &decodedPayload); unmarshalErr != nil {
			log.Printf("Failed to unmarshal the request body contents as JSON, updating the request context with an error response")
			r = r.WithContext(
				context.WithValue(r.Context(), HANDLER_STATUS_CODE, http.StatusBadRequest),
			)
			r = r.WithContext(
				context.WithValue(
					r.Context(),
					HANDLER_RESPONSE_BODY,
					model.ErrorResponse{Reason: "BODY_DECODE_FAILED", Message: "Request body did not contain valid JSON"},
				))
			defer RecoverResponse(w, r)
			panic("Failed to parse request body contents as JSON")
		}

		log.Printf("Request body decoding completed, updating the request context with the resulting data")
		next.ServeHTTP(w, r.WithContext(
			context.WithValue(
				r.Context(),
				DECODED_JSON_BODY,
				decodedPayload,
			)))
	}
}

func ExtractPathParameters(next http.HandlerFunc, parts ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathParamMap := make(map[string]string)

		for _, p := range parts {
			if val := r.PathValue(p); val == "" {
				log.Printf("Expected a dynamic value for %s but found none", p)
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

			} else {
				log.Printf("Found dynamic path value %s=%s", p, val)
				pathParamMap[p] = val
			}
		}

		r = r.WithContext(
			context.WithValue(
				r.Context(),
				PATH_VALUE_MAPPING,
				pathParamMap,
			))
		next(w, r)
	}
}

func EmitResponseAsJSON[ResponseType any](w http.ResponseWriter, r *http.Request) {

	log.Printf("Beginning JSON response construction")
	emitter := json.NewEncoder(w)
	ctx := r.Context()
	defer RecoverResponse(w, r)

	responseCode, okGotStatusCode := ctx.Value(HANDLER_STATUS_CODE).(int)
	if !okGotStatusCode {
		panic("Handler chain did not set status code correctly")
	}

	responseBody, okGotResponseBody := ctx.Value(HANDLER_RESPONSE_BODY).(ResponseType)
	if !okGotResponseBody {
		panic("Handler chain did not prepare body as expected")
	}

	log.Printf("Gathered response elements needed... emitting as JSON")
	w.WriteHeader(responseCode)
	emitter.Encode(responseBody)
}
