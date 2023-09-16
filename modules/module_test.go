package modules

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewModule(t *testing.T) {
	m := NewModule("test", func(handle HTTPHandleFunc) {
	})
	assert.NotNil(t, m)
}
