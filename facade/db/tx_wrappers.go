package db

import (
	"context"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/update"
)

// TxUpdate is a wrapper to call tx.Update()
// TODO: Do we need this wrapper or can use Dalgo mocks?
var TxUpdate = func(ctx context.Context, tx dal.ReadwriteTransaction, key *dal.Key, updates []update.Update, opts ...dal.Precondition) error {
	return tx.Update(ctx, key, updates, opts...)
}
