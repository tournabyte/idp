/*
 * package api defines the server net/http server instance used for processing requests for idp service
 */
package api

import (
	"context"
	"log"
	"net/http"
	"unicode"

	"github.com/tournabyte/idp/model"
)

type PalindromeCheckRequest struct {
	Query string `json:"s"`
}

type PalindromeCheckResponse struct {
	Input  string `json:"input"`
	Result bool   `json:"result"`
}

func PalindromeCheck(emitResponse http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if requestBody, ok := r.Context().Value(DECODED_JSON_BODY).(PalindromeCheckRequest); !ok {
			log.Printf("Body not present or of incorrect type %T, updating the request context with an error response", requestBody)
			r = r.WithContext(
				context.WithValue(
					r.Context(),
					HANDLER_STATUS_CODE,
					http.StatusBadRequest,
				))
			r = r.WithContext(
				context.WithValue(
					r.Context(),
					HANDLER_RESPONSE_BODY,
					model.ErrorResponse{Reason: "NO_INPUT_PROVIDED", Message: "Handler requires an input to check"},
				))
			emitResponse(w, r)
			defer RecoverResponse(w, r)
			panic("Missing context: expected body")
		} else {
			log.Printf("Body obtained from request body, processing the data")
			s := []rune(requestBody.Query)
			i, j := 0, len(s)-1

			for i < j {
				for i < j && !unicode.IsLetter(s[i]) && !unicode.IsDigit(s[i]) {
					i++
				}
				for i < j && !unicode.IsLetter(s[j]) && !unicode.IsDigit(s[j]) {
					j--
				}

				if unicode.ToLower(s[i]) != unicode.ToLower(s[j]) {
					log.Printf("Violation found with input, updating the request context with a false response")
					r = r.WithContext(
						context.WithValue(
							r.Context(),
							HANDLER_STATUS_CODE,
							http.StatusOK,
						))
					r = r.WithContext(
						context.WithValue(
							r.Context(),
							HANDLER_RESPONSE_BODY,
							PalindromeCheckResponse{Input: requestBody.Query, Result: false},
						))
					emitResponse(w, r)
					return
				}
				i++
				j--
			}

			log.Printf("No violations found with the input, updating the request context with a true response")
			r = r.WithContext(
				context.WithValue(
					r.Context(),
					HANDLER_STATUS_CODE,
					http.StatusOK,
				))
			r = r.WithContext(
				context.WithValue(
					r.Context(),
					HANDLER_RESPONSE_BODY,
					PalindromeCheckResponse{Input: requestBody.Query, Result: true},
				))
			log.Printf("ctx.Value(HANDLER_STATUS_CODE) = %#v", r.Context().Value(HANDLER_STATUS_CODE))
			emitResponse(w, r)

		}
	}
}
