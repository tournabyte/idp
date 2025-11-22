/*
 * package model describes the data types utilized by the idp service
 */

package model

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Account struct {
	Id    bson.ObjectID `bson:"_id,omitempty"`
	Email string        `bson:"email"`
}

type InsertOneDocumment interface {
	InsertOne(ctx context.Context, doc any, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
}

type FindOneDocument interface {
	FindOne(ctx context.Context, filter any, opts ...*options.FindOneOptions) *mongo.SingleResult
}

type CreateAndReadOneDocument interface {
	InsertOneDocumment
	FindOneDocument
}

type TournabyteAccountRepository struct {
	collection CreateAndReadOneDocument
}

func NewTournabyteAccountRepository(col CreateAndReadOneDocument) *TournabyteAccountRepository {
	return &TournabyteAccountRepository{collection: col}
}

func (r *TournabyteAccountRepository) Create(ctx context.Context, account *Account) error {
	result, err := r.collection.InsertOne(ctx, account)
	if err != nil {
		return err
	}
	account.Id = result.InsertedID.(bson.ObjectID)
	return nil
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
