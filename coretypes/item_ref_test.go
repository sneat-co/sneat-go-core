package coretypes

import (
	"net/url"
	"testing"
)

// TestItemRef_Validate covers the "@" reserved-separator rules from sneat-specs
// Decision 0002: a bare itemID and a well-formed "itemID@{spaceID}" composite
// are valid; a trailing "@" (empty space), a leading "@" (empty doc id), and
// more than one "@" (an embedded separator in the document id) are rejected.
func TestItemRef_Validate(t *testing.T) {
	newRef := func(itemID string) ItemRef {
		return ItemRef{ExtID: "contactus", Collection: "contacts", ItemID: itemID}
	}
	tests := []struct {
		name    string
		ref     ItemRef
		wantErr bool
	}{
		{name: "bare", ref: newRef("item1"), wantErr: false},
		{name: "space_bound_composite", ref: newRef("item1@space1"), wantErr: false},
		{name: "trailing_at_empty_space", ref: newRef("item1@"), wantErr: true},
		{name: "leading_at_empty_doc", ref: newRef("@space1"), wantErr: true},
		{name: "double_at", ref: newRef("item1@space1@x"), wantErr: true},
		{name: "missing_item_id", ref: newRef(""), wantErr: true},
		{name: "missing_ext_id", ref: ItemRef{Collection: "contacts", ItemID: "item1"}, wantErr: true},
		{name: "missing_collection", ref: ItemRef{ExtID: "contactus", ItemID: "item1"}, wantErr: true},
		{name: "doc_id_contains_whitespace", ref: newRef("item 1"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ref.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("ItemRef.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestItemRef_DocID verifies stripping of the optional "@{spaceID}" suffix.
func TestItemRef_DocID(t *testing.T) {
	tests := []struct {
		itemID string
		want   string
	}{
		{itemID: "item1", want: "item1"},
		{itemID: "item1@space1", want: "item1"},
		{itemID: "item1@", want: "item1"},
	}
	for _, tt := range tests {
		ref := ItemRef{ExtID: "contactus", Collection: "contacts", ItemID: tt.itemID}
		if got := ref.DocID(); got != tt.want {
			t.Errorf("ItemRef.DocID() for %q = %q, want %q", tt.itemID, got, tt.want)
		}
	}
}

func TestItemRef_ID(t *testing.T) {
	ref := ItemRef{ExtID: "contactus", Collection: "contacts", ItemID: "item1"}
	want := "m=contactus&c=contacts&i=item1"
	if got := ref.ID(); got != want {
		t.Errorf("ItemRef.ID() = %q, want %q", got, want)
	}
}

func TestItemRef_String(t *testing.T) {
	ref := ItemRef{ExtID: "contactus", Collection: "contacts", ItemID: "item1"}
	want := "{ExtID=contactus,Collection=contacts,ItemID=item1}"
	if got := ref.String(); got != want {
		t.Errorf("ItemRef.String() = %q, want %q", got, want)
	}
}

// TestNewFullItemRef ports dbo4linkage's TestNewContactFullRef: a spaceID
// appends the "@{spaceID}" suffix, an empty spaceID resolves to the spaceless
// system namespace (sneat-specs Decision 0002), and an empty itemID panics.
func TestNewFullItemRef(t *testing.T) {
	type args struct {
		spaceID SpaceID
		itemID  string
	}
	tests := []struct {
		name string
		args args
		want ItemRef
	}{
		{
			name: "normal",
			args: args{
				spaceID: "test-space-id",
				itemID:  "test-contact-id",
			},
			want: ItemRef{
				ExtID:      "contactus",
				Collection: "contacts",
				ItemID:     "test-contact-id@test-space-id",
			},
		},
		{
			// sneat-specs Decision 0002: an empty spaceID denotes the spaceless
			// system namespace, so no "@{spaceID}" suffix is appended (no panic).
			name: "empty space resolves to system namespace",
			args: args{
				spaceID: "",
				itemID:  "test-contact-id",
			},
			want: ItemRef{
				ExtID:      "contactus",
				Collection: "contacts",
				ItemID:     "test-contact-id",
			},
		},
		{
			name: "panic on empty itemID",
			args: args{
				spaceID: "test-space-id",
				itemID:  "",
			},
			want: ItemRef{
				ExtID:      "contactus",
				Collection: "contacts",
				ItemID:     "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.itemID == "" {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("NewFullItemRef() did not panic on empty itemID")
					}
				}()
			}
			if got := NewFullItemRef("contactus", "contacts", tt.args.spaceID, tt.args.itemID); got != tt.want {
				t.Errorf("NewFullItemRef() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewItemRefSameSpace(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		got := NewItemRefSameSpace("contactus", "contacts", "item1")
		want := ItemRef{ExtID: "contactus", Collection: "contacts", ItemID: "item1"}
		if got != want {
			t.Errorf("NewItemRefSameSpace() = %v, want %v", got, want)
		}
	})
	t.Run("panics_on_space_suffix", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("expected panic on itemID containing '@'")
			}
		}()
		NewItemRefSameSpace("contactus", "contacts", "item1@space1")
	})
	t.Run("panics_on_empty_ext_id", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("expected panic on empty extID")
			}
		}()
		NewItemRefSameSpace("", "contacts", "item1")
	})
	t.Run("panics_on_empty_collection", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("expected panic on empty collection")
			}
		}()
		NewItemRefSameSpace("contactus", "", "item1")
	})
	t.Run("panics_on_empty_item_id", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("expected panic on empty itemID")
			}
		}()
		NewItemRefSameSpace("contactus", "contacts", "")
	})
}

// TestNewItemRefFromQueryString_BothShapes verifies that the presence of the
// "s" (space) query parameter is the sole discriminator (sneat-specs Decision
// 0002): when present, the itemID carries an "@{spaceID}" suffix (space-bound);
// when absent, the itemID stays bare (spaceless system namespace). Ported from
// dbo4linkage's with_related_and_ids_test.go.
func TestNewItemRefFromQueryString_BothShapes(t *testing.T) {
	tests := []struct {
		name       string
		values     url.Values
		wantItemID string
	}{
		{
			name:       "space_bound",
			values:     url.Values{"m": {"contactus"}, "c": {"contacts"}, "i": {"item1"}, "s": {"space1"}},
			wantItemID: "item1@space1",
		},
		{
			name:       "system_namespace_no_space",
			values:     url.Values{"m": {"contactus"}, "c": {"contacts"}, "i": {"item1"}},
			wantItemID: "item1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := NewItemRefFromQueryString(tt.values)
			if err != nil {
				t.Fatalf("NewItemRefFromQueryString() error = %v", err)
			}
			if ref.ItemID != tt.wantItemID {
				t.Errorf("NewItemRefFromQueryString() ItemID = %q, want %q", ref.ItemID, tt.wantItemID)
			}
		})
	}
}

func TestNewItemRefFromQueryString_Errors(t *testing.T) {
	tests := []struct {
		name   string
		values url.Values
	}{
		{name: "missing_module", values: url.Values{"c": {"contacts"}, "i": {"item1"}}},
		{name: "missing_collection", values: url.Values{"m": {"contactus"}, "i": {"item1"}}},
		{name: "missing_item_id", values: url.Values{"m": {"contactus"}, "c": {"contacts"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewItemRefFromQueryString(tt.values); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}
