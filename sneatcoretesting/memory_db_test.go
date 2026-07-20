package sneatcoretesting_test

import (
	"context"
	"testing"

	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/sneatcoretesting"
)

func TestSetupMemoryDB(t *testing.T) {
	db := sneatcoretesting.SetupMemoryDB(t)
	got, err := facade.GetSneatDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != db {
		t.Error("SetupMemoryDB did not install the created database")
	}
}

func TestNewMemoryDB_RejectsReadsAfterWrites(t *testing.T) {
	db := sneatcoretesting.NewMemoryDB()
	key := dal.NewKeyWithID("records", "strict-ordering")
	err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		if err := tx.Set(ctx, dal.NewRecordWithData(key, new(struct{}))); err != nil {
			return err
		}
		return tx.Get(ctx, dal.NewRecordWithData(key, new(struct{})))
	})
	if err == nil {
		t.Fatal("strict test DB accepted a read after a write")
	}
}
