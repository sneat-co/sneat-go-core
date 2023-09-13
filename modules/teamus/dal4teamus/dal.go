package dal4teamus

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/facade/db"
	"time"
)

var txUpdate = func(ctx context.Context, tx dal.ReadwriteTransaction, key *dal.Key, data []dal.Update, opts ...dal.Precondition) error {
	return db.TxUpdate(ctx, tx, key, data, opts...)
}

func txUpdateTeam(ctx context.Context, tx dal.ReadwriteTransaction, timestamp time.Time, team TeamContext, data []dal.Update, opts ...dal.Precondition) error {
	if err := team.Data.Validate(); err != nil {
		return fmt.Errorf("team record is not valid: %w", err)
	}
	team.Data.Version++
	data = append(data,
		dal.Update{Field: "v", Value: team.Data.Version},
		dal.Update{Field: "timestamp", Value: timestamp},
	)
	return txUpdate(ctx, tx, team.Key, data, opts...)
}
