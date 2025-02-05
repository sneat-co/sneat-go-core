package facade

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
)

// GetSneatDB creates a new DB for a given context
var GetSneatDB = func(ctx context.Context) (db dal.DB, err error) {
	err = fmt.Errorf("%w: facade.GetSneatDB(context.Context) (dal.DB, error)", ErrNotInitialized)
	panic(err)
}
