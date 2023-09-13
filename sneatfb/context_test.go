package sneatfb

import (
	"context"
	"firebase.google.com/go/v4/auth"
	"testing"
)

func TestNewFirebaseContext(t *testing.T) {
	t.Skip("TODO: implement")
	ctx := context.WithValue(context.Background(), &firebaseTokenContextKey, &auth.Token{UID: "TEST_user"})
	_, err := NewContextWithFirestoreClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create FirestoreContext: %v", err)
	}
	//if ctx == nil {
	//	t.Fatalf("Context value is nil")
	//}
}
