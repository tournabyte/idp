/*
 * package model describes the data types utilized by the idp service
 */

package model

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MockCollectionHandle struct {
	mock.Mock
}

func (m *MockCollectionHandle) InsertOne(ctx context.Context, doc any, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, doc)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *MockCollectionHandle) FindOne(ctx context.Context, filter any, opts ...*options.FindOneOptions) *mongo.SingleResult {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.SingleResult)
}

type AccountRepositoryOperationsTestSuite struct {
	suite.Suite
	repo TournabyteAccountRepository
}

func TestAccountRepositoryOperations(t *testing.T) {
	suite.Run(t, new(AccountRepositoryOperationsTestSuite))
}

func (s *AccountRepositoryOperationsTestSuite) TestCreate() {
	ctx := context.TODO()
	account := Account{Email: "test@example.io"}
	insertedID := bson.NewObjectID()

	mockCollection := new(MockCollectionHandle)
	mockCollection.On("InsertOne", ctx, &account).Return(&mongo.InsertOneResult{InsertedID: insertedID}, nil)
	s.repo = *NewTournabyteAccountRepository(mockCollection)

	err := s.repo.Create(ctx, &account)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), insertedID, account.Id)
	mockCollection.AssertExpectations(s.T())

}

func (s *AccountRepositoryOperationsTestSuite) TestFindById_Success() {
	ctx := context.TODO()
	oid := bson.NewObjectID()
	filter := bson.D{{Key: "_id", Value: oid}}
	want := Account{Id: oid, Email: "test@example.com"}

	mockCollection := new(MockCollectionHandle)
	mockCollection.On("FindOne", ctx, filter).Return(mongo.NewSingleResultFromDocument(&want, nil, nil))
	s.repo = *NewTournabyteAccountRepository(mockCollection)

	acc, err := s.repo.FindById(ctx, oid.Hex())

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), &want, acc)
	mockCollection.AssertExpectations(s.T())
}

func (s *AccountRepositoryOperationsTestSuite) TestFindById_NotFound() {
	ctx := context.TODO()
	oid := bson.NewObjectID()
	filter := bson.D{{Key: "_id", Value: oid}}
	want := Account{Id: oid, Email: "test@example.com"}

	mockCollection := new(MockCollectionHandle)
	expected := mongo.NewSingleResultFromDocument(&want, mongo.ErrNoDocuments, nil)
	mockCollection.On("FindOne", ctx, filter).Return(expected)
	s.repo = *NewTournabyteAccountRepository(mockCollection)

	acc, err := s.repo.FindById(ctx, oid.Hex())

	assert.Nil(s.T(), acc)
	assert.True(s.T(), errors.Is(err, mongo.ErrNoDocuments))
	mockCollection.AssertExpectations(s.T())
}

func (s *AccountRepositoryOperationsTestSuite) TestFindById_InvalidHex() {
	ctx := context.TODO()
	mockCollection := new(MockCollectionHandle)
	s.repo = *NewTournabyteAccountRepository(mockCollection)

	acc, err := s.repo.FindById(ctx, "not-a-hex")

	assert.Nil(s.T(), acc)
	assert.True(s.T(), errors.Is(err, bson.ErrInvalidHex))
}
