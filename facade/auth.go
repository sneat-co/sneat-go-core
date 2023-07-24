package facade

import (
	"context"
	"strings"
)

// AuthContext defines auth context interface
type AuthContext interface {
	User(ctx context.Context, authRequired bool) (User, error)
}

type user struct {
	id string
}

// ID returns user's ID
func (v user) ID() string {
	return v.id
}

// User defines an interface for a user context that provides user ID.
type User interface {
	ID() string
}

// NewUser creates new user context
func NewUser(id string) User {
	if strings.TrimSpace(id) == "" {
		panic("user id is empty string")
	}
	return user{id: id}
}
