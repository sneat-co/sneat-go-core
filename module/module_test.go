package module

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewModule(t *testing.T) {
	m := NewModule("test")
	assert.NotNil(t, m)
}
