package entityreferences

import (
	"errors"
	"strings"
	"testing"

	"github.com/sneat-co/sneat-go-core/coretypes"
)

func TestValidationRequestAndMissingError(t *testing.T) {
	if err := (ValidationRequest{SpaceID: "space1", IDs: []string{"a", "b"}}).Validate(); err != nil {
		t.Fatal(err)
	}
	for _, request := range []ValidationRequest{{}, {SpaceID: "bad/space"}, {SpaceID: coretypes.SpaceID(strings.Repeat("a", 31))}, {SpaceID: "space1", IDs: []string{"a", "a"}}, {SpaceID: "space1", IDs: []string{" bad"}}, {SpaceID: "space1", IDs: []string{"a/b"}}} {
		if request.Validate() == nil {
			t.Fatalf("accepted %+v", request)
		}
	}
	if err := (ValidationRequest{SpaceID: coretypes.SpotSpaceID("abc123")}).Validate(); err != nil {
		t.Fatalf("reserved Spot Space ID rejected: %v", err)
	}
	err := MissingError{IDs: []string{"a"}}
	if !errors.Is(err, ErrReferenceNotFound) {
		t.Fatalf("err=%v", err)
	}
}
