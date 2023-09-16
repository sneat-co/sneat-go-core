package tests

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo2buntdb"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/modules/contactus/briefs4contactus"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dal4contactus"
	"github.com/sneat-co/sneat-go-core/modules/memberus/briefs4memberus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dal4teamus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/models4teamus"
	"github.com/sneat-co/sneat-go-core/modules/userus/models4userus"
	"testing"
)

// SetupOption options for test setup
type SetupOption struct {
	name string
	f    func(ctx context.Context, tx dal.ReadwriteTransaction) (err error)
}

// NewMockDB create a new in-memory mock database
func NewMockDB(t *testing.T, options ...SetupOption) dal.DB {
	t.Helper()
	db := dalgo2buntdb.NewInMemoryMockDB(t)
	ctx := context.Background()
	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		for _, setup := range options {
			if err := setup.f(ctx, tx); err != nil {
				t.Fatalf("failed to setup option %v: %v", setup.name, err)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to setup mock database: %v", err)
	}
	facade.GetDatabase = func(ctx context.Context) dal.DB {
		return db
	}
	return db
}

// WithProfile1 create 1st test profile
func WithProfile1() SetupOption {
	return SetupOption{
		name: "WithProfile1",
		f: func(ctx context.Context, tx dal.ReadwriteTransaction) (err error) {
			if err = createUsers(ctx, tx); err != nil {
				return err
			}
			if err = createTeams(ctx, tx); err != nil {
				return err
			}
			return err
		},
	}
}

func createUsers(ctx context.Context, tx dal.ReadwriteTransaction) (err error) {
	_, err = createUser(ctx, tx, "user1", &models4userus.UserDto{
		Email: "first.user@example.com",
		ContactBase: briefs4contactus.ContactBase{
			ContactBrief: briefs4contactus.ContactBrief{
				Name: &dbmodels.Name{
					Full: "First user",
				},
			},
		},
	})
	if err != nil {
		return err
	}
	_, err = createUser(ctx, tx, "user2", &models4userus.UserDto{
		ContactBase: briefs4contactus.ContactBase{
			ContactBrief: briefs4contactus.ContactBrief{
				Name: &dbmodels.Name{
					First: "Second",
					Last:  "UserDto",
				},
			},
		},
		Email: "second.user@example.com",
	})
	if err != nil {
		return err
	}
	return err
}

func createUser(ctx context.Context, tx dal.ReadwriteTransaction, id string, data *models4userus.UserDto) (record dal.Record, err error) {
	key := models4userus.NewUserKey(id)
	record = dal.NewRecordWithData(key, data)
	if err = tx.Set(ctx, record); err != nil {
		return nil, err
	}
	return
}

func createTeams(ctx context.Context, tx dal.ReadwriteTransaction) (err error) {
	team1 := dal4teamus.NewTeamContextWithDto("team1", &models4teamus.TeamDto{
		TeamBrief: models4teamus.TeamBrief{
			Type:  "team",
			Title: "First team",
			WithRequiredCountryID: dbmodels.WithRequiredCountryID{
				CountryID: "IE",
			},
		},
		NumberOf: map[string]int{
			"members": 1,
		},
	})
	team1Contactus := dal4contactus.NewContactusTeamContext(team1.ID)
	team1Contactus.Data.AddContact("m1", &briefs4contactus.ContactBrief{
		WithUserID: dbmodels.WithUserID{
			UserID: "user1",
		},
		Type:   briefs4contactus.ContactTypePerson,
		Gender: "unknown",
		Name: &dbmodels.Name{
			Full: "First user",
		},
		Title: "First user",
		WithRoles: dbmodels.WithRoles{
			Roles: []string{briefs4memberus.TeamMemberRoleTeamMember, briefs4memberus.TeamMemberRoleContributor},
		},
	})
	_, err = createTeam(ctx, tx, team1, team1Contactus)
	if err != nil {
		return err
	}
	return err
}

func createTeam(ctx context.Context, tx dal.ReadwriteTransaction, team dal4teamus.TeamContext, contactusTeam dal4contactus.ContactusTeamContext) (record dal.Record, err error) {
	for _, m := range contactusTeam.Data.Contacts {
		if m.UserID != "" {
			team.Data.AddUserID(m.UserID)
			contactusTeam.Data.AddUserID(m.UserID)
		}
	}
	if err := team.Data.Validate(); err != nil {
		return nil, fmt.Errorf("invalid team data: %w", err)
	}
	if err = tx.Set(ctx, team.Record); err != nil {
		return nil, err
	}

	if err := contactusTeam.Data.Validate(); err != nil {
		return nil, fmt.Errorf("invalid team data: %w", err)
	}
	if err = tx.Set(ctx, contactusTeam.Record); err != nil {
		return nil, err
	}

	//now := time.Now()
	//retro := &models4retrospectus.Retrospective{
	//	TimeLastAction: &now,
	//	Settings: models4retrospectus.RetrospectiveSettings{
	//		MaxVotesPerUser: 3,
	//	},
	//	Meeting: models4meetingus.Meeting{
	//		WithUserIDs: dbmodels.WithUserIDs{
	//			UserIDs: team.Data.UserIDs,
	//		},
	//	},
	//	Stage: models4retrospectus.StageReview,
	//	Items: []*models4retrospectus.RetroItem{
	//		{ID: "goods", Title: "Good stuff", Children: []*models4retrospectus.RetroItem{
	//			{ID: "g1", Title: "First item", Created: now},
	//			{ID: "g2", Title: "Second item", Created: now},
	//			{ID: "g3", Title: "Third item", Created: now},
	//		}},
	//	},
	//}
	//for contactID, m := range contactusTeam.Data.Contacts {
	//	if m.HasRole(briefs4memberus.TeamMemberRoleTeamMember) {
	//		contact := &models4meetingus.MeetingMemberBrief{
	//			ContactBrief: *m,
	//		}
	//		contact.AddRole(briefs4memberus.TeamMemberRoleContributor)
	//		retro.AddContact(team.ID, contactID, contact)
	//	}
	//}
	//if _, err = createRetrospective(ctx, tx, team.ID, "retro1", retro); err != nil {
	//	err = fmt.Errorf("failed to create retro: %w", err)
	//	return record, err
	//}
	return
}

//func createRetrospective(ctx context.Context, tx dal.ReadwriteTransaction, teamID, retroID string, retroData *models4retrospectus.Retrospective) (record dal.Record, err error) {
//	if err := retroData.Validate(); err != nil {
//		return nil, fmt.Errorf("invalid team retroData: %w", err)
//	}
//	retroKey := models4retrospectus.NewRetrospectiveKey(retroID, dal4teamus.NewTeamKey(teamID))
//	record = dal.NewRecordWithData(retroKey, retroData)
//	if err = tx.Set(ctx, record); err != nil {
//		return nil, err
//	}
//	return
//}
