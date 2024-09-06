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

// RunReadwriteTransaction is a helper wrapper that created a facade DB instance and runs a transaction
func RunReadwriteTransaction(ctx context.Context, f func(ctx context.Context, tx dal.ReadwriteTransaction) error, options ...dal.TransactionOption) (err error) {
	var db dal.DB
	if db, err = GetSneatDB(ctx); err != nil {
		return fmt.Errorf("failed to GetSneatDB() inside facade.RunReadwriteTransaction(): %w", err)
	}
	return db.RunReadwriteTransaction(ctx, f, options...)
}
