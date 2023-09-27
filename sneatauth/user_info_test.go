package sneatauth

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetUserInfo(t *testing.T) {
	assert.Panics(t, func() {
		_, _ = GetUserInfo(context.Background(), "u1")
	})
}
