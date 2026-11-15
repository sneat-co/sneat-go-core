// Copyright 2026 Sneat Co.
package firestorepolicy

import (
	"context"
	"slices"
	"testing"

	"github.com/dal-go/dalgo/access"
	"github.com/dal-go/record"
)

// AskStage owns Space records and two global paths outside any Space: the
// public show note ext/askstage/showStamps/{stampID} and the short-address
// claims ext/askstage/slugNamespaces/show/claims/{code}. It must therefore not
// be space-only.
func TestAskstageExtensionPlacement(t *testing.T) {
	r, _, err := DefaultRootRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(r.Extensions, "askstage") {
		t.Fatal("askstage is not a registered extension")
	}
	if slices.Contains(r.SpaceOnlyExtensions, "askstage") {
		t.Fatal("askstage must not be space-only: it owns global paths")
	}
	policy, err := r.RootPolicy()
	if err != nil {
		t.Fatal(err)
	}
	global := record.NewKeyWithID("ext", "askstage")
	space := record.NewKeyWithID("spaces", "space1")
	spaceExt := record.NewKeyWithParentAndID(space, "ext", "askstage")
	tests := []struct {
		name    string
		key     *record.Key
		allowed bool
	}{
		{"global show stamp", record.NewKeyWithParentAndID(global, "showStamps", "stamp1"), true},
		{"global show code claim", record.NewKeyWithParentAndID(record.NewKeyWithParentAndID(global, "slugNamespaces", "show"), "claims", "k7f2xm"), true},
		{"space record", spaceExt, true},
		{"space show record", record.NewKeyWithParentAndID(spaceExt, "shows", "show1"), true},
		{"unknown extension", record.NewKeyWithParentAndID(record.NewKeyWithID("ext", "unknown-askstage"), "showStamps", "stamp1"), false},
		{"unapproved askstage root", record.NewKeyWithID("askstage", "show1"), false},
		{"unapproved showStamps root", record.NewKeyWithID("showStamps", "stamp1"), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, op := range []access.Operations{access.Insert, access.Set, access.Update, access.Delete, access.Truncate} {
				request := access.Request{Operation: op, Resources: []access.Resource{access.RecordResourceForKey(tc.key)}}
				if got := policy.Decide(context.Background(), request).Allowed; got != (tc.allowed && op != access.Truncate) {
					t.Fatalf("%s %s allowed=%v", op, tc.key, got)
				}
			}
		})
	}
}
