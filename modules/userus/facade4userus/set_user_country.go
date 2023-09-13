package facade4userus

import (
	"context"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/strongo/validation"
)

type SetUserCountryRequest struct {
	CountryID string `json:"countryID"`
}

func (v SetUserCountryRequest) Validate() error {
	if v.CountryID == "" {
		return validation.NewErrRequestIsMissingRequiredField("countryID")
	}
	if len(v.CountryID) != 2 {
		return validation.NewErrBadRequestFieldValue("countryID", "must be 2 characters long")
	}
	return nil
}

func SetUserCountry(ctx context.Context, userContext facade.User, request SetUserCountryRequest) (err error) {
	return RunUserWorker(ctx, userContext, func(ctx context.Context, tx dal.ReadwriteTransaction, userWorkerParams *UserWorkerParams) error {
		if userWorkerParams.User.Data.CountryID != request.CountryID {
			userWorkerParams.User.Data.CountryID = request.CountryID
			userWorkerParams.UserUpdates = append(userWorkerParams.UserUpdates,
				dal.Update{Field: "countryID", Value: request.CountryID})
		}
		return nil
	})
}
