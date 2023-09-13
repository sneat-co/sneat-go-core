package facade4userus

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/modules/userus/dto4userus"
	"github.com/sneat-co/sneat-go-core/modules/userus/models4userus"
	"strings"
	"time"
)

// CreateUser creates user record in DB
func CreateUser(ctx context.Context, userID string, request dto4userus.CreateUserRequestWithRemoteClientInfo) error {
	db := facade.GetDatabase(ctx)
	user := models4userus.NewUserContext(userID)
	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		if err := TxGetUserByID(ctx, tx, user.Record); !dal.IsNotFound(err) {
			return err // Might be nil or not related to "record not found"
		}
		user.Dto.Created.At = time.Now()
		if request.Creator != "" {
			request.RemoteClient.HostOrApp = request.Creator
		}
		user.Dto.Created.Client = request.RemoteClient
		{ // Set user's names
			user.Dto.Name.Full = models4userus.CleanTitle(request.Title)
			if strings.Contains(user.Dto.Name.Full, " ") {
				user.Dto.Defaults = &models4userus.UserDefaults{
					ShortNames: models4userus.GetShortNames(user.Dto.Name.Full),
				}
			}
		}
		user.Dto.Email = strings.TrimSpace(request.Email)
		user.Dto.Emails = []dbmodels.PersonEmail{
			{Type: "primary", Address: user.Dto.Email},
		}
		if user.Dto.Gender == "" {
			user.Dto.Gender = "unknown"
		}
		if user.Dto.AgeGroup == "" {
			user.Dto.AgeGroup = "unknown"
		}
		if err := user.Dto.Validate(); err != nil {
			return fmt.Errorf("not able to create user record: %w", err)
		}
		if err := tx.Insert(ctx, user.Record); err != nil {
			return fmt.Errorf("failed to create user record: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create user record in database: %w", err)
	}
	return nil
}
