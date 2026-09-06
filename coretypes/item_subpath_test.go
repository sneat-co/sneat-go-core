// Copyright 2026 Sneat Co.
package coretypes

import (
	"encoding/json"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestItemSubPathStableIdentity(t *testing.T) {
	segments := []ItemSubPathSegment{{Field: "items"}, {Key: "id", Value: "a/b~c=@x"}}
	path, err := FormatItemSubPath(segments...)
	if err != nil || path != "/items/@id=a~1b~0c=@x" {
		t.Fatalf("format: %q, %v", path, err)
	}
	got, err := ParseItemSubPath(path)
	if err != nil || !reflect.DeepEqual(got, segments) {
		t.Fatalf("parse: %#v, %v", got, err)
	}
	ref := ItemRef{ExtID: "listus", Collection: "lists", ItemID: "list@space", SubPath: path}
	if err := ref.Validate(); err != nil {
		t.Fatal(err)
	}
	if ref.DocID() != "list" {
		t.Fatalf("subpath changed containing document: %s", ref.DocID())
	}
	query, err := url.ParseQuery(ref.ID())
	if err != nil {
		t.Fatal(err)
	}
	roundTrip, err := NewItemRefFromQueryString(query)
	if err != nil || roundTrip != ref {
		t.Fatalf("reference roundtrip: %#v, %v", roundTrip, err)
	}
	other := ref
	other.SubPath = "/items/@id=other"
	if other.ID() == ref.ID() {
		t.Fatal("different embedded items share identity")
	}
	other.SubPath = ""
	if other.ID() != "m=listus&c=lists&i=list@space" {
		t.Fatal("legacy identity changed")
	}
	// A legacy identifier containing query punctuation cannot impersonate a
	// scoped reference, even though old RecordID validation accepts punctuation.
	other.ItemID += "&p=" + url.QueryEscape(path)
	if other.ID() == ref.ID() {
		t.Fatal("legacy document ID collides with embedded identity")
	}
	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ItemRef
	if err := json.Unmarshal(data, &decoded); err != nil || decoded != ref {
		t.Fatalf("JSON roundtrip: %#v, %v", decoded, err)
	}
	other.SubPath = ""
	data, _ = json.Marshal(other)
	if strings.Contains(string(data), "subPath") {
		t.Fatal("legacy JSON must omit subPath")
	}
}

func TestItemSubPathRejectsUnstableAndExecutablePaths(t *testing.T) {
	for _, path := range []string{
		"/", "items/@id=a", "/items/0", "/items/*", "/items/../secret", "/@id=a",
		"/items/@id=", "/items/@=a", "/items/@id=a~", "/items/@id=a~2", "/items/@id= a",
		"/items/@id=a\n", "/items/@id=\xff", "/items[?(@.id=='a')]", "/items//id",
		strings.Repeat("/field", ItemSubPathMaxSegments+1), "/items/@id=" + strings.Repeat("a", 513),
	} {
		t.Run(path, func(t *testing.T) {
			if _, err := ParseItemSubPath(path); err == nil {
				t.Fatalf("accepted invalid path %q", path)
			}
			ref := ItemRef{ExtID: "listus", Collection: "lists", ItemID: "a", SubPath: path}
			if ref.Validate() == nil {
				t.Fatal("ItemRef accepted invalid path")
			}
			if _, err := NewItemRefFromQueryString(url.Values{"m": {"listus"}, "c": {"lists"}, "i": {"a"}, "p": {path}}); err == nil {
				t.Fatal("query accepted invalid path")
			}
		})
	}
	for _, path := range []string{"", "/items/@id=a", "/latestReceipt/obligations/@lineID=line1"} {
		if _, err := ParseItemSubPath(path); err != nil {
			t.Fatalf("valid path %q: %v", path, err)
		}
	}
}

func TestItemSubPathRejectsAmbiguousSegments(t *testing.T) {
	for _, segments := range [][]ItemSubPathSegment{
		{{Field: "items", Key: "id", Value: "a"}}, {{Key: "id", Value: "a"}}, {{Field: "items"}, {}},
	} {
		if _, err := FormatItemSubPath(segments...); err == nil {
			t.Fatalf("accepted %#v", segments)
		}
	}
}
