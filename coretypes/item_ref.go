package coretypes

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/sneat-co/sneat-go-core/validate"
	"github.com/strongo/validation"
)

// SpaceItemIDSeparator is the reserved separator used to append an optional
// "@{spaceID}" suffix to an ItemRef's ItemID (sneat-specs Decision 0002):
// https://github.com/sneat-co/sneat-specs/blob/main/spec/decisions/0002-reserved-extension-space-ids.md
const SpaceItemIDSeparator = "@"

// ItemRef is a persistence-neutral reference to an item owned by an
// extension's collection, e.g. a Contactus contact or a Calendarius
// happening. It was promoted from
// sneat-core-modules/linkage/dbo4linkage to sneat-go-core/coretypes to break
// a module import cycle between sneat-core-modules and ext-contactus/backend;
// sneat-core-modules keeps its own ItemRef as a type alias to this one.
type ItemRef struct {
	ExtID      ExtID  `json:"module" firestore:"module"` // TODO: change to `json:"extID" firestore:"extID"`?
	Collection string `json:"collection" firestore:"collection"`
	ItemID     string `json:"itemID" firestore:"itemID"`
	//SpaceID    SpaceID  `json:"spaceID,omitempty" firestore:"spaceID,omitempty"`
}

// NewItemRefSameSpace returns an ItemRef for an item that lives in the same
// space as the caller, so itemID must not carry a "@{spaceID}" suffix.
func NewItemRefSameSpace(extID ExtID, collection, itemID string) ItemRef {
	if strings.Contains(itemID, SpaceItemIDSeparator) {
		panic("itemID must not contain a spaceID separated by '@'")
	}
	return newItemRef(extID, collection, itemID)
}

func newItemRef(extID ExtID, collection, itemID string) ItemRef {
	if extID == "" {
		panic("extID is required")
	}
	if collection == "" {
		panic("collection is required")
	}
	if itemID == "" {
		panic("itemID is required")
	}
	return ItemRef{
		//SpaceID:    spaceID,
		ExtID:      extID,
		Collection: collection,
		ItemID:     itemID,
	}
}

// NewFullItemRef returns an ItemRef in an arbitrary space.
//
// specscore: decisions/0002-reserved-extension-space-ids
// Ref serialization: omit the "@{spaceID}" suffix for the spaceless system
// namespace. See sneat-specs Decision 0002:
// https://github.com/sneat-co/sneat-specs/blob/main/spec/decisions/0002-reserved-extension-space-ids.md
func NewFullItemRef(extID ExtID, collection string, spaceID SpaceID, itemID string) ItemRef {
	if itemID == "" {
		panic("itemID is required for a full item reference")
	}
	if spaceID == "" {
		// Spaceless system namespace: no "@{spaceID}" suffix is appended.
		// The record resolves under /ext/{ext-id}/{collection}/{item-id}.
		return newItemRef(extID, collection, itemID)
	}
	return newItemRef(extID, collection, itemID+SpaceItemIDSeparator+string(spaceID))
}

// NewItemRefFromQueryString builds an ItemRef from URL query parameters:
// "m" (extension/module ID), "c" (collection), "i" (item ID), and an optional
// "s" (space ID) that gets appended to the item ID as a "@{spaceID}" suffix.
func NewItemRefFromQueryString(values url.Values) (itemRef ItemRef, err error) {
	if itemRef.ExtID = ExtID(values.Get("m")); strings.TrimSpace(string(itemRef.ExtID)) == "" {
		return itemRef, errors.New("extension ID 'm' parameter is required")
	}
	if itemRef.Collection = values.Get("c"); strings.TrimSpace(itemRef.Collection) == "" {
		return itemRef, errors.New("collectionID 'c' parameter is required")
	}
	if itemRef.ItemID = values.Get("i"); strings.TrimSpace(itemRef.ItemID) == "" {
		return itemRef, errors.New("itemID 'i' parameter is required")
	}
	if spaceID := values.Get("s"); spaceID != "" {
		itemRef.ItemID = itemRef.ItemID + SpaceItemIDSeparator + spaceID
	}
	return
}

// ID returns a stable, order-sensitive string identifier for the item
// reference. The order is important for the RelatedIDs field.
func (v ItemRef) ID() string {
	return fmt.Sprintf("m=%s&c=%s&i=%s", v.ExtID, v.Collection, v.ItemID)
}

// String implements fmt.Stringer.
func (v ItemRef) String() string {
	return fmt.Sprintf("{ExtID=%s,Collection=%s,ItemID=%s}", v.ExtID, v.Collection, v.ItemID)
}

// DocID returns the bare document id with any optional "@{spaceID}" suffix
// removed. For a bare itemID it returns the itemID unchanged.
func (v ItemRef) DocID() string {
	if i := strings.Index(v.ItemID, SpaceItemIDSeparator); i >= 0 {
		return v.ItemID[:i]
	}
	return v.ItemID
}

// Validate checks that the item reference has all required fields and that
// any optional "@{spaceID}" suffix on ItemID is well-formed.
func (v ItemRef) Validate() error {
	// SpaceID can be empty for global collections like Happening
	if v.ExtID == "" {
		return validation.NewErrRecordIsMissingRequiredField("moduleID")
	}
	if v.Collection == "" {
		return validation.NewErrRecordIsMissingRequiredField("collection")
	}
	if v.ItemID == "" {
		return validation.NewErrRecordIsMissingRequiredField("itemID")
	}
	// The ItemID may carry at most one optional "@{spaceID}" suffix (sneat-specs
	// Decision 0002). "@" is the reserved space separator, so a document id must
	// never contain it: split on the first "@" and require both the document id
	// and the spaceID segment to be non-empty, and the spaceID to hold no further
	// "@". A bare itemID (no "@") is the common case.
	docID := v.ItemID
	if i := strings.Index(v.ItemID, SpaceItemIDSeparator); i >= 0 {
		docID = v.ItemID[:i]
		spaceID := v.ItemID[i+1:]
		if docID == "" {
			return validation.NewErrBadRecordFieldValue("itemID", "empty document id before '@' separator")
		}
		if spaceID == "" {
			return validation.NewErrBadRecordFieldValue("itemID", "empty spaceID after '@' separator")
		}
		if strings.Contains(spaceID, SpaceItemIDSeparator) {
			return validation.NewErrBadRecordFieldValue("itemID", "must not contain more than one '@' separator")
		}
	}
	if err := validate.RecordID(docID); err != nil {
		return validation.NewErrBadRecordFieldValue("itemID", err.Error())
	}
	return nil
}
