// Copyright 2026 Sneat Co.
package firestorepolicy

import (
	"context"
	"testing"

	"github.com/dal-go/dalgo/access"
	"github.com/dal-go/record"
)

func TestSizeusExtensionPlacement(t *testing.T) {
	r, _, err := DefaultRootRegistry()
	if err != nil {
		t.Fatal(err)
	}
	policy, err := r.RootPolicy()
	if err != nil {
		t.Fatal(err)
	}
	space := record.NewKeyWithID("spaces", "space1")
	// Matches Sizeus dbo4sizeus.NewSizeRecordKey and its space-module key.
	size := record.NewKeyWithParentAndID(record.NewKeyWithParentAndID(space, "ext", "sizeus"), "sizes", "size1")
	tests := []struct {
		name    string
		key     *record.Key
		allowed bool
	}{
		{"owned size", size, true},
		{"unknown extension", record.NewKeyWithParentAndID(record.NewKeyWithParentAndID(space, "ext", "unknown-sizeus"), "sizes", "size1"), false},
		{"unapproved sizeus root", record.NewKeyWithID("sizeus", "size1"), false},
		{"unapproved sizes root", record.NewKeyWithID("sizes", "size1"), false},
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
