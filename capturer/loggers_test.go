package capturer

import (
	"context"
	"errors"
	"testing"
)

func TestNoOpErrorLogger_LogError(t *testing.T) {
	l := NoOpErrorLogger{}

	t.Run("nil_values", func(t *testing.T) {
		l.LogError(nil, nil)
	})
	t.Run("with_values", func(t *testing.T) {
		l.LogError(context.Background(), errors.New("test error"))
	})
}
