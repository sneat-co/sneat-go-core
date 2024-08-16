package facade

import "strings"

// UserContext defines an interface for a AuthUserContext context that provides AuthUserContext ID.
type UserContext interface {
	GetUserID() string
}

// NewUserContext creates new AuthUserContext context
func NewUserContext(id string) (userCtx UserContext) {
	if strings.TrimSpace(id) == "" {
		panic("AuthUserContext id is empty string")
	}
	userCtx = AuthUserContext{ID: id}
	return
}
