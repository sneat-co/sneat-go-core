package facade

import (
	"context"
	"testing"
)

func TestContextWithUserID(t *testing.T) {
	const userID = "123"
	ctx := NewContextWithUserID(context.Background(), userID)
	if got := GetUserIDFromContext(ctx); got != userID {
		t.Errorf("GetUserIDFromContext() = %v, want %v", got, userID)
	}
}
