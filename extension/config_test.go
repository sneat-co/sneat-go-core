package extension

import (
	"context"
	"net/http"
	"testing"

	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/stretchr/testify/assert"
	"github.com/strongo/delaying"
)

// noopHandle is a HTTPHandleFunc that does nothing.
var noopHandle HTTPHandleFunc = func(method, path string, handler http.HandlerFunc) {}

// noopMustRegisterDelayFunc is a mustRegisterDelayFunc that returns a no-op Delayer.
var noopMustRegisterDelayFunc = func(key string, i any) delaying.Delayer {
	return delaying.NewDelayer(key, func() {},
		func(c context.Context, params delaying.Params, args ...interface{}) error { return nil },
		func(c context.Context, params delaying.Params, args ...[]interface{}) error { return nil },
	)
}

func TestNewExtension_noOptions(t *testing.T) {
	m := NewExtension("test")
	assert.NotNil(t, m)
	assert.Equal(t, coretypes.ExtID("test"), m.ID())
	// Register with nil args should not panic when no registrars are set.
	assert.NotPanics(t, func() {
		m.Register(NewModuleRegistrationArgs(nil, nil))
	})
	// Exercise internal() for coverage (it is a package-private marker method).
	m.(*config).internal()
}

func TestNewExtension_nilFuncs(t *testing.T) {
	// Passing nil functions is allowed by the API; they are stored and called.
	m := NewExtension("test",
		RegisterRoutes(nil),
		RegisterDelays(nil),
		RegisterNotificator(nil),
	)
	assert.NotNil(t, m)
}

func TestRegisterRoutes_composes(t *testing.T) {
	var calls []int
	r1 := func(handle HTTPHandleFunc) { calls = append(calls, 1) }
	r2 := func(handle HTTPHandleFunc) { calls = append(calls, 2) }

	m := NewExtension("test", RegisterRoutes(r1), RegisterRoutes(r2))
	m.Register(NewModuleRegistrationArgs(noopHandle, nil))

	assert.Equal(t, []int{1, 2}, calls, "both RegisterRoutes options must be called in order")
}

func TestRegisterDelays_composes(t *testing.T) {
	var calls []int
	d1 := func(mustRegisterFunc func(key string, i any) delaying.Delayer) { calls = append(calls, 1) }
	d2 := func(mustRegisterFunc func(key string, i any) delaying.Delayer) { calls = append(calls, 2) }

	m := NewExtension("test", RegisterDelays(d1), RegisterDelays(d2))
	m.Register(NewModuleRegistrationArgs(nil, noopMustRegisterDelayFunc))

	assert.Equal(t, []int{1, 2}, calls, "both RegisterDelays options must be called in order")
}

func TestRegisterNotificator_composes(t *testing.T) {
	var calls []int
	n1 := func(f CreateNotificationFunc) { calls = append(calls, 1) }
	n2 := func(f CreateNotificationFunc) { calls = append(calls, 2) }

	m := NewExtension("test", RegisterNotificator(n1), RegisterNotificator(n2))
	m.Register(NewModuleRegistrationArgs(nil, nil))

	assert.Equal(t, []int{1, 2}, calls, "both RegisterNotificator options must be called in order")
}

func TestRegisterRoutes_panicOnNilHandle(t *testing.T) {
	m := NewExtension("test", RegisterRoutes(func(handle HTTPHandleFunc) {}))
	assert.Panics(t, func() {
		m.Register(NewModuleRegistrationArgs(nil, nil))
	})
}

func TestRegisterDelays_panicOnNilMustRegisterDelayFunc(t *testing.T) {
	m := NewExtension("test", RegisterDelays(func(mustRegisterFunc func(key string, i any) delaying.Delayer) {}))
	assert.Panics(t, func() {
		m.Register(NewModuleRegistrationArgs(nil, nil))
	})
}
