package facade4memberus

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/facade/db"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"time"
)

var txUpdate = func(ctx context.Context, tx dal.ReadwriteTransaction, key *dal.Key, data []dal.Update, opts ...dal.Precondition) error {
	return db.TxUpdate(ctx, tx, key, data, opts...)
}

// updateMembersGroup wrapper to update TeamIDs
func updateMembersGroup(ctx context.Context, tx dal.ReadwriteTransaction, timestamp time.Time, membersGroup dbmodels.Versioned, key *dal.Key, data []dal.Update, opts ...dal.Precondition) error {
	if err := membersGroup.Validate(); err != nil {
		return fmt.Errorf("membersGroup record is not valid: %w", err)
	}
	data = append(data,
		dal.Update{Field: "v", Value: membersGroup.IncreaseVersion(timestamp)},
		dal.Update{Field: "timestamp", Value: timestamp},
	)
	return txUpdate(ctx, tx, key, data, opts...)
}
