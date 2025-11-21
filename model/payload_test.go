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

func TestValidCreateAccountRequestDecoding(t *testing.T) {
	suite.Run(t, new(CreateAccountRequestDecodeTestSuite))
}

func (s *CreateAccountRequestDecodeTestSuite) SetupTest() {
	s.stream = []byte("{\"email\": \"test@example.com\"}")
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
