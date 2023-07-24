package capturer

import "context"

// NoOpErrorLogger implements logger interface but does nothing
type NoOpErrorLogger struct {
}

// LogError imitates logging error but actually does nothing
func (NoOpErrorLogger) LogError(ctx context.Context, err error) {
}
