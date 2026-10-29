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
	if got := r.CompareRoots(r.LegacyRoots); len(got.Legacy) != 6 || len(got.Unexpected) != 0 {
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
	return d.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error { return tx.Insert(ctx, row, options...) })
}
func (d *directPolicyDB) Set(ctx context.Context, row record.Record) error {
	d.calls++
	return d.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error { return tx.Set(ctx, row) })
}
func (d *directPolicyDB) Update(ctx context.Context, key *record.Key, changes []update.Update, options ...dal.Precondition) error {
	d.calls++
	return d.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Update(ctx, key, changes, options...)
	})
}
func (d *directPolicyDB) Delete(ctx context.Context, key *record.Key) error {
	d.calls++
	return d.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error { return tx.Delete(ctx, key) })
}

func TestEveryLegacyRootWriteEligibility(t *testing.T) {
	registry, _, err := DefaultRootRegistry()
	if err != nil {
		t.Fatal(err)
	}
	pending := []string{}
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

// The approved game identity permits its new namespace while retaining old
// match records for observation with no write authority.
func TestBiddingTicTacToeNamespaceCutover(t *testing.T) {
	raw := dalgo2memory.New(dalgo2memory.FirestoreProfile())
	db, err := ProtectRootCollections(raw)
	if err != nil {
		t.Fatal(err)
	}
	ctx := access.WithPolicy(context.Background(), access.MustPolicy("game-caller", access.Root(access.Allow(access.ReadWrite))))
	owner := record.NewKeyWithID("ext", "bidding-tictactoe")
	key := record.NewKeyWithParentAndID(owner, "matches", "fresh")
	if got := key.String(); got != "ext/bidding-tictactoe/matches/fresh" {
		t.Fatal(got)
	}
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Insert(ctx, record.NewRecordWithData(key, map[string]any{"move": 1}))
	}); err != nil {
		t.Fatal(err)
	}
	if exists, err := raw.Exists(ctx, key); err != nil || !exists {
		t.Fatalf("new namespace unavailable: %v %v", exists, err)
	}
	generated := record.NewRecordWithData(record.NewIncompleteKey("matches", reflect.String, owner), map[string]any{"move": 1})
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Insert(ctx, generated, dal.WithAdapterGeneratedID())
	}); err != nil {
		t.Fatal(err)
	}
	if generated.Key().ID == nil || generated.Key().ID == "" {
		t.Fatal("generated match ID missing")
	}
	old := record.NewRecordWithData(record.NewKeyWithID("bttt-matches", "old"), map[string]any{"move": 1})
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error { return tx.Set(ctx, old) }); !errors.Is(err, access.ErrAccessDenied) {
		t.Fatalf("old root recreated: %v", err)
	}
	registry, _, err := DefaultRootRegistry()
	if err != nil {
		t.Fatal(err)
	}
	got := registry.CompareRoots([]string{"bttt-matches"})
	if len(got.Legacy) != 0 || !reflect.DeepEqual(got.Unexpected, []string{"bttt-matches"}) {
		t.Fatal(got)
	}
}

func TestApprovedCoordinationNamespaceCutovers(t *testing.T) {
	raw := dalgo2memory.New(dalgo2memory.FirestoreProfile())
	db, err := ProtectRootCollections(raw)
	if err != nil {
		t.Fatal(err)
	}
	ctx := access.WithPolicy(context.Background(), access.MustPolicy("coordination-caller", access.Root(access.Allow(access.ReadWrite))))
	for _, tc := range []struct{ legacy, owner, leaf string }{
		{"botPlatforms", "bots", "platforms"},
		{"moneyEvents", "paymentus", "moneyEvents"},
		{"oauthDeviceAuthorizations", "auth", "deviceAuthorizations"},
		{"oauthDeviceUserCodes", "auth", "deviceUserCodes"},
	} {
		t.Run(tc.legacy, func(t *testing.T) {
			key := record.NewKeyWithParentAndID(record.NewKeyWithID("ext", tc.owner), tc.leaf, "fresh")
			if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
				return tx.Insert(ctx, record.NewRecordWithData(key, map[string]any{"v": 1}))
			}); err != nil {
				t.Fatal(err)
			}
			if exists, err := raw.Exists(ctx, key); err != nil || !exists {
				t.Fatalf("approved namespace unavailable: %v %v", exists, err)
			}
			old := record.NewRecordWithData(record.NewKeyWithID(tc.legacy, "old"), map[string]any{"v": 1})
			if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error { return tx.Set(ctx, old) }); !errors.Is(err, access.ErrAccessDenied) {
				t.Fatalf("legacy root writable: %v", err)
			}
		})
	}
}

// Root approval changes placement only. Caller authorization remains mandatory.
func TestReferrersApprovedRootPreservesCallerPolicy(t *testing.T) {
	registry, _, err := DefaultRootRegistry()
	if err != nil {
		t.Fatal(err)
	}
	comparison := registry.CompareRoots([]string{"referrers"})
	if len(comparison.Legacy) != 0 || len(comparison.Unexpected) != 0 {
		t.Fatal(comparison)
	}
	raw := dalgo2memory.New(dalgo2memory.FirestoreProfile())
	db, err := ProtectRootCollections(&directPolicyDB{DB: raw})
	if err != nil {
		t.Fatal(err)
	}
	writer := db.(dal.WriteSession)
	ctx := access.WithPolicy(context.Background(), access.MustPolicy("referrer-server", access.Root(access.Allow(access.ReadWrite))))
	row := func(id string) record.Record {
		return record.NewRecordWithData(record.NewKeyWithID("referrers", id), map[string]any{"n": 1})
	}
	if err := writer.Insert(ctx, row("direct")); err != nil {
		t.Fatal(err)
	}
	if err := writer.SetMulti(ctx, []record.Record{row("batch-a"), row("batch-b")}); err != nil {
		t.Fatal(err)
	}
	generated := record.NewRecordWithData(record.NewIncompleteKey("referrers", reflect.String, nil), map[string]any{"n": 1})
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Insert(ctx, generated, dal.WithAdapterGeneratedID())
	}); err != nil {
		t.Fatal(err)
	}
	if generated.Key().ID == nil || generated.Key().ID == "" {
		t.Fatal("missing generated referrer ID")
	}
	denied := access.WithPolicy(context.Background(), access.MustPolicy("client-denied", access.Root(access.Deny(access.ReadWrite))))
	if err := writer.Set(denied, row("denied")); !errors.Is(err, access.ErrAccessDenied) {
		t.Fatalf("placement expanded caller permission: %v", err)
	}
	if exists, _ := raw.Exists(ctx, row("denied").Key()); exists {
		t.Fatal("denied write reached adapter")
	}
}
