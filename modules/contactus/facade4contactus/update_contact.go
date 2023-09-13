package facade4contactus

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
)

var updateContact = func(ctx context.Context, tx dal.ReadwriteTransaction, key *dal.Key, updates []dal.Update) error {
	if err := tx.Update(ctx, key, updates); err != nil {
		return fmt.Errorf("failed to update parent contact record: %w", err)
	}
	return nil
}
