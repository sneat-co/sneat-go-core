package tests

import (
	"github.com/sneat-co/sneat-go-core/module"
	"github.com/stretchr/testify/assert"
	"github.com/strongo/delaying"
	"net/http"
	"testing"
)

func VerifyModule(t *testing.T, m module.Module) {
	VerifyModuleWithDelays(t, m, false)
}

// VerifyModuleWithDelays is used in module's package tests to verify that module is correctly implements registration
func VerifyModuleWithDelays(t *testing.T, m module.Module, expectsDelays bool) {
	assert.NotNil(t, m)
	assert.NotEmpty(t, m.ID())
	assert.NotNil(t, m.Register)

	var handleCalled int
	var delaysCalled int

	handle := func(method, path string, handler http.HandlerFunc) {
		handleCalled++
	}
	mustRegisterDelayFunc := func(key string, i any) delaying.Function {
		delaysCalled++
		return nil
	}

	regArgs := module.NewModuleRegistrationArgs(
		handle,
		mustRegisterDelayFunc,
	)
	m.Register(regArgs)
	assert.Greater(t, handleCalled, 0)
	if expectsDelays {
		if delaysCalled == 0 {
			t.Errorf("Module %s expects delays but mustRegisterDelayFunc has not been called", m.ID())
		}
	} else if delaysCalled > 0 {
		t.Errorf("Module %s does not expect delays but mustRegisterDelayFunc has been called", m.ID())
	}
}
