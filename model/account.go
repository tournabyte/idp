/*
 * package model describes the data types utilized by the idp service
 */

package model

import (
	"time"
)

type Account struct {
	UserId    string    `bson:"user_id"`
	Email     string    `bson:"email"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}
