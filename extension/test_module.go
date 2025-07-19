package extension

import (
	"context"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/strongo/delaying"
	"net/http"
	"testing"
)

type Expected struct {
	ExtID         coretypes.ExtID
	HandlersCount int
	DelayersCount int
}

func AssertExtension(t *testing.T, m Config, expected Expected) {
	if m == nil {
		t.Fatalf("Config() must not return nil")
	}
	if m.ID() != expected.ExtID {
		t.Fatalf("Config().userID() must return %q but got %q", expected.ExtID, m.ID())
	}
	var handle HTTPHandleFunc
	handlersCount := 0
	handle = func(method, path string, handler http.HandlerFunc) {
		handlersCount++
	}
	delayersCount := 0
	mustRegisterDelayFunc := func(key string, i any) delaying.Delayer {
		delayersCount++
		enqueueWork := func(c context.Context, params delaying.Params, args ...interface{}) error {
			return nil
		}
		enqueueWorkMulti := func(c context.Context, params delaying.Params, args ...[]interface{}) error {
			return nil
		}
		return delaying.NewDelayer(key, func() {}, enqueueWork, enqueueWorkMulti)
	}
	delaying.Init(mustRegisterDelayFunc)
	args := NewModuleRegistrationArgs(handle, mustRegisterDelayFunc)
	m.Register(args)
	if handlersCount != expected.HandlersCount {
		t.Errorf("Config().Register() must register %d handlers but got %d", expected.HandlersCount, handlersCount)
	}
	if delayersCount != expected.DelayersCount {
		t.Errorf("Config().Register() must register %d delayers but got %d", expected.DelayersCount, delayersCount)
	}
}
