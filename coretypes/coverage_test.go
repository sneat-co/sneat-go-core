package coretypes

import (
	"strings"
	"testing"
)

func TestEntityRef_ComprehensiveCoverage(t *testing.T) {
	// NewEntityRef
	ref := NewEntityRef("contact", "123")
	if string(ref) != "contact:123" {
		t.Fatalf("unexpected ref: %s", ref)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on empty kind")
			}
		}()
		NewEntityRef("", "123")
	}()
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on empty entityID")
			}
		}()
		NewEntityRef("contact", "")
	}()

	// NewEntityRefWithSpaceID
	refWithSpace := NewEntityRefWithSpaceID("contact", "123", "space1")
	if string(refWithSpace) != "contact:123@space1" {
		t.Fatalf("unexpected ref with space: %s", refWithSpace)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on empty spaceID")
			}
		}()
		NewEntityRefWithSpaceID("contact", "123", "")
	}()

	// AddSpaceID
	refAdded := EntityRef("contact:123").AddSpaceID("space2")
	if string(refAdded) != "contact:123@space2" {
		t.Fatalf("unexpected refAdded: %s", refAdded)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on empty spaceID in AddSpaceID")
			}
		}()
		EntityRef("contact:123").AddSpaceID("")
	}()

	// Kind and ID without separator
	noSep := EntityRef("noseparator")
	if noSep.Kind() != "" {
		t.Fatalf("expected empty Kind, got %s", noSep.Kind())
	}
	if noSep.ID() != "" {
		t.Fatalf("expected empty ID, got %s", noSep.ID())
	}

	// Validate edge cases
	if err := EntityRef("   ").Validate(); err == nil {
		t.Fatal("expected error on all whitespace")
	}
	if err := EntityRef(" contact:123").Validate(); err == nil {
		t.Fatal("expected error on leading whitespace")
	}
	if err := EntityRef("contact123").Validate(); err == nil {
		t.Fatal("expected error on missing separator")
	}
	if err := EntityRef(":123").Validate(); err == nil {
		t.Fatal("expected error on missing kind")
	}
	if err := EntityRef("contact:").Validate(); err == nil {
		t.Fatal("expected error on missing ID")
	}
	if err := EntityRef("contact:@space1").Validate(); err == nil {
		t.Fatal("expected error on missing ID with space")
	}
}

func TestItemRef_StringCoverage(t *testing.T) {
	ref := ItemRef{
		ExtID:      "listus",
		Collection: "lists",
		ItemID:     "item1",
		SubPath:    "/items/@id=x",
	}
	expected := "{ExtID=listus,Collection=lists,ItemID=item1,SubPath=/items/@id=x}"
	if ref.String() != expected {
		t.Fatalf("expected %s, got %s", expected, ref.String())
	}
}

func TestItemSubPath_Coverage(t *testing.T) {
	// Too many segments in FormatItemSubPath
	tooMany := make([]ItemSubPathSegment, ItemSubPathMaxSegments+1)
	for i := range tooMany {
		tooMany[i] = ItemSubPathSegment{Field: "f"}
	}
	if _, err := FormatItemSubPath(tooMany...); err == nil {
		t.Fatal("expected error on too many segments in FormatItemSubPath")
	}

	// Empty parts in FormatItemSubPath
	p, err := FormatItemSubPath()
	if err != nil || p != "" {
		t.Fatalf("expected empty path and nil err, got %q, %v", p, err)
	}

	// Format exceeds max bytes
	hugeSegments := []ItemSubPathSegment{
		{Field: "items"},
		{Key: "id", Value: strings.Repeat("a", 500)},
		{Key: "id", Value: strings.Repeat("b", 500)},
		{Key: "id", Value: strings.Repeat("c", 500)},
		{Key: "id", Value: strings.Repeat("d", 500)},
		{Key: "id", Value: strings.Repeat("e", 500)},
	}
	if _, err := FormatItemSubPath(hugeSegments...); err == nil {
		t.Fatal("expected error on exceeding max bytes in FormatItemSubPath")
	}

	// Invalid selector without '='
	if _, err := ParseItemSubPath("/items/@noequal"); err == nil {
		t.Fatal("expected error on selector without '='")
	}

	// Control character in selector value
	ctrlSegment := []ItemSubPathSegment{
		{Field: "items"},
		{Key: "id", Value: "hello\tworld"},
	}
	if _, err := FormatItemSubPath(ctrlSegment...); err == nil {
		t.Fatal("expected error on control character in selector value")
	}
}

func TestSpaceRef_Coverage(t *testing.T) {
	// SpaceRef without separator
	personalRef := SpaceRef("personal")
	if personalRef.SpaceType() != SpaceTypePersonal {
		t.Fatalf("expected SpaceTypePersonal, got %s", personalRef.SpaceType())
	}
	if personalRef.SpaceID() != "" {
		t.Fatalf("expected empty SpaceID, got %s", personalRef.SpaceID())
	}

	invalidRef := SpaceRef("custom_space_id")
	if invalidRef.SpaceType() != "" {
		t.Fatalf("expected empty SpaceType, got %s", invalidRef.SpaceType())
	}
	if invalidRef.SpaceID() != "custom_space_id" {
		t.Fatalf("expected SpaceID 'custom_space_id', got %s", invalidRef.SpaceID())
	}
}

