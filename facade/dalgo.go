package facade

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/dal-go/dalgo/dal"
	dalgoFirestore "github.com/dal-go/dalgo2firestore"
	"os"
	"strings"
)

func getProjectID() string {
	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "GAE_APPLICATION=") {
			if strings.HasSuffix(v, "sneat-team") {
				return "sneat-team"
			}
			if strings.HasSuffix(v, "sneat-eu") {
				return "sneat-eu"
			}
			if strings.HasSuffix(v, "sneatapp") {
				return "sneatapp"
			}
		}
	}
	return "demo-local-sneat-app"
}

var projectID = getProjectID()

// GetDatabase creates a new DB for a given context
var GetDatabase = func(ctx context.Context) dal.Database {

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		panic(err)
	}
	return dalgoFirestore.NewDatabase("sneat", client)
}
