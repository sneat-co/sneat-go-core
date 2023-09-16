package facade4userus

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/modules/contactus/briefs4contactus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/facade4teamus"
	"github.com/sneat-co/sneat-go-core/modules/userus/dto4userus"
	"github.com/sneat-co/sneat-go-core/modules/userus/models4userus"
	"time"
)

// InitUserRecord sets user title
func InitUserRecord(ctx context.Context, userContext facade.User, request dto4userus.InitUserRecordRequest) (user models4userus.UserContext, err error) {
	if err = request.Validate(); err != nil {
		err = fmt.Errorf("%w: %v", facade.ErrBadRequest, err.Error())
		return
	}
	userID := userContext.GetID()
	err = runReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		user, err = initUserRecordTxWorker(ctx, tx, userID, request)
		return err
	})
	if err != nil {
		user.Dto = nil
		return user, fmt.Errorf("failet to init user record: %w", err)
	}
	if request.Team != nil {
		var hasTeamOfSameType bool
		for _, team := range user.Dto.Teams {
			if team.Type == request.Team.Type {
				hasTeamOfSameType = true
				break
			}
		}
		if !hasTeamOfSameType && request.Team != nil {
			if _, err = facade4teamus.CreateTeam(ctx, userContext, *request.Team); err != nil {
				err = fmt.Errorf("failed to create team for user: %w", err)
				return
			}
		}
	}

	return
}

func initUserRecordTxWorker(ctx context.Context, tx dal.ReadwriteTransaction, userID string, request dto4userus.InitUserRecordRequest) (user models4userus.UserContext, err error) {
	var isNewUser bool
	user = models4userus.NewUserContext(userID)
	if err = TxGetUserByID(ctx, tx, user.Record); err != nil {
		if dal.IsNotFound(err) {
			isNewUser = true
		} else {
			return
		}
	}
	if isNewUser {
		if err = createUserRecord(ctx, tx, request, user); err != nil {
			err = fmt.Errorf("faield to create user record: %w", err)
			return
		}
	} else if err = updateUserRecordWithInitData(ctx, tx, request, user); err != nil {
		err = fmt.Errorf("faield to update user record: %w", err)
		return
	}
	return
}

func createUserRecord(ctx context.Context, tx dal.ReadwriteTransaction, request dto4userus.InitUserRecordRequest, user models4userus.UserContext) error {
	user.Dto.Status = "active"
	user.Dto.Type = briefs4contactus.ContactTypePerson
	user.Dto.CountryID = "--"
	user.Dto.AgeGroup = "unknown"
	user.Dto.Gender = "unknown"
	if request.Name != nil {
		user.Dto.Name = request.Name
	}
	user.Dto.Created.At = time.Now()
	user.Dto.Created.Client = request.RemoteClient
	user.Dto.Email = request.Email
	user.Dto.EmailVerified = request.EmailIsVerified
	user.Dto.Emails = []dbmodels.PersonEmail{
		{
			Type:         "primary",
			Address:      request.Email,
			Verified:     request.EmailIsVerified,
			AuthProvider: request.AuthProvider,
		},
	}
	if request.IanaTimezone != "" {
		user.Dto.Timezone = &dbmodels.Timezone{
			Iana: request.IanaTimezone,
		}
	}
	if user.Dto.Title == "" && user.Dto.Name.IsEmpty() {
		user.Dto.Title = user.Dto.Email
	}
	if err := user.Dto.Validate(); err != nil {
		return fmt.Errorf("user record prepared for insert is not valid: %w", err)
	}
	if err := tx.Insert(ctx, user.Record); err != nil {
		return fmt.Errorf("failed to insert user record: %w", err)
	}
	return nil
}

func updateUserRecordWithInitData(ctx context.Context, tx dal.ReadwriteTransaction, request dto4userus.InitUserRecordRequest, user models4userus.UserContext) error {
	var updates []dal.Update
	if name := request.Name; name != nil {
		if name.Full == "" && !name.IsEmpty() {
			name.Full = name.Title()
		}
		if !name.IsEmpty() {
			updates = append(updates, dal.Update{Field: "name", Value: name})
		}
		user.Dto.Name = name
	}

	if request.IanaTimezone != "" && (user.Dto.Timezone == nil || user.Dto.Timezone.Iana == "") {
		if user.Dto.Timezone == nil {
			user.Dto.Timezone = &dbmodels.Timezone{}
		}
		user.Dto.Timezone.Iana = request.IanaTimezone
		updates = append(updates, dal.Update{Field: "timezone.iana", Value: request.IanaTimezone})
	}
	if user.Dto.Title == user.Dto.Email && user.Dto.Name != nil && !user.Dto.Name.IsEmpty() {
		user.Dto.Title = ""
		updates = append(updates, dal.Update{Field: "title", Value: dal.DeleteField})
	}
	if len(updates) > 0 {
		if err := user.Dto.Validate(); err != nil {
			return fmt.Errorf("user record prepared for update is not valid: %w", err)
		}
		if err := tx.Update(ctx, user.Key, updates); err != nil {
			return err
		}
	}
	return nil
}
