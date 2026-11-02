package scheduledwork

import (
	"context"

	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/strongo/delaying"
)

// Executor is an adapter to the EXISTING shared delayed-task runtime. It does
// not claim work, start transactions, choose business policy or mark jobs done.
// The owning run function must reload current business/handler/revision and use
// PrepareExecution in the SAME transaction as its business effect or intent.
// Provider IO must remain outside transaction callbacks.
type Executor struct {
	producer coretypes.ExtID
	handler  string
	queue    string
	delayer  delaying.Delayer
}

// RegisterExecutor runs once during backend registration. Producer and handler
// are server-owned configuration, not client-selectable values. Payloads contain
// only the Space-owned logical reference, generation and opaque claim token.
func RegisterExecutor(register func(string, any) delaying.Delayer, producer coretypes.ExtID, handler, queue string, run func(context.Context, ClaimRef) error) (*Executor, error) {
	if register == nil || run == nil || !identifier(string(producer)) || !identifier(handler) || !identifier(queue) {
		return nil, ErrInvalid
	}
	e := &Executor{producer: producer, handler: handler, queue: queue}
	worker := func(ctx context.Context, space, subject, action string, generation int64, token string) error {
		claim := ClaimRef{Ref: Ref{SpaceID: coretypes.SpaceID(space), Producer: producer, SubjectID: subject, Action: action}, Generation: generation, Token: token}
		if err := e.validate(claim); err != nil {
			return err
		}
		// Returning run's error unchanged lets the existing authenticated task host
		// apply its normal delivery retry. The owning handler decides stale no-op.
		return run(ctx, claim)
	}
	e.delayer = register(handler, worker)
	if e.delayer == nil {
		return nil, ErrInvalid
	}
	return e, nil
}

func (e *Executor) validate(claim ClaimRef) error {
	if e == nil || claim.Ref.Producer != e.producer || claim.Generation < 1 || claim.Generation > MaxGeneration || !identifier(claim.Token) {
		return ErrInvalid
	}
	return claim.Ref.Validate()
}

// Enqueue runs only AFTER the claim transaction commits. An enqueue failure or
// unknown reply leaves durable leased work for indexed expiry recovery; neither
// enqueue success nor failure is evidence of business completion. Correctness
// does not depend on the queue accepting an arbitrary future scheduling horizon.
func (e *Executor) Enqueue(ctx context.Context, claim ClaimRef) error {
	if err := e.validate(claim); err != nil {
		return err
	}
	if e.delayer == nil {
		return ErrInvalid
	}
	return e.delayer.EnqueueWork(ctx, delaying.With(e.queue, e.handler, 0), string(claim.Ref.SpaceID), claim.Ref.SubjectID, claim.Ref.Action, claim.Generation, claim.Token)
}
