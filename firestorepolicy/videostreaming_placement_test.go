// Copyright 2026 Sneat Co.
package firestorepolicy

import (
	"context"
	"slices"
	"testing"

	"github.com/dal-go/dalgo/access"
	"github.com/dal-go/record"
)

// Video streaming channels outlive any Space, so the module owns the global
// path ext/videostreaming/platforms/{platform}/channels/{channelID} and its
// destinations, backgrounds and streamEvents subcollections. It must not be
// space-only.
func TestVideostreamingExtensionPlacement(t *testing.T) {
	r, _, err := DefaultRootRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(r.Extensions, "videostreaming") {
		t.Fatal("videostreaming is not a registered extension")
	}
	if slices.Contains(r.SpaceOnlyExtensions, "videostreaming") {
		t.Fatal("videostreaming must not be space-only: it owns global channel records")
	}
	policy, err := r.RootPolicy()
	if err != nil {
		t.Fatal(err)
	}
	global := record.NewKeyWithID("ext", "videostreaming")
	channel := record.NewKeyWithParentAndID(record.NewKeyWithParentAndID(global, "platforms", "youtube"), "channels", "UCa1B2c3D4e5F6g7H8i9J0kL")
	tests := []struct {
		name    string
		key     *record.Key
		allowed bool
	}{
		{"channel", channel, true},
		{"destination", record.NewKeyWithParentAndID(channel, "destinations", "dV3r9x"), true},
		{"background", record.NewKeyWithParentAndID(channel, "backgrounds", "bg_k2Pz"), true},
		{"stream event", record.NewKeyWithParentAndID(channel, "streamEvents", "ev1"), true},
		{"unknown extension", record.NewKeyWithParentAndID(record.NewKeyWithParentAndID(record.NewKeyWithID("ext", "unknown-videostreaming"), "platforms", "youtube"), "channels", "x"), false},
		{"unapproved videostreaming root", record.NewKeyWithID("videostreaming", "x"), false},
		{"unapproved channels root", record.NewKeyWithID("channels", "UCa1B2c3D4e5F6g7H8i9J0kL"), false},
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
