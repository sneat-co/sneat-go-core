package emails

import "context"

// Client interface
type Client interface {
	Send(ctx context.Context, email Email) (Sent, error)
}

// Sent interface
type Sent interface {
	MessageID() string
}
