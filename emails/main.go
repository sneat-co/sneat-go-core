package emails

import "context"

var _client Client

// Init initializes client
func Init(client Client) {
	if client == nil {
		panic("client == nil")
	}
	_client = client
}

// Send sends an email
func Send(ctx context.Context, email Email) (Sent, error) {
	return _client.Send(ctx, email)
}
