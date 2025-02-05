package consts4dal

import (
	"context"
	"time"
)

const DefaultDeadLine = 5 * time.Second

var WithDefaultDeadLine = func(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithDeadline(ctx, time.Now().Add(DefaultDeadLine))
}
