package consts4dal

import (
	"context"
	"testing"
)

func TestWithDefaultDeadLine(t *testing.T) {
	ctx, cancel := WithDefaultDeadLine(context.Background())
	defer cancel()
	deadline, ok := ctx.Deadline()
	if !ok || deadline.IsZero() {
		t.Fatal("expected deadline to be set")
	}
}
