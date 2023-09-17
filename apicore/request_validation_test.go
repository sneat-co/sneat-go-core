package apicore

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAuthTokenFromHttpRequest(t *testing.T) {
	assert.Nil(t, GetAuthTokenFromHttpRequest)
}
