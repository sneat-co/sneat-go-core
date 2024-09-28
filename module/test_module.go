package module

import (
	"context"
	"github.com/strongo/delaying"
	"net/http"
	"testing"
)

type Expected struct {
	ModuleID      string
	HandlersCount int
	DelayersCount int
}

func AssertModule(t *testing.T, m Module, expected Expected) {
	if m == nil {
		t.Fatalf("Module() must not return nil")
	}
	if m.ID() != expected.ModuleID {
		t.Fatalf("Module().ID() must return %q but got %q", expected.ModuleID, m.ID())
	}
	var handle HTTPHandleFunc
	handlersCount := 0
	handle = func(method, path string, handler http.HandlerFunc) {
		handlersCount++
	}
	delayersCount := 0
	mustRegisterDelayFunc := func(key string, i any) delaying.Function {
		delayersCount++
		enqueueWork := func(c context.Context, params delaying.Params, args ...interface{}) error {
			return nil
		}
		enqueueWorkMulti := func(c context.Context, params delaying.Params, args ...[]interface{}) error {
			return nil
		}
		return delaying.NewFunction(key, func() {}, enqueueWork, enqueueWorkMulti)
	}
	delaying.Init(mustRegisterDelayFunc)
	args := NewModuleRegistrationArgs(handle, mustRegisterDelayFunc)
	m.Register(args)
	if handlersCount != expected.HandlersCount {
		t.Errorf("Module().Register() must register %d handlers but got %d", expected.HandlersCount, handlersCount)
	}
	if delayersCount != expected.DelayersCount {
		t.Errorf("Module().Register() must register %d delayers but got %d", expected.DelayersCount, delayersCount)
	}
}
