package entityreferences

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/coretypes"
)

const MaxReferences = 64

var ErrReferenceNotFound = errors.New("Space entity reference not found")

// ValidationRequest names entities owned by one extension in one Space. The
// concrete owning extension resolves existence using the caller's DALgo read
// session, allowing validation and a dependent write to share one transaction.
type ValidationRequest struct {
	SpaceID coretypes.SpaceID
	IDs     []string
}

func (r ValidationRequest) Validate() error {
	if coretypes.ValidateSpaceID(r.SpaceID) != nil || len(r.IDs) > MaxReferences {
		return fmt.Errorf("invalid Space entity reference request")
	}
	seen := make(map[string]struct{}, len(r.IDs))
	for _, id := range r.IDs {
		if id == "" || len(id) > 100 || strings.TrimSpace(id) != id || strings.ContainsAny(id, "/\r\n") {
			return fmt.Errorf("invalid Space entity reference ID")
		}
		if _, ok := seen[id]; ok {
			return fmt.Errorf("duplicate Space entity reference ID")
		}
		seen[id] = struct{}{}
	}
	return nil
}

type Validator interface {
	ValidateExisting(ctx context.Context, reader dal.ReadSession, request ValidationRequest) error
}

type MissingError struct{ IDs []string }

func (e MissingError) Error() string {
	return fmt.Sprintf("%v: %s", ErrReferenceNotFound, strings.Join(e.IDs, ","))
}
func (e MissingError) Unwrap() error { return ErrReferenceNotFound }
