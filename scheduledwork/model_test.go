package scheduledwork

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/sneat-co/sneat-go-core/coretypes"
)

func sample() Work {
	now := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
	return Work{Spec: Spec{Ref: Ref{SpaceID: "personal", Producer: "commitius", SubjectID: "cmt_goal", Action: "deadline"}, Handler: "commitius.deadline.v1", SubjectRevision: "1", RunAt: now}, Generation: 1, State: Pending, NextAttemptAt: now}
}

func TestStableSpaceOwnedIdentity(t *testing.T) {
	r := sample().Spec.Ref
	key, err := r.Key()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(key.String(), "spaces/personal/ext/core/scheduledWork/work_") {
		t.Fatal("not Space owned", key.String())
	}
	again, _ := r.Key()
	if key.String() != again.String() {
		t.Fatal("unstable identity")
	}
	for _, change := range []func(*Ref){func(r *Ref) { r.SpaceID = "other" }, func(r *Ref) { r.Producer = "calendarius" }, func(r *Ref) { r.SubjectID = "new-reminder" }, func(r *Ref) { r.Action = "vote" }} {
		other := r
		change(&other)
		k, err := other.Key()
		if err != nil || k.String() == key.String() {
			t.Fatal("distinct work aliased", k, err)
		}
	}
	for _, r := range []Ref{{}, {SpaceID: "bad/space", Producer: "commitius", SubjectID: "c", Action: "deadline"}, {SpaceID: "personal", Producer: "", SubjectID: "c", Action: "deadline"}, {SpaceID: "personal", Producer: "commitius", SubjectID: "", Action: "deadline"}, {SpaceID: "personal", Producer: "commitius", SubjectID: "../c", Action: "deadline"}, {SpaceID: "personal", Producer: "commitius", SubjectID: " c ", Action: "deadline"}, {SpaceID: "personal", Producer: "commitius", SubjectID: strings.Repeat("x", 257), Action: "deadline"}, {SpaceID: "personal", Producer: "commitius", SubjectID: "c", Action: "bad/action"}} {
		if key, err := r.Key(); !errors.Is(err, ErrInvalid) || key != nil {
			t.Fatal("invalid identity acquired a key", key, err)
		}
	}
}

func TestWorkMetadataAndLifecycleValidation(t *testing.T) {
	if err := sample().Validate(); err != nil {
		t.Fatal(err)
	}
	for name, change := range map[string]func(*Work){
		"space":                  func(w *Work) { w.Spec.Ref.SpaceID = coretypes.SpaceID("") },
		"handler-empty":          func(w *Work) { w.Spec.Handler = "" },
		"handler-long":           func(w *Work) { w.Spec.Handler = strings.Repeat("x", 129) },
		"handler-invalid":        func(w *Work) { w.Spec.Handler = "invalid/handler" },
		"revision":               func(w *Work) { w.Spec.SubjectRevision = "" },
		"due-zero":               func(w *Work) { w.Spec.RunAt = time.Time{} },
		"precision":              func(w *Work) { w.Spec.RunAt = w.Spec.RunAt.Add(time.Nanosecond) },
		"generation":             func(w *Work) { w.Generation = 0 },
		"generation-overflow":    func(w *Work) { w.Generation = MaxGeneration + 1 },
		"attempt-negative":       func(w *Work) { w.Attempts = -1 },
		"attempt-time":           func(w *Work) { w.NextAttemptAt = w.Spec.RunAt.Add(-time.Second) },
		"effect-bound":           func(w *Work) { w.EffectKey = strings.Repeat("x", 257) },
		"error-bound":            func(w *Work) { w.FailureCode = strings.Repeat("x", 129) },
		"unknown-state":          func(w *Work) { w.State = "future" },
		"unearned-completion":    func(w *Work) { w.State = Completed },
		"unexpected-receipt":     func(w *Work) { w.CompletedAt = &w.Spec.RunAt },
		"pending-lease":          func(w *Work) { w.Claim = &Claim{} },
		"missing-lease":          func(w *Work) { w.State = Leased },
		"cancel-dispatched":      func(w *Work) { w.State = Cancelled; w.EffectKey = "already-dispatched" },
		"unexplained-quarantine": func(w *Work) { w.State = Quarantined },
	} {
		t.Run(name, func(t *testing.T) {
			w := sample()
			change(&w)
			if !errors.Is(w.Validate(), ErrInvalid) {
				t.Fatal("invalid work accepted", w)
			}
		})
	}
	for _, state := range []State{Cancelled, Quarantined, Completed, Leased} {
		w := sample()
		w.State = state
		switch state {
		case Quarantined:
			w.FailureCode = "provider-uncertain"
			w.EffectKey = "same-effect"
		case Completed:
			w.Attempts = 1
			w.CompletedAt = &w.Spec.RunAt
		case Leased:
			w.Attempts = 1
			w.Claim = &Claim{Token: "opaque-token", At: w.NextAttemptAt, Until: w.NextAttemptAt.Add(time.Minute)}
		}
		if err := w.Validate(); err != nil {
			t.Fatal(state, err)
		}
	}
}

func TestClaimValidationAndCompletionTime(t *testing.T) {
	for _, change := range []func(*Work){func(w *Work) { w.Claim.Token = "bad/token" }, func(w *Work) { w.Claim.Until = w.Claim.At }, func(w *Work) { w.Claim.At = w.Spec.RunAt.Add(-time.Second) }, func(w *Work) { w.Attempts = 0 }} {
		w := sample()
		w.State = Leased
		w.Attempts = 1
		w.Claim = &Claim{Token: "token", At: w.Spec.RunAt, Until: w.Spec.RunAt.Add(time.Minute)}
		change(&w)
		if !errors.Is(w.Validate(), ErrInvalid) {
			t.Fatal("invalid claim passed")
		}
	}
	w := sample()
	w.State = Completed
	w.Attempts = 1
	early := w.Spec.RunAt.Add(-time.Second)
	w.CompletedAt = &early
	if !errors.Is(w.Validate(), ErrInvalid) {
		t.Fatal("early completion passed")
	}
}
