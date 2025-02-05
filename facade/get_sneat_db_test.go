package facade

import (
	"context"
	"errors"
	"testing"
)

func TestGetDatabase(t *testing.T) {
	ctx := context.Background()
	defer mustPanicWithErrNotInitialized(t)
	_, _ = GetSneatDB(ctx)
	//if db != nil {
	//	t.Error("GetSneatDB() = ", db)
	//}
	//if !errors.Is(err, ErrNotInitialized) {
	//	t.Error("Expected to return ErrNotInitialized")
	//}
}

func mustPanicWithErrNotInitialized(t *testing.T) {
	if r := recover(); r == nil {
		t.Error("Expected to panic")
	} else if err := r.(error); !errors.Is(err, ErrNotInitialized) {
		t.Error("Expected to return ErrNotInitialized")
	}
}
