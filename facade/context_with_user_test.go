package facade

import (
	"context"
	"testing"
)

func TestNewContextWithUser(t *testing.T) {
	ctx := NewContextWithUser(context.Background(), "123")
	userCtx := ctx.User()
	if userCtx.GetUserID() != "123" {
		t.Errorf("userCtx.GetUserIDFromContext() != \"123\": %v", userCtx.GetUserID())
	}
}

func TestGetUserContext(t *testing.T) {
	var ctx context.Context = NewContextWithUser(context.Background(), "123")
	var key = "abc"
	ctx = context.WithValue(ctx, &key, "def")
	userCtx := GetUserContext(ctx)
	if userCtx == nil {
		t.Fatal("userCtx == nil")
	}
	if userCtx.GetUserID() != "123" {
		t.Errorf("userCtx.GetUserIDFromContext() != \"123\": %v", userCtx.GetUserID())
	}
}

func Test_contextWithUser_User(t *testing.T) {
	userCtx := NewUserContext("123")
	ctx := contextWithUser{user: userCtx}
	if ctx.User() != userCtx {
		t.Error("ctx.User() != userCtx")
	}
}
