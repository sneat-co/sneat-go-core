package emails

import (
	"context"
	"testing"
)

func TestInit(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("panic expected")
		}
	}()
	Init(nil)
}

func TestSend(t *testing.T) {
	t.Run("should_panic", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("panic expected")
			}
		}()
		_, _ = Send(context.Background(), Email{})
	})
}
