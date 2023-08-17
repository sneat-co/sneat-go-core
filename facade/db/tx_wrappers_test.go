package db

import (
	"context"
	"testing"
)

func TestTxUpdate(t *testing.T) {
	t.Run("should_panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		_ = TxUpdate(context.TODO(), nil, nil, nil)
	})
}
