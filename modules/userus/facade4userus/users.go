package facade4userus

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/modules/userus/models4userus"
	"log"
	"time"
)

// Facade interface
type Facade interface {
	GetByID(ctx context.Context, user dal.Record) error
}

// DbFacade is a facade interface
type DbFacade struct {
	//db dal.DB
}

//func GetByID(ctx context.Context, id dal.MeetingRecord, get func(context.Context, dal.MeetingRecord) error) error {
//	return get()
//}

// TxFacade is a facade interface
//type TxFacade struct {
//	tx dal.ReadwriteTransaction
//}

// GetUserByID load user record by ContactID
var GetUserByID = func(ctx context.Context, getter dal.ReadSession, user dal.Record) (err error) {
	if err = getter.Get(ctx, user); err != nil {
		return fmt.Errorf("failed to get user record by user=%v: %w", user.Key().ID, err)
	}
	return nil
}

// TxGetUserByID load user record by ContactID within transaction
var TxGetUserByID = func(ctx context.Context, transaction dal.ReadwriteTransaction, user dal.Record) (
	err error,
) {
	return transaction.Get(ctx, user)
}

// TxUpdateUser update user record
var TxUpdateUser = func(
	ctx context.Context,
	transaction dal.ReadwriteTransaction,
	timestamp time.Time,
	userKey *dal.Key,
	data []dal.Update,
	opts ...dal.Precondition,
) error {
	if transaction == nil {
		panic("transaction == nil")
	}
	if userKey == nil {
		panic("userKey == nil")
	}
	data = append(data,
		dal.Update{Field: "timestamp", Value: timestamp},
	)
	return transaction.Update(ctx, userKey, data, opts...)
}

// TxGetUsers load user records
var TxGetUsers = func(ctx context.Context, tx dal.ReadwriteTransaction, users []dal.Record) (err error) {
	if users == nil {
		panic("api4meetingus == nil")
	}
	if len(users) == 0 {
		return nil
	}
	return tx.GetMulti(ctx, users)
}

// UserWorkerParams passes data to a team worker
type UserWorkerParams struct {
	Started     time.Time
	User        models4userus.User
	UserUpdates []dal.Update
}

type userWorker = func(ctx context.Context, tx dal.ReadwriteTransaction, userWorkerParams *UserWorkerParams) (err error)

var RunUserWorker = func(ctx context.Context, user facade.User, worker userWorker) (err error) {
	if user == nil {
		panic("user == nil")
	}
	params := UserWorkerParams{
		User:    models4userus.NewUser(user.GetID()),
		Started: time.Now(),
	}
	db := facade.GetDatabase(ctx)
	return db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) (err error) {
		if err = tx.Get(ctx, params.User.Record); err != nil {
			return fmt.Errorf("failed to load user record: %w", err)
		}
		if err = params.User.Data.Validate(); err != nil {
			log.Printf("WARNING: user record loaded from DB is not valid: %v: data=%+v", err, params.User.Data)
		}
		if err = worker(ctx, tx, &params); err != nil {
			return fmt.Errorf("failed to execute teamWorker: %w", err)
		}
		if err = params.User.Data.Validate(); err != nil {
			return fmt.Errorf("user record is not valid after update: %w", err)
		}
		if len(params.UserUpdates) > 0 {
			if err = TxUpdateUser(ctx, tx, params.Started, params.User.Key, params.UserUpdates); err != nil {
				return fmt.Errorf("failed to update team record: %w", err)
			}
		}
		return err
	})
}
