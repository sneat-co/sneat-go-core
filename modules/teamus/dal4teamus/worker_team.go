package dal4teamus

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dto4teamus"
	"log"
	"reflect"
	"time"
)

type teamWorker = func(ctx context.Context, tx dal.ReadwriteTransaction, teamWorkerParams *TeamWorkerParams) (err error)

func NewTeamWorkerParams(userID, teamID string) TeamWorkerParams {
	return TeamWorkerParams{
		UserID:  userID,
		Team:    NewTeamContext(teamID),
		Started: time.Now(),
	}
}

// TeamWorkerParams passes data to a team worker
type TeamWorkerParams struct {
	UserID  string
	Started time.Time
	//
	Team        TeamContext
	TeamUpdates []dal.Update
}

// ModuleTeamWorkerParams passes data to a team worker
type ModuleTeamWorkerParams[D TeamModuleData] struct {
	TeamWorkerParams
	TeamModuleEntry   record.DataWithID[string, D]
	TeamModuleUpdates []dal.Update
}

type TeamModuleData interface {
	Validate() error
}

func RunModuleTeamWorker[D TeamModuleData](
	ctx context.Context,
	user facade.User,
	request dto4teamus.TeamRequest,
	moduleID string,
	worker func(ctx context.Context, tx dal.ReadwriteTransaction, teamWorkerParams *ModuleTeamWorkerParams[D]) (err error),
) (err error) {
	var d D
	teamWorkerParams := NewTeamWorkerParams(user.GetID(), request.TeamID)
	params := ModuleTeamWorkerParams[D]{
		TeamWorkerParams: teamWorkerParams,
		TeamModuleEntry: record.NewDataWithID("",
			dal.NewKeyWithParentAndID(teamWorkerParams.Team.Key, Collection, moduleID),
			reflect.New(reflect.TypeOf(d)).Elem().Interface().(D),
		),
	}

	db := facade.GetDatabase(ctx)
	return db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) (err error) {
		if err := tx.GetMulti(ctx, []dal.Record{params.Team.Record, params.TeamModuleEntry.Record}); err != nil && !dal.IsNotFound(err) {
			return fmt.Errorf("failed to get team & team module records: %w", err)
		}
		if err = worker(ctx, tx, &params); err != nil {
			return fmt.Errorf("failed to execute module team worker: %w", err)
		}

		return nil
	})
}

// RunTeamWorker executes a team worker
var RunTeamWorker = func(ctx context.Context, user facade.User, request dto4teamus.TeamRequest, worker teamWorker) (err error) {
	if user == nil {
		panic("user is nil")
	}
	if err := request.Validate(); err != nil {
		return fmt.Errorf("team request is not valid: %w", err)
	}
	userID := user.GetID()
	if userID == "" {
		err = facade.ErrUnauthorized
		return
	}
	db := facade.GetDatabase(ctx)
	return db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) (err error) {
		params := NewTeamWorkerParams(userID, request.TeamID)
		if err = tx.Get(ctx, params.Team.Record); err != nil {
			return fmt.Errorf("failed to load team record: %w", err)
		}
		if err = params.Team.Data.Validate(); err != nil {
			log.Printf("WARNING: team record loaded from DB is not valid: %v: dto=%+v", err, params.Team.Data)
		}
		if err = worker(ctx, tx, &params); err != nil {
			return fmt.Errorf("failed to execute team worker: %w", err)
		}
		if len(params.TeamUpdates) > 0 {
			if err = TxUpdateTeam(ctx, tx, params.Started, params.Team, params.TeamUpdates); err != nil {
				return fmt.Errorf("failed to update team record: %w", err)
			}
		}
		return err
	})
}

// CreateTeamItem creates a team item
func CreateTeamItem[D TeamModuleData](
	ctx context.Context,
	user facade.User,
	counter string,
	teamRequest dto4teamus.TeamRequest,
	moduleID string,
	worker func(
		ctx context.Context,
		tx dal.ReadwriteTransaction,
		teamWorkerParams *ModuleTeamWorkerParams[D],
	) (err error),
) (err error) {
	if worker == nil {
		panic("worker is nil")
	}
	if err := teamRequest.Validate(); err != nil {
		return err
	}
	err = RunModuleTeamWorker(ctx, user, teamRequest, moduleID,
		func(ctx context.Context, tx dal.ReadwriteTransaction, params *ModuleTeamWorkerParams[D]) (err error) {
			if err := worker(ctx, tx, params); err != nil {
				return fmt.Errorf("failed to execute team worker passed to CreateTeamItem: %w", err)
			}
			if counter != "" {
				if err = incrementCounter(&params.TeamWorkerParams, counter); err != nil {
					return fmt.Errorf("failed to incement teams counter=%v: %w", counter, err)
				}
			}
			if err = params.Team.Data.Validate(); err != nil {
				return fmt.Errorf("team record is not valid after performing worker: %w", err)
			}
			return
		})
	if err != nil {
		return fmt.Errorf("failed to create a team item: %w", err)
	}
	return nil
}
