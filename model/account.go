/*
 * package model describes the data types utilized by the idp service
 */

package model

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Account struct {
	Id    bson.ObjectID `bson:"_id,omitempty"`
	Email string        `bson:"email"`
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
	_, err := r.collection.InsertOne(ctx, account)
	return err
}

func (r *TournabyteAccountRepository) FindById(ctx context.Context, idHex string) (*Account, error) {
	var account Account
	var filter bson.D

	oid, convertIdErr := bson.ObjectIDFromHex(idHex)

	if convertIdErr != nil {
		return nil, convertIdErr
	}

	filter = bson.D{{Key: "_id", Value: oid}}
	findDocumentErr := r.collection.FindOne(ctx, filter).Decode(&account)
	if findDocumentErr == mongo.ErrNoDocuments {
		return nil, findDocumentErr
	}
	return &account, nil
}
