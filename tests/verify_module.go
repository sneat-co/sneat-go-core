package tests

import (
	"github.com/sneat-co/sneat-go-core/modules"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func VerifyModule(t *testing.T, m modules.Module) {
	assert.NotNil(t, m)
	assert.NotEmpty(t, m.ID())
	assert.NotNil(t, m.Register)
	var handleCalled int
	handle := func(method, path string, handler http.HandlerFunc) {
		handleCalled++
	}
	regArgs := modules.NewModuleRegistrationArgs(handle)
	m.Register(regArgs)
	assert.Greater(t, handleCalled, 0)
}
