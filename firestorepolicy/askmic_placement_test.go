// Copyright 2026 Sneat Co.
package firestorepolicy

import (
	"context"
	"slices"
	"testing"

	"github.com/dal-go/dalgo/access"
	"github.com/dal-go/record"
)

// AskMic owns Space records and two global paths outside any Space: the
// public show note ext/askmic/showStamps/{stampID} and the short-address
// claims ext/askmic/slugNamespaces/show/claims/{code}. It must therefore not
// be space-only.
func TestAskmicExtensionPlacement(t *testing.T) {
	r, _, err := DefaultRootRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(r.Extensions, "askmic") {
		t.Fatal("askmic is not a registered extension")
	}
	if slices.Contains(r.SpaceOnlyExtensions, "askmic") {
		t.Fatal("askmic must not be space-only: it owns global paths")
	}
	policy, err := r.RootPolicy()
	if err != nil {
		t.Fatal(err)
	}
	global := record.NewKeyWithID("ext", "askmic")
	space := record.NewKeyWithID("spaces", "space1")
	spaceExt := record.NewKeyWithParentAndID(space, "ext", "askmic")
	tests := []struct {
		name    string
		key     *record.Key
		allowed bool
	}{
		{"global show stamp", record.NewKeyWithParentAndID(global, "showStamps", "stamp1"), true},
		{"global show code claim", record.NewKeyWithParentAndID(record.NewKeyWithParentAndID(global, "slugNamespaces", "show"), "claims", "k7f2xm"), true},
		{"space record", spaceExt, true},
		{"space show record", record.NewKeyWithParentAndID(spaceExt, "shows", "show1"), true},
		{"unknown extension", record.NewKeyWithParentAndID(record.NewKeyWithID("ext", "unknown-askmic"), "showStamps", "stamp1"), false},
		{"unapproved askmic root", record.NewKeyWithID("askmic", "show1"), false},
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
