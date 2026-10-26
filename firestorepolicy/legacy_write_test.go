package firestorepolicy

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dal-go/dalgo/access"
	"github.com/dal-go/dalgo/adapters/dalgo2memory"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/dal-go/record/update"
	"reflect"
	"testing"
)

func TestLegacyWritableRootsRequiredAndObservationPreserved(t *testing.T) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(rootRegistryJSON, &object); err != nil {
		t.Fatal(err)
	}
	for _, missing := range []bool{true, false} {
		if missing {
			delete(object, "legacyWritableRoots")
		} else {
			object["legacyWritableRoots"] = json.RawMessage("null")
		}
		data, _ := json.Marshal(object)
		if _, err := LoadRootRegistry(data); err == nil {
			t.Fatal("missing/null writable inventory accepted")
		}
	}
	r, _, err := DefaultRootRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if got := r.CompareRoots(r.LegacyRoots); len(got.Legacy) != 48 || len(got.Unexpected) != 0 {
		t.Fatal(got)
	}
	r.LegacyWritableRoots = []string{}
	data, _ := json.Marshal(r)
	if _, err = LoadRootRegistry(data); err != nil {
		t.Fatalf("explicit empty write set rejected: %v", err)
	}
}

func (d *directPolicyDB) Insert(ctx context.Context, row record.Record, options ...dal.InsertOption) error {
	d.calls++
	return d.DB.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error { return tx.Insert(ctx, row, options...) })
}
func (d *directPolicyDB) Set(ctx context.Context, row record.Record) error {
	d.calls++
	return d.DB.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error { return tx.Set(ctx, row) })
}
func (d *directPolicyDB) Update(ctx context.Context, key *record.Key, changes []update.Update, options ...dal.Precondition) error {
	d.calls++
	return d.DB.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Update(ctx, key, changes, options...)
	})
}
func (d *directPolicyDB) Delete(ctx context.Context, key *record.Key) error {
	d.calls++
	return d.DB.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error { return tx.Delete(ctx, key) })
}

func TestEveryLegacyRootWriteEligibility(t *testing.T) {
	registry, _, err := DefaultRootRegistry()
	if err != nil {
		t.Fatal(err)
	}
	pending := []string{"botPlatforms", "bttt-matches", "moneyEvents", "oauthDeviceAuthorizations", "oauthDeviceUserCodes", "referrers"}
	if !reflect.DeepEqual(registry.LegacyWritableRoots, pending) {
		t.Fatal(registry.LegacyWritableRoots)
	}
	policy, err := registry.RootPolicy()
	if err != nil {
		t.Fatal(err)
	}
	for _, root := range registry.LegacyRoots {
		t.Run(root, func(t *testing.T) {
			allowed := containsRoot(pending, root)
			key := record.NewKeyWithID(root, "retained")
			for _, op := range []access.Operations{access.Insert, access.Set, access.Update, access.Delete} {
				for _, target := range []*record.Key{key, record.NewKeyWithParentAndID(key, "descendants", "child"), record.NewIncompleteKey(root, reflect.String, nil)} {
					if got := policy.Decide(context.Background(), access.Request{Operation: op, Resources: []access.Resource{access.RecordResourceForKey(target)}}).Allowed; got != allowed {
						t.Fatalf("%s allowed=%v want=%v", op, got, allowed)
					}
				}
			}
			raw := dalgo2memory.New(dalgo2memory.FirestoreProfile())
			ctx := access.WithPolicy(context.Background(), access.MustPolicy("permissive-caller", access.Root(access.Allow(access.ReadWrite))))
			row := func() record.Record {
				return record.NewRecordWithData(key, &struct {
					N int `json:"n"`
				}{N: 1})
			}
			if err := raw.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error { return tx.Set(ctx, row()) }); err != nil {
				t.Fatal(err)
			}
			direct := &directPolicyDB{DB: raw}
			db, err := ProtectRootCollections(direct)
			if err != nil {
				t.Fatal(err)
			}
			if err := db.Get(ctx, row()); err != nil {
				t.Fatal(err)
			}
			if allowed {
				if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error { return tx.Set(ctx, row()) }); err != nil {
					t.Fatal(err)
				}
				return
			}
			good := record.NewRecordWithData(record.NewKeyWithID("users", "batch-good"), map[string]any{"n": 1})
			actions := []struct {
				name string
				run  func(dal.WriteSession) error
			}{
				{"insert", func(tx dal.WriteSession) error { return tx.Insert(ctx, row()) }},
				{"set", func(tx dal.WriteSession) error { return tx.Set(ctx, row()) }},
				{"update", func(tx dal.WriteSession) error {
					return tx.Update(ctx, key, []update.Update{update.ByFieldName("n", 2)})
				}},
				{"delete", func(tx dal.WriteSession) error { return tx.Delete(ctx, key) }},
				{"batch-set", func(tx dal.WriteSession) error { return tx.SetMulti(ctx, []record.Record{good, row()}) }},
				{"batch-insert", func(tx dal.WriteSession) error { return tx.InsertMulti(ctx, []record.Record{good, row()}) }},
				{"generated", func(tx dal.WriteSession) error {
					return tx.Insert(ctx, record.NewRecordWithData(record.NewIncompleteKey(root, reflect.String, nil), map[string]any{"n": 1}), dal.WithAdapterGeneratedID())
				}},
			}
			writer, ok := db.(dal.WriteSession)
			if !ok {
				t.Fatal("direct write capability missing")
			}
			for _, action := range actions {
				t.Run(action.name, func(t *testing.T) {
					err := action.run(writer)
					if !errors.Is(err, access.ErrAccessDenied) || direct.calls != 0 {
						t.Fatalf("direct denial=%v adapterCalls=%d", err, direct.calls)
					}
					err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error { return action.run(tx) })
					if !errors.Is(err, access.ErrAccessDenied) {
						t.Fatalf("transaction denial=%v", err)
					}
					if exists, _ := raw.Exists(ctx, good.Key()); exists {
						t.Fatal("partial mixed batch escaped")
					}
					if exists, _ := raw.Exists(ctx, key); !exists {
						t.Fatal("retained record removed")
					}
				})
			}
		})
	}
}
