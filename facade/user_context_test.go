package facade

import "testing"

func TestNewUserContext(t *testing.T) {
	userCtx := NewUserContext("123")
	if userCtx == nil {
		t.Fatal("userCtx == nil")
	}
	if userCtx.GetUserID() != "123" {
		t.Errorf("userCtx.GetUserIDFromContext() != \"123\": %v", userCtx.GetUserID())
	}
}
