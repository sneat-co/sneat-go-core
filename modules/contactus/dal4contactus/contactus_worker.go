package dal4contactus

// TODO: Should it be in DAL?

import (
	"context"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dal4teamus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dto4teamus"
)

type ContactusTeamWorkerParams struct {
	dal4teamus.TeamWorkerParams
	ContactusTeam        ContactusTeamContext
	ContactusTeamUpdates []dal.Update
}

func NewContactusTeamWorkerParams(userID, teamID string) *ContactusTeamWorkerParams {
	return &ContactusTeamWorkerParams{
		TeamWorkerParams: dal4teamus.NewTeamWorkerParams(userID, teamID),
		ContactusTeam:    NewContactusTeamContext(teamID),
	}
}

func RunReadonlyContactusTeamWorker(
	ctx context.Context,
	user facade.User,
	request dto4teamus.TeamRequest,
	worker func(ctx context.Context, tx dal.ReadTransaction, params *ContactusTeamWorkerParams) (err error),
) error {
	params := NewContactusTeamWorkerParams(user.GetID(), request.TeamID)
	db := facade.GetDatabase(ctx)
	return db.RunReadonlyTransaction(ctx, func(ctx context.Context, tx dal.ReadTransaction) (err error) {
		return worker(ctx, tx, params)
	})
}

func RunContactusTeamWorker(
	ctx context.Context,
	user facade.User,
	request dto4teamus.TeamRequest,
	worker func(ctx context.Context, tx dal.ReadwriteTransaction, params *ContactusTeamWorkerParams) (err error),
) error {
	params := NewContactusTeamWorkerParams(user.GetID(), request.TeamID)
	db := facade.GetDatabase(ctx)
	return db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) (err error) {
		if err = worker(ctx, tx, params); err != nil {
			return err
		}
		if len(params.ContactusTeamUpdates) > 0 {
			if err = tx.Update(ctx, params.ContactusTeam.Key, params.ContactusTeamUpdates); err != nil {
				return err
			}
		}
		return nil
	})
}
