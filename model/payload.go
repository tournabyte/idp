/*
 * package model describes the data types utilized by the idp service
 */

package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CreateAccountRequest struct {
	NewAccountEmail    string `json:"email"`
	NewAccountPassword string `json:"password"`
}

type BasicAccountInfoResponse struct {
	AccountIdentifier  bson.ObjectID `json:"id"`
	AccountContact     string        `json:"email"`
	AccountCreatedTime time.Time     `json:"created"`
	AccountModifiedAt  time.Time     `json:"modified"`
}

type ErrorResponse struct {
	Reason  string `json:"reason"`
	Message string `json:"err_msg"`
}

type LoginAttempt struct {
	LoginId     string `json:"authenticate_as"`
	LoginSecret string `json:"passphrase"`
}

type SuccessfulAuthenticationResponse struct {
	Token string `json:"token"`
}
