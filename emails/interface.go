package emails

// Client interface
type Client interface {
	Send(email Email) (Sent, error)
}

// Sent interface
type Sent interface {
	MessageID() string
}
