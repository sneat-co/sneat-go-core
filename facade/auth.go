package facade

import (
	"context"
	"strings"
)

// AuthContext defines auth context interface
type AuthContext interface {
	User(ctx context.Context, authRequired bool) (User, error)
}

type AuthUser struct {
	ID string
}

// GetID returns AuthUser's ID
func (v AuthUser) GetID() string {
	return v.ID
}

// User defines an interface for a AuthUser context that provides AuthUser ID.
type User interface {
	GetID() string
}

// NewUser creates new AuthUser context
func NewUser(id string) User {
	if strings.TrimSpace(id) == "" {
		panic("AuthUser id is empty string")
	}
	return AuthUser{ID: id}
}
