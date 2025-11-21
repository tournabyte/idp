/*
 * package model describes the data types utilized by the idp service
 */

package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CreateAccountRequestDecodeTestSuite struct {
	suite.Suite
	stream []byte
}

func TestCreateAccountRequestDecoding(t *testing.T) {
	suite.Run(t, new(CreateAccountRequestDecodeTestSuite))
}

func (s *CreateAccountRequestDecodeTestSuite) SetupTest() {
	s.stream = []byte("{\"email\":\"test@example.com\"}")
}

func (s *CreateAccountRequestDecodeTestSuite) TestDecodeToStructuredData() {
	var result CreateAccountRequest
	if assert.Nil(s.T(), json.Unmarshal(s.stream, &result)) {
		assert.Equal(s.T(), "test@example.com", result.NewAccountEmail)
	}
}

func (s *CreateAccountRequestDecodeTestSuite) TestDecodeFailsWithMisnamedField() {
	s.stream[2] = '_' // changes "email" key to "_mail"
	var result CreateAccountRequest

	assert.Nil(s.T(), json.Unmarshal(s.stream, &result))
	assert.Equal(s.T(), result.NewAccountEmail, "") // Field was not populated
}

type CreateAccountRequestEncodeTestSuite struct {
	suite.Suite
	value CreateAccountRequest
}

func TestValidCreateAccountRequestEncoding(t *testing.T) {
	suite.Run(t, new(CreateAccountRequestEncodeTestSuite))
}

func (s *CreateAccountRequestEncodeTestSuite) SetupTest() {
	s.value = CreateAccountRequest{NewAccountEmail: "test@example.com"}
}

func (s *CreateAccountRequestEncodeTestSuite) TestEncodeStructuredDataToStream() {
	stream, streamErr := json.Marshal(s.value)

	assert.Nil(s.T(), streamErr)
	assert.Equal(s.T(), []byte("{\"email\":\"test@example.com\"}"), stream)
}

type ErrorResponseDecodeTestSuite struct {
	suite.Suite
	stream []byte
}

func TestErrorResponseDecoding(t *testing.T) {
	suite.Run(t, new(ErrorResponseDecodeTestSuite))
}

func (s *ErrorResponseDecodeTestSuite) SetupTest() {
	s.stream = []byte("{\"reason\":\"OOPS\",\"err_msg\":\"because of this\"}")
}

func (s *ErrorResponseDecodeTestSuite) TestDecodeToStructuredData() {
	var result ErrorResponse
	if assert.Nil(s.T(), json.Unmarshal(s.stream, &result)) {
		assert.Equal(s.T(), "OOPS", result.Reason)
		assert.Equal(s.T(), "because of this", result.Message)
	}
}

func (s *ErrorResponseDecodeTestSuite) TestDecodeFailsWithMisnamedField() {
	s.stream[21] = ' ' // changes "err_msg" key to "err msg"
	var result ErrorResponse

	assert.Nil(s.T(), json.Unmarshal(s.stream, &result))
	assert.Equal(s.T(), result.Reason, "OOPS") // Field still populated
	assert.Equal(s.T(), result.Message, "")    // Field was not populated
}

type ErrorResponseEncodeTestSuite struct {
	suite.Suite
	value ErrorResponse
}

func TestErrorResponseEncoding(t *testing.T) {
	suite.Run(t, new(ErrorResponseEncodeTestSuite))
}

func (s *ErrorResponseEncodeTestSuite) SetupTest() {
	s.value = ErrorResponse{Reason: "OOPS", Message: "because of this"}
}

func (s *ErrorResponseEncodeTestSuite) TestEncodeStructuredDataToStream() {
	stream, streamErr := json.Marshal(s.value)

	assert.Nil(s.T(), streamErr)
	assert.Equal(s.T(), []byte("{\"reason\":\"OOPS\",\"err_msg\":\"because of this\"}"), stream)
}
