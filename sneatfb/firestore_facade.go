package sneatfb

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"os"
)

// FirestoreContext provides context for Firestore
type FirestoreContext struct {
	context.Context
	Firestore *firestore.Client
	UID       string
}

var _ context.Context = (*FirestoreContext)(nil)

var firestoreClientContextKey = "firestoreClientContextKey"

// NewContextWithFirestoreClient creates new Firestore context
func NewContextWithFirestoreClient(ctx context.Context) (*FirestoreContext, error) {
	//var fbConfig firebase.Config
	//fbConfig = firebase.Config{
	//
	//}
	//var opts []option.ClientOption
	//if os.Getenv("FIREBASE_AUTH_EMULATOR_HOST") != "" {
	//	opts = append(opts, option.WithoutAuthentication())
	//}

	//if uid == "" {
	//	firestoreProjectId = "sneat-team"
	//}

	googleCloudProjectID := os.Getenv("GCLOUD_PROJECT")
	if googleCloudProjectID == "" {
		googleCloudProjectID = "sneat-team"
	}
	client, err := firestore.NewClient(ctx, googleCloudProjectID)
	if err != nil {
		err = fmt.Errorf("failed to create firestore client: %w", err)
	}
	token := FirebaseTokenFromContext(ctx)
	return &FirestoreContext{
		Context:   context.WithValue(ctx, &firestoreClientContextKey, client),
		Firestore: client,
		UID:       token.UID,
	}, err
}
