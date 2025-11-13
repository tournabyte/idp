/*
 * package model describes the data types utilized by the idp service
 */

package model

type CreateAccountRequest struct {
	NewAccountEmail string `json:"email"`
}

type ErrorResponse struct {
	Reason  string `json:"reason"`
	Message string `json:"err_msg"`
}
