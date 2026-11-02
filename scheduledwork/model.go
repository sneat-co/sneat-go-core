// Package scheduledwork defines Space-owned background work. It reuses the host
// delayed executor; it does not authorize or implement a product's side effects.
package scheduledwork

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/dal-go/record"
	"github.com/sneat-co/sneat-go-core/coretypes"
)

const Collection = "scheduledWork"
const MaxGeneration int64 = 9007199254740991

var ErrInvalid = errors.New("invalid scheduled work")

type State string

const (
	Pending     State = "pending"
	Leased      State = "leased"
	Completed   State = "completed"
	Cancelled   State = "cancelled"
	Quarantined State = "quarantined"
)

// Ref identifies one logical action. A new business entity (e.g. a replacement
// Debtus reminder) has a new SubjectID, preserving its own business history.
type Ref struct {
	SpaceID   coretypes.SpaceID `json:"spaceID" firestore:"spaceID"`
	Producer  coretypes.ExtID   `json:"producer" firestore:"producer"`
	SubjectID string            `json:"subjectID" firestore:"subjectID"`
	Action    string            `json:"action" firestore:"action"`
}

func identifier(value string) bool {
	if len(value) == 0 || len(value) > 128 {
		return false
	}
	for _, c := range value {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '-' || c == '_' || c == '.':
		default:
			return false
		}
	}
	return true
}

func (r Ref) Validate() error {
	if coretypes.ValidateSpaceID(r.SpaceID) != nil {
		return fmt.Errorf("%w: owning Space", ErrInvalid)
	}
	if !identifier(string(r.Producer)) || !identifier(r.Action) || len(r.SubjectID) == 0 || len(r.SubjectID) > 256 || strings.TrimSpace(r.SubjectID) != r.SubjectID || strings.Contains(r.SubjectID, "/") || strings.IndexFunc(r.SubjectID, unicode.IsControl) >= 0 || record.ValidateStringID(r.SubjectID) != nil {
		return fmt.Errorf("%w: logical identity", ErrInvalid)
	}
	return nil
}

// Key cannot use Core's global fallback: Validate rejects an absent Space.
// Length framing separates identities without trusting delimiters in IDs.
func (r Ref) Key() (*record.Key, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	var framed strings.Builder
	for _, value := range []string{string(r.SpaceID), string(r.Producer), r.SubjectID, r.Action} {
		framed.WriteString(strconv.Itoa(len(value)))
		framed.WriteByte(':')
		framed.WriteString(value)
	}
	digest := sha256.Sum256([]byte(framed.String()))
	return record.NewKeyWithParentAndID(coretypes.NewSpaceModuleKey(r.SpaceID, "core"), Collection, "work_"+hex.EncodeToString(digest[:])), nil
}

type Spec struct {
	Ref     Ref    `json:"ref" firestore:"ref"`
	Handler string `json:"handler" firestore:"handler"`
	// SubjectRevision is an opaque owning-domain revision, not generic consent.
	SubjectRevision string    `json:"subjectRevision" firestore:"subjectRevision"`
	RunAt           time.Time `json:"runAt" firestore:"runAt"`
}

func instant(t time.Time) bool { return !t.IsZero() && t.Nanosecond()%1000 == 0 }

func (s Spec) Validate() error {
	if err := s.Ref.Validate(); err != nil {
		return err
	}
	if !identifier(s.Handler) || len(s.SubjectRevision) == 0 || len(s.SubjectRevision) > 128 || !instant(s.RunAt) {
		return fmt.Errorf("%w: handler, revision or execution instant", ErrInvalid)
	}
	return nil
}

type Claim struct {
	Token string    `json:"token" firestore:"token"`
	At    time.Time `json:"at" firestore:"at"`
	Until time.Time `json:"until" firestore:"until"`
}

// Work contains execution metadata only. It is server-private. Quarantine is
// unresolved work, never evidence of successful business completion.
type Work struct {
	Spec          Spec      `json:"spec" firestore:"spec"`
	Generation    int64     `json:"generation" firestore:"generation"`
	State         State     `json:"state" firestore:"state"`
	NextAttemptAt time.Time `json:"nextAttemptAt" firestore:"nextAttemptAt"`
	Attempts      int64     `json:"attempts" firestore:"attempts"`
	Claim         *Claim    `json:"claim,omitempty" firestore:"claim,omitempty"`
	// EffectKey persists the owning idempotency identity after external intent
	// commits. Rescheduling/cancellation cannot silently erase that uncertainty.
	EffectKey   string     `json:"effectKey,omitempty" firestore:"effectKey,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty" firestore:"completedAt,omitempty"`
	FailureCode string     `json:"failureCode,omitempty" firestore:"failureCode,omitempty"`
}

func (w Work) Validate() error {
	if err := w.Spec.Validate(); err != nil {
		return err
	}
	if w.Generation < 1 || w.Generation > MaxGeneration || w.Attempts < 0 || w.Attempts > MaxGeneration || !instant(w.NextAttemptAt) || w.NextAttemptAt.Before(w.Spec.RunAt) || len(w.EffectKey) > 256 || len(w.FailureCode) > 128 {
		return fmt.Errorf("%w: execution metadata", ErrInvalid)
	}
	switch w.State {
	case Leased:
		if w.Claim == nil || !identifier(w.Claim.Token) || !instant(w.Claim.At) || !instant(w.Claim.Until) || !w.Claim.Until.After(w.Claim.At) || w.Claim.At.Before(w.NextAttemptAt) || w.Attempts < 1 {
			return fmt.Errorf("%w: execution claim", ErrInvalid)
		}
	case Pending, Completed, Cancelled, Quarantined:
		if w.Claim != nil {
			return fmt.Errorf("%w: unexpected execution claim", ErrInvalid)
		}
	default:
		return fmt.Errorf("%w: unsupported execution state", ErrInvalid)
	}
	if w.State == Completed {
		if w.Attempts < 1 || w.CompletedAt == nil || !instant(*w.CompletedAt) || w.CompletedAt.Before(w.Spec.RunAt) {
			return fmt.Errorf("%w: completion receipt", ErrInvalid)
		}
	} else if w.CompletedAt != nil {
		return fmt.Errorf("%w: unexpected completion receipt", ErrInvalid)
	}
	if w.State == Cancelled && w.EffectKey != "" {
		return fmt.Errorf("%w: external intent requires reconciliation", ErrInvalid)
	}
	if w.State == Quarantined && w.FailureCode == "" {
		return fmt.Errorf("%w: quarantine requires a recovery reason", ErrInvalid)
	}
	return nil
}
