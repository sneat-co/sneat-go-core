package facade

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
)

// GetDatabase creates a new DB for a given context
var GetDatabase = func(ctx context.Context) (dal.DB, error) {
	return nil, fmt.Errorf("%w: facade.GetDatabase(context.Context) (dal.DB, error)", ErrNotInitialized)
}

// RunReadwriteTransaction is a helper wrapper that created a facade DB instance and runs a transaction
func RunReadwriteTransaction(ctx context.Context, f func(ctx context.Context, tx dal.ReadwriteTransaction) error) (err error) {
	var db dal.DB
	if db, err = GetDatabase(ctx); err != nil {
		return err
	}
	return db.RunReadwriteTransaction(ctx, f)
}
