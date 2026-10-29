// Copyright 2026 Sneat Co.
package firestorepolicy

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/dal-go/dalgo/access"
	"github.com/dal-go/dalgo/adapters/dalgo2memory"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/dal-go/record/update"
)

func TestRegistrySnapshotAndComparison(t *testing.T) {
	r, digest, err := DefaultRootRegistry()
	if err != nil || len(digest) != 64 || len(r.ApprovedRoots) != 7 || len(r.LegacyRoots) != 6 || len(r.LegacyWritableRoots) != 0 {
		t.Fatalf("registry %v %s %v", r, digest, err)
	}
	got := r.CompareRoots([]string{"users", "invites", "checkoutCustomers", "newRoot", "newRoot", "datatugNeverApproved", "ext"})
	if !reflect.DeepEqual(got.Unexpected, []string{"checkoutCustomers", "datatugNeverApproved", "newRoot"}) || len(got.Legacy) != 0 {
		t.Fatal(got)
	}
	r.ApprovedRoots[0] = "mutated"
	again, _, _ := DefaultRootRegistry()
	if again.ApprovedRoots[0] == "mutated" {
		t.Fatal("mutable registry escaped")
	}
}
func TestRegistryRejectsInvalidArtifacts(t *testing.T) {
	for _, data := range []string{"", `{}`, string(rootRegistryJSON) + `{}`, strings.Replace(string(rootRegistryJSON), `"schemaVersion": 2`, `"schemaVersion": 1`, 1), strings.Replace(string(rootRegistryJSON), `"database": "(default)"`, `"database": "other"`, 1), strings.Replace(string(rootRegistryJSON), `"schemaVersion": 2`, `"unrecognized": true, "schemaVersion": 2`, 1)} {
		if _, err := LoadRootRegistry([]byte(data)); err == nil {
			t.Fatal("accepted invalid registry", data)
		}
	}
	r, _, _ := DefaultRootRegistry()
	for _, mutate := range []func(*RootRegistry){
		func(r *RootRegistry) { r.ApprovedRoots = append(r.ApprovedRoots, "users") },
		func(r *RootRegistry) { r.LegacyRoots = append(r.LegacyRoots, "users") },
		func(r *RootRegistry) { r.LegacyRoots = append(r.LegacyRoots, "bad/name") },
		func(r *RootRegistry) { r.Extensions = append(r.Extensions, r.Extensions[0]) },
		func(r *RootRegistry) { r.Extensions = append(r.Extensions, "bad/name") },
		func(r *RootRegistry) { r.Extensions = nil },
		func(r *RootRegistry) { r.SpaceOnlyExtensions = []string{"not-registered"} },
		func(r *RootRegistry) { r.SpaceOnlyExtensions = []string{"sizeus", "sizeus"} },
		func(r *RootRegistry) { r.LegacyWritableRoots = nil },
		func(r *RootRegistry) { r.LegacyWritableRoots = []string{"users"} },
		func(r *RootRegistry) { r.LegacyWritableRoots = []string{"unrecognized"} },
		func(r *RootRegistry) { r.LegacyWritableRoots = []string{"bad/name"} },
		func(r *RootRegistry) { r.LegacyWritableRoots = []string{"slugs", "slugs"} },
		func(r *RootRegistry) { r.ApprovedRoots = []string{"users"} },
	} {
		raw, _ := json.Marshal(r)
		var copy RootRegistry
		_ = json.Unmarshal(raw, &copy)
		mutate(&copy)
		if _, err := copy.RootPolicy(); err == nil {
			t.Fatal("accepted invalid policy")
		}
	}
}
func TestPolicyPlacementMatrixAndReadPassThrough(t *testing.T) {
	r, _, _ := DefaultRootRegistry()
	p, err := r.RootPolicy()
	if err != nil {
		t.Fatal(err)
	}
	keys := []struct {
		k     *record.Key
		allow bool
	}{
		{record.NewKeyWithID("referrers", "r"), true}, {record.NewKeyWithID("users", "u"), true}, {record.NewKeyWithID("invites", "i"), true}, {record.NewKeyWithID("checkoutCustomers", "c"), false},
		{record.NewKeyWithID("checkoutCustomerTypo", "c"), false}, {record.NewKeyWithID("newRoot", "id"), false},
		{record.NewKeyWithParentAndID(record.NewKeyWithID("ext", "checkout"), "customers", "c"), true},
		{record.NewKeyWithParentAndID(record.NewKeyWithID("ext", "unknown-extension"), "rows", "x"), false},
		{record.NewKeyWithParentAndID(record.NewKeyWithID("spaces", "s"), "ext", "datatug"), true},
		{record.NewKeyWithParentAndID(record.NewKeyWithID("spaces", "s"), "ext", "unknown-extension"), false},
		{record.NewKeyWithParentAndID(record.NewKeyWithID("users", "u"), "ext", "unknown-extension"), false},
		{record.NewKeyWithParentAndID(record.NewKeyWithID("users", "u"), "ext", "datatug"), true},
		{record.NewKeyWithParentAndID(record.NewKeyWithID("newRoot", "x"), "children", "c"), false},
	}
	for _, tc := range keys {
		for _, op := range []access.Operations{access.Insert, access.Set, access.Update, access.Delete, access.Truncate} {
			req := access.Request{Operation: op, Resources: []access.Resource{access.RecordResourceForKey(tc.k)}}
			allowed := p.Decide(context.Background(), req).Allowed
			if allowed != (tc.allow && op != access.Truncate) {
				t.Fatalf("%s %s allowed=%v", op, tc.k, allowed)
			}
			if (p.Authorize(context.Background(), req) == nil) != allowed {
				t.Fatal("authorize/decide disagree")
			}
		}
	}
	for _, op := range []access.Operations{access.Get, access.Exists, access.Query} {
		for _, res := range []access.Resource{access.OpaqueQuery("sql"), access.CollectionGroup("items"), access.RecordResourceForKey(record.NewKeyWithID("anything", "x"))} {
			if err := p.Authorize(context.Background(), access.Request{Operation: op, Resources: []access.Resource{res}}); err != nil {
				t.Fatal(err)
			}
		}
	}
}
func TestMandatoryPolicyTransactionsAndBatchAtomicity(t *testing.T) {
	raw := dalgo2memory.New(dalgo2memory.FirestoreProfile())
	db, err := ProtectRootCollections(raw)
	if err != nil {
		t.Fatal(err)
	}
	ctx := access.WithPolicy(context.Background(), access.MustPolicy("caller-allows", access.Root(access.Allow(access.ReadWrite))))
	good := record.NewRecordWithData(record.NewKeyWithID("users", "good"), map[string]any{"n": 1})
	bad := record.NewRecordWithData(record.NewKeyWithID("unapproved", "bad"), map[string]any{"n": 1})
	for _, op := range []func(dal.ReadwriteTransaction) error{
		func(tx dal.ReadwriteTransaction) error { return tx.Insert(ctx, bad) }, func(tx dal.ReadwriteTransaction) error { return tx.Set(ctx, bad) },
		func(tx dal.ReadwriteTransaction) error {
			return tx.Update(ctx, bad.Key(), []update.Update{update.ByFieldName("n", 2)})
		}, func(tx dal.ReadwriteTransaction) error { return tx.Delete(ctx, bad.Key()) },
		func(tx dal.ReadwriteTransaction) error { return tx.SetMulti(ctx, []record.Record{good, bad}) },
	} {
		err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error { return op(tx) })
		if !errors.Is(err, access.ErrAccessDenied) {
			t.Fatalf("denial missing: %v", err)
		}
		if exists, _ := raw.Exists(ctx, good.Key()); exists {
			t.Fatal("partial batch escaped")
		}
	}
	if err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error { return tx.Set(ctx, good) }); err != nil {
		t.Fatal(err)
	}
	if exists, _ := raw.Exists(ctx, good.Key()); !exists {
		t.Fatal("approved write lost")
	}
	// A restrictive caller policy remains restrictive even for an approved path.
	denied := access.WithPolicy(context.Background(), access.MustPolicy("caller-denies", access.Root(access.Deny(access.Write))))
	if err = db.RunReadwriteTransaction(denied, func(ctx context.Context, tx dal.ReadwriteTransaction) error { return tx.Set(ctx, good) }); !errors.Is(err, access.ErrAccessDenied) {
		t.Fatal(err)
	}
}
func TestPolicyConstructionFailures(t *testing.T) {
	old := rootRegistryJSON
	rootRegistryJSON = []byte(`{`)
	t.Cleanup(func() { rootRegistryJSON = old })
	if _, _, err := DefaultRootRegistry(); err == nil {
		t.Fatal("bad artifact accepted")
	}
	if _, err := ProtectRootCollections(dalgo2memory.New(dalgo2memory.FirestoreProfile())); err == nil {
		t.Fatal("bad artifact bypassed")
	}
	rootRegistryJSON = old
	oldCompiler := newRootAccessPolicy
	newRootAccessPolicy = func(string, ...access.Rule) (*access.AccessPolicy, error) { return nil, errors.New("compiler failed") }
	t.Cleanup(func() { newRootAccessPolicy = oldCompiler })
	if _, err := ProtectRootCollections(dalgo2memory.New(dalgo2memory.FirestoreProfile())); err == nil {
		t.Fatal("compiler error bypassed")
	}
	newRootAccessPolicy = oldCompiler
	if _, err := ProtectRootCollections(nil); err == nil {
		t.Fatal("nil DB accepted")
	}
}

