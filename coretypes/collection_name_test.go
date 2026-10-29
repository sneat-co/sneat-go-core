package coretypes

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/dal-go/dalgo/access"
	"github.com/dal-go/record"
)

func TestCollectionNamespaceRejectedAtEveryReferenceBoundary(t *testing.T) {
	for _, collection := range []string{"", "/", "one/two", "/one", "one/", "one//two"} {
		t.Run(collection, func(t *testing.T) {
			if ValidateCollectionName(collection) == nil {
				t.Fatal("accepted a collection outside the single-component contract")
			}
			ref := ItemRef{ExtID: "example", Collection: collection, ItemID: "item1"}
			if ref.Validate() == nil {
				t.Fatal("ItemRef accepted invalid collection")
			}
			declaration := ModuleCollectionRef{ModuleID: ref.ExtID, Collection: collection}
			if declaration.Validate() == nil {
				t.Fatal("protected declaration accepted invalid collection")
			}
			if _, err := NewItemRefFromQueryString(url.Values{"m": {"example"}, "c": {collection}, "i": {"item1"}}); err == nil {
				t.Fatal("query parser accepted invalid collection")
			}
			for _, construct := range []func(){
				func() { NewItemRefSameSpace("example", collection, "item1") },
				func() { NewFullItemRef("example", collection, "space1", "item1") },
				func() { NewFullItemRef("example", collection, "", "item1") },
			} {
				func() {
					defer func() {
						if recover() == nil {
							t.Error("constructor accepted invalid collection")
						}
					}()
					construct()
				}()
			}
			if ref.Collection != collection || declaration.Collection != collection {
				t.Fatal("validation changed the supplied namespace")
			}
		})
	}
}

func TestCollectionNamespacePreservesValidComponentsAndIdentity(t *testing.T) {
	// Existing product collections and literal punctuation remain supported.
	// Percent sequences and backslashes are not decoded or interpreted as paths.
	for _, collection := range []string{"contacts", "projects", "happenings", "spaces", "read_models", "report-v1", "réponses", "a.b", "a%2Fb", `a\b`, "two words"} {
		t.Run(collection, func(t *testing.T) {
			if err := ValidateCollectionName(collection); err != nil {
				t.Fatal(err)
			}
			ref := NewFullItemRef("example", collection, "space1", "item1")
			ref.SubPath = "/items/@id=a~1b"
			if err := ref.Validate(); err != nil {
				t.Fatal(err)
			}
			declaration := ModuleCollectionRef{ModuleID: ref.ExtID, Collection: collection}
			if err := declaration.Validate(); err != nil {
				t.Fatal(err)
			}
			values, err := url.ParseQuery(ref.ID())
			if err != nil {
				t.Fatal(err)
			}
			roundTrip, err := NewItemRefFromQueryString(values)
			if err != nil || roundTrip != ref {
				t.Fatalf("reference changed during round trip: %#v, %v", roundTrip, err)
			}
			key := record.NewKeyWithParentAndID(NewSpaceModuleKey("space1", ref.ExtID), collection, ref.DocID())
			parts := strings.Split(key.String(), "/")
			if len(parts) != 6 || parts[4] != collection || key.Collection() != declaration.Collection {
				t.Fatalf("logical and serialized collection differ: %v", parts)
			}
			resource := access.RecordResourceForKey(key)
			if resource.String() != "/"+key.String() {
				t.Fatal("policy and storage target differ")
			}
			policy, err := access.NewPolicy("exact-collection", access.Scope("spaces", "space1", access.Scope("ext", ref.ExtID, access.Collection(collection, access.Allow(access.Update)))))
			if err != nil {
				t.Fatal(err)
			}
			request := access.Request{Operation: access.Update, Resources: []access.Resource{resource}}
			if err := policy.Authorize(context.Background(), request); err != nil {
				t.Fatalf("owning collection policy no longer matches its storage key: %v", err)
			}
			other := record.NewKeyWithParentAndID(key.Parent(), "other", ref.DocID())
			request.Resources = []access.Resource{access.RecordResourceForKey(other)}
			if err := policy.Authorize(context.Background(), request); err == nil {
				t.Fatal("exact collection policy accepted a different collection")
			}
		})
	}
}

func TestCollectionNamespacePreservesEscapedIDsAndExistingNameRules(t *testing.T) {
	// Collection confinement does not add provider-specific character rules or
	// change the separate escaping of extension/document IDs.
	for _, collection := range []string{" ", " leading", ".", "..", "a\x00b", "a\xffb"} {
		if err := ValidateCollectionName(collection); err != nil {
			t.Fatalf("added an unrelated collection naming restriction: %v", err)
		}
		ref := NewItemRefSameSpace("example", collection, "item1")
		if err := ref.Validate(); err != nil {
			t.Fatal(err)
		}
	}
	ref := NewFullItemRef("example/name", "items", "space1", "item/name")
	if err := ref.Validate(); err != nil {
		t.Fatal(err)
	}
	if ref.ExtID != "example/name" || ref.DocID() != "item/name" || ref.ItemID != "item/name@space1" {
		t.Fatal("collection validation changed an ID or Space suffix")
	}
	key := record.NewKeyWithParentAndID(NewSpaceModuleKey("space1", ref.ExtID), ref.Collection, ref.DocID())
	if key.String() != "spaces/space1/ext/example%2Fname/items/item%2Fname" {
		t.Fatal("ID escaping changed")
	}
	// The query parser retains its previous blank-name rejection.
	if _, err := NewItemRefFromQueryString(url.Values{"m": {"example"}, "c": {" "}, "i": {"item1"}}); err == nil {
		t.Fatal("query parser lost its existing blank-name rejection")
	}
}
