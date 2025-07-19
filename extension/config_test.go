package extension

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewExtension(t *testing.T) {
	m := NewExtension("test",
		RegisterRoutes(nil),
		RegisterDelays(nil),
		RegisterNotificator(nil),
	)
	assert.NotNil(t, m)
}
