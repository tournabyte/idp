/*
 * package model describes the data types utilized by the idp service
 */

package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Account struct {
	UserId    string    `bson:"user_id"`
	Email     string    `bson:"email"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

type TournabyteAccountRepository struct {
	collection *mongo.Collection
}

func NewTournabyteAccountRepository(db *mongo.Database) *TournabyteAccountRepository {
	return &TournabyteAccountRepository{
		db.Collection("accounts"),
	}
}

func (r *TournabyteAccountRepository) Create(ctx context.Context, account *Account) error {
	now := time.Now()
	account.CreatedAt = now
	account.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, account)
	return err
}

func (r *TournabyteAccountRepository) FindOne(ctx context.Context, filter bson.M) (*Account, error) {
	var account Account

	err := r.collection.FindOne(ctx, filter).Decode(&account)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &account, err
}