// directPolicyDB exposes the optional write-session capability while persisting
// accepted batches through the real memory adapter. Calls are the denial oracle.
type directPolicyDB struct {
	dal.DB
	dal.WriteSession
	calls int
}

func (d *directPolicyDB) SetMulti(ctx context.Context, rows []record.Record) error {
	d.calls++
	return d.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error { return tx.SetMulti(ctx, rows) })
}
func (d *directPolicyDB) InsertMulti(ctx context.Context, rows []record.Record, options ...dal.InsertOption) error {
	d.calls++
	return d.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.InsertMulti(ctx, rows, options...)
	})
}

func TestGeneratedKeysAndDirectWriteSession(t *testing.T) {
	raw := dalgo2memory.New(dalgo2memory.FirestoreProfile())
	db, err := ProtectRootCollections(raw)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for _, collection := range []string{"newRoot", "ext"} {
		row := record.NewRecordWithData(record.NewIncompleteKey(collection, reflect.String, nil), map[string]any{"n": 1})
		err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			return tx.Insert(ctx, row, dal.WithAdapterGeneratedID())
		})
		if !errors.Is(err, access.ErrAccessDenied) {
			t.Fatalf("incomplete %s accepted: %v", collection, err)
		}
	}
	generated := record.NewRecordWithData(record.NewIncompleteKey("users", reflect.String, nil), map[string]any{"n": 1})
	if err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Insert(ctx, generated, dal.WithAdapterGeneratedID())
	}); err != nil {
		t.Fatal(err)
	}
	if generated.Key().ID == nil || generated.Key().ID == "" {
		t.Fatal("approved incomplete key was not generated")
	}
	direct := &directPolicyDB{DB: raw}
	secured, err := ProtectRootCollections(direct)
	if err != nil {
		t.Fatal(err)
	}
	writer, ok := secured.(dal.WriteSession)
	if !ok {
		t.Fatal("direct write capability lost")
	}
	good := record.NewRecordWithData(record.NewKeyWithID("users", "direct-good"), map[string]any{"n": 1})
	bad := record.NewRecordWithData(record.NewKeyWithID("unknownDirect", "bad"), map[string]any{"n": 1})
	for _, action := range []func() error{func() error { return writer.SetMulti(ctx, []record.Record{good, bad}) }, func() error { return writer.InsertMulti(ctx, []record.Record{good, bad}) }} {
		if err = action(); !errors.Is(err, access.ErrAccessDenied) {
			t.Fatal("direct batch accepted", err)
		}
		if direct.calls != 0 {
			t.Fatal("denied direct batch reached adapter")
		}
		if exists, _ := raw.Exists(ctx, good.Key()); exists {
			t.Fatal("partial direct batch escaped")
		}
	}
	if err = writer.SetMulti(ctx, []record.Record{good}); err != nil {
		t.Fatal(err)
	}
	if direct.calls != 1 {
		t.Fatal("approved direct batch did not reach adapter")
	}
}
