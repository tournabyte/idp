/*
 * package model describes the data types utilized by the idp service
 */

package model

type CreateAccountRequest struct {
	NewAccountEmail string `json:"email"`
}
