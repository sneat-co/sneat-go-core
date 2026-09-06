package entityreferences

import (
	"errors"
	"testing"
)

func TestValidationRequestAndMissingError(t *testing.T) {
	if err := (ValidationRequest{SpaceID: "space1", IDs: []string{"a", "b"}}).Validate(); err != nil {
		t.Fatal(err)
	}
	for _, request := range []ValidationRequest{{}, {SpaceID: "space1", IDs: []string{"a", "a"}}, {SpaceID: "space1", IDs: []string{" bad"}}, {SpaceID: "space1", IDs: []string{"a/b"}}} {
		if request.Validate() == nil {
			t.Fatalf("accepted %+v", request)
		}
	}
	err := MissingError{IDs: []string{"a"}}
	if !errors.Is(err, ErrReferenceNotFound) {
		t.Fatalf("err=%v", err)
	}
}
