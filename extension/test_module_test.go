package extension

import (
	"context"
	"net/http"
	"testing"

	"github.com/sneat-co/sneat-go-core/coretypes"
)

type fakeConfig struct {
	Config
	id       coretypes.ExtID
	register func(args RegistrationArgs)
}

func (f fakeConfig) ID() coretypes.ExtID {
	return f.id
}

func (f fakeConfig) Register(args RegistrationArgs) {
	if f.register != nil {
		f.register(args)
	}
}

func TestAssertExtension(t *testing.T) {
	t.Run("nil_config", func(t *testing.T) {
		mockT := new(testing.T)
		done := make(chan bool)
		go func() {
			defer func() {
				_ = recover()
				done <- true
			}()
			AssertExtension(mockT, nil, Expected{})
		}()
		<-done
		if !mockT.Failed() {
			t.Fatal("expected failure on nil config")
		}
	})

	t.Run("wrong_id", func(t *testing.T) {
		mockT := new(testing.T)
		cfg := fakeConfig{id: "actual"}
		done := make(chan bool)
		go func() {
			defer func() {
				_ = recover()
				done <- true
			}()
			AssertExtension(mockT, cfg, Expected{ExtID: "expected"})
		}()
		<-done
		if !mockT.Failed() {
			t.Fatal("expected failure on wrong ID")
		}
	})

	t.Run("wrong_handlers_count", func(t *testing.T) {
		mockT := new(testing.T)
		cfg := fakeConfig{
			id: "ext1",
			register: func(args RegistrationArgs) {
				args.Handle()("GET", "/test", func(http.ResponseWriter, *http.Request) {})
			},
		}
		AssertExtension(mockT, cfg, Expected{ExtID: "ext1", HandlersCount: 2})
		if !mockT.Failed() {
			t.Fatal("expected failure on wrong handlers count")
		}
	})

	t.Run("wrong_delayers_count", func(t *testing.T) {
		dummyDelayWorker()
		mockT := new(testing.T)
		cfg := fakeConfig{
			id: "ext1",
			register: func(args RegistrationArgs) {
				_ = args.MustRegisterDelayFunc()("key1", nil)
			},
		}
		AssertExtension(mockT, cfg, Expected{ExtID: "ext1", DelayersCount: 2})
		if !mockT.Failed() {
			t.Fatal("expected failure on wrong delayers count")
		}
	})

	t.Run("success_with_delayers_and_handlers", func(t *testing.T) {
		cfg := fakeConfig{
			id: "ext1",
			register: func(args RegistrationArgs) {
				args.Handle()("GET", "/test", func(http.ResponseWriter, *http.Request) {})
				delayer := args.MustRegisterDelayFunc()("key1", nil)
				_ = delayer.EnqueueWork(context.Background(), nil)
				_ = delayer.EnqueueWorkMulti(context.Background(), nil)
			},
		}
		AssertExtension(t, cfg, Expected{
			ExtID:         "ext1",
			HandlersCount: 1,
			DelayersCount: 1,
		})
	})
}
