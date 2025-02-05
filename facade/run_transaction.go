package facade

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/consts4dal"
)

// RunReadwriteTransaction is a helper wrapper that created a facade DB instance and runs a transaction
func RunReadwriteTransaction(ctx context.Context, f func(ctx context.Context, tx dal.ReadwriteTransaction) error, options ...dal.TransactionOption) (err error) {
	if _, ok := ctx.Deadline(); !ok {
		ctx, _ = consts4dal.WithDefaultDeadLine(ctx)
	}
	var db dal.DB
	if db, err = GetSneatDB(ctx); err != nil {
		return fmt.Errorf("failed to GetSneatDB() inside facade.RunReadwriteTransaction(): %w", err)
	}
	return db.RunReadwriteTransaction(ctx, f, options...)
}