func TestWithSharedMap_Comprehensive(t *testing.T) {
	// ShareDetail.Validate
	sd := ShareDetail{}
	if err := sd.Validate(); err == nil {
		t.Fatal("expected error on empty ByUserID")
	}
	sd.ByUserID = "user1"
	if err := sd.Validate(); err == nil {
		t.Fatal("expected error on empty Permissions")
	}
	sd.Permissions = []Permission{""}
	if err := sd.Validate(); err == nil {
		t.Fatal("expected error on empty permission string")
	}
	sd.Permissions = []Permission{" read "}
	if err := sd.Validate(); err == nil {
		t.Fatal("expected error on whitespace in permission")
	}
	sd.Permissions = []Permission{"read", "write"}
	if err := sd.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// WithSharedMap.Validate
	wsm := &WithSharedMap{}
	if err := wsm.Validate(); err != nil {
		t.Fatalf("unexpected error on empty WithSharedMap: %v", err)
	}

	// Validate SharedTo errors
	wsm.SharedTo = SharedToMap{
		" badRef": map[string]ShareDetail{
			"space1": sd,
		},
	}
	if err := wsm.Validate(); err == nil {
		t.Fatal("expected error on entityRef with leading space in SharedTo")
	}
	wsm.SharedTo = SharedToMap{
		"invalidRef": map[string]ShareDetail{
			"space1": sd,
		},
	}
	if err := wsm.Validate(); err == nil {
		t.Fatal("expected error on invalid entityRef in SharedTo")
	}
	wsm.SharedTo = SharedToMap{
		"contact:123": map[string]ShareDetail{
			"space1": {ByUserID: ""},
		},
	}
	if err := wsm.Validate(); err == nil {
		t.Fatal("expected error on invalid ShareDetail in SharedTo")
	}
	wsm.SharedTo = SharedToMap{
		"contact:123": map[string]ShareDetail{
			"space1": sd,
		},
	}
	if err := wsm.Validate(); err != nil {
		t.Fatalf("unexpected error on valid SharedTo: %v", err)
	}

	// Validate SharedFrom errors
	wsm.SharedTo = nil
	wsm.SharedFrom = SharedFromMap{
		" space1 ": map[string]ShareDetail{
			"contact:123": sd,
		},
	}
	if err := wsm.Validate(); err == nil {
		t.Fatal("expected error on spaceID with leading/trailing space in SharedFrom")
	}
	wsm.SharedFrom = SharedFromMap{
		"space1": map[string]ShareDetail{
			"invalidRef": sd,
		},
	}
	if err := wsm.Validate(); err == nil {
		t.Fatal("expected error on invalid entityRef in SharedFrom")
	}
	wsm.SharedFrom = SharedFromMap{
		"space1": map[string]ShareDetail{
			"contact:123": {ByUserID: ""},
		},
	}
	if err := wsm.Validate(); err == nil {
		t.Fatal("expected error on invalid ShareDetail in SharedFrom")
	}
	wsm.SharedFrom = SharedFromMap{
		"space1": map[string]ShareDetail{
			"contact:123": sd,
		},
	}
	if err := wsm.Validate(); err != nil {
		t.Fatalf("unexpected error on valid SharedFrom: %v", err)
	}

	// AddSharedTo panics
	wsm = &WithSharedMap{}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on empty entityRef in AddSharedTo")
			}
		}()
		_, _ = wsm.AddSharedTo("", "space1", sd)
	}()
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on empty spaceID in AddSharedTo")
			}
		}()
		_, _ = wsm.AddSharedTo("contact:123", "", sd)
	}()

	// AddSharedTo first time
	u, err := wsm.AddSharedTo("contact:123", "space1", sd)
	if err != nil || u == nil {
		t.Fatalf("expected update, got err: %v", err)
	}
	// AddSharedTo identical (no-op)
	u, err = wsm.AddSharedTo("contact:123", "space1", sd)
	if err != nil || u != nil {
		t.Fatalf("expected nil update for duplicate, got %v, %v", u, err)
	}
	// AddSharedTo modified
	sd2 := sd
	sd2.ByUserID = "user2"
	u, err = wsm.AddSharedTo("contact:123", "space1", sd2)
	if err != nil || u == nil {
		t.Fatalf("expected update on change, got %v, %v", u, err)
	}

	// AddSharedFrom panics
	wsm = &WithSharedMap{}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on empty spaceID in AddSharedFrom")
			}
		}()
		_, _ = wsm.AddSharedFrom("", "contact:123", sd)
	}()
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on empty entityRef in AddSharedFrom")
			}
		}()
		_, _ = wsm.AddSharedFrom("space1", "", sd)
	}()

	// AddSharedFrom first time
	u, err = wsm.AddSharedFrom("space1", "contact:123", sd)
	if err != nil || u == nil {
		t.Fatalf("expected update, got err: %v", err)
	}
	// AddSharedFrom identical (no-op)
	u, err = wsm.AddSharedFrom("space1", "contact:123", sd)
	if err != nil || u != nil {
		t.Fatalf("expected nil update for duplicate, got %v, %v", u, err)
	}
	// AddSharedFrom modified
	u, err = wsm.AddSharedFrom("space1", "contact:123", sd2)
	if err != nil || u == nil {
		t.Fatalf("expected update on change, got %v, %v", u, err)
	}
}
