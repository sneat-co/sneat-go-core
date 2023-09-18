package verify

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultJsonWithAuthRequired(t *testing.T) {
	assert.True(t, DefaultJsonWithAuthRequired.AuthenticationRequired())
	assert.Equal(t, MinJSONRequestSize, DefaultJsonWithAuthRequired.MinimumContentLength())
	assert.Equal(t, DefaultMaxJSONRequestSize, DefaultJsonWithAuthRequired.MaximumContentLength())
}

func TestDefaultJsonWithNoAuthRequired(t *testing.T) {
	assert.False(t, DefaultJsonWithNoAuthRequired.AuthenticationRequired())
	assert.Equal(t, MinJSONRequestSize, DefaultJsonWithNoAuthRequired.MinimumContentLength())
	assert.Equal(t, DefaultMaxJSONRequestSize, DefaultJsonWithNoAuthRequired.MaximumContentLength())
}

func TestNoContentAuthRequired(t *testing.T) {
	assert.True(t, NoContentAuthRequired.AuthenticationRequired())
	assert.Equal(t, int64(0), NoContentAuthRequired.MinimumContentLength())
	assert.Equal(t, int64(0), NoContentAuthRequired.MaximumContentLength())
}
