package db

import (
	"context"
	"github.com/dal-go/dalgo/dal"
)

// TxUpdate is a wrapper to call tx.Update()
// TODO: Do we need this wrapper or can use Dalgo mocks?
var TxUpdate = func(ctx context.Context, tx dal.ReadwriteTransaction, key *dal.Key, updates []dal.Update, opts ...dal.Precondition) error {
	return tx.Update(ctx, key, updates, opts...)
}
