package facade

import (
	"context"
	"errors"
	"testing"
)

func TestGetDatabase(t *testing.T) {
	db, err := GetDatabase(context.Background())
	if db != nil {
		t.Error("GetDatabase() = ", db)
	}
	if !errors.Is(err, ErrNotInitialized) {
		t.Error("Expected to return ErrNotInitialized")
	}
}
