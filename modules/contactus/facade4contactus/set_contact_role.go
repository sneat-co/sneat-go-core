package facade4contactus

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dal4contactus"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dto4contactus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dal4teamus"
	"github.com/strongo/slice"
)

// SetContactRoles sets contacts address
func SetContactRoles(ctx context.Context, user facade.User, request dto4contactus.SetContactRolesRequest) (err error) {
	if err = request.Validate(); err != nil {
		return
	}
	err = dal4teamus.RunTeamWorker(ctx, user, request.TeamRequest,
		func(ctx context.Context, tx dal.ReadwriteTransaction, params *dal4teamus.TeamWorkerParams) (err error) {
			return setContactRolesTxWorker(ctx, tx, params, request)
		},
	)
	if err != nil {
		return fmt.Errorf("failed to set contact status: %w", err)
	}
	return nil
}

func setContactRolesTxWorker(
	ctx context.Context, tx dal.ReadwriteTransaction, params *dal4teamus.TeamWorkerParams,
	request dto4contactus.SetContactRolesRequest,
) (err error) {
	contact := dal4contactus.NewContactContext(params.Team.ID, request.ContactID)
	if err = tx.Get(ctx, contact.Record); err != nil {
		return fmt.Errorf("failed to get contact record: %w", err)
	}

	if err := contact.Data.Validate(); err != nil {
		return fmt.Errorf("contact DTO is not valid after loading from DB: %w", err)
	}
	for _, role := range request.Remove {
		contact.Data.Roles = slice.RemoveInPlace(role, contact.Data.Roles)
	}
	contact.Data.Roles = append(contact.Data.Roles, request.Add...)
	if err := contact.Data.Validate(); err != nil {
		return fmt.Errorf("contact DTO is not valid after setting address: %w", err)
	}
	if err := tx.Update(ctx, contact.Key, []dal.Update{{Field: "address", Value: contact.Data.Address}}); err != nil {
		return err
	}
	return nil
}
