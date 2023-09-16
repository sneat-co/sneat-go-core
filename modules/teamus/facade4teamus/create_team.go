package facade4teamus

import (
	"context"
	"errors"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/gosimple/slug"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dal4contactus"
	"github.com/sneat-co/sneat-go-core/modules/memberus/briefs4memberus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dal4teamus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dto4teamus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/models4teamus"
	"github.com/sneat-co/sneat-go-core/modules/userus/models4userus"
	"github.com/strongo/random"
	"strings"
)

// CreateTeam creates TeamIDs record
func CreateTeam(ctx context.Context, userContext facade.User, request dto4teamus.CreateTeamRequest) (response dto4teamus.CreateTeamResponse, err error) {
	//var TeamKey *firestore.DocumentRef
	db := facade.GetDatabase(ctx)
	err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) (err error) {
		response, err = createTeamTxWorker(ctx, userContext.GetID(), tx, request)
		return err
	})
	if err != nil {
		return response, fmt.Errorf("failed to create a new team of type [%v]: %w", request.Type, err)
	}
	return response, nil
}

func createTeamTxWorker(ctx context.Context, userID string, tx dal.ReadwriteTransaction, request dto4teamus.CreateTeamRequest) (response dto4teamus.CreateTeamResponse, err error) {
	if strings.TrimSpace(userID) == "" {
		return response, facade.ErrUnauthenticated
	}
	var memberID string
	user := models4userus.NewUserContext(userID)
	if err = tx.Get(ctx, user.Record); err != nil {
		return
	}

	memberID, err = dbmodels.GenerateIDFromNameOrRandom(*user.Dto.Name, nil)
	if err != nil {
		return response, fmt.Errorf("failed to generate  member ContactID: %w", err)
	}

	roles := []string{
		briefs4memberus.TeamMemberRoleTeamMember,
		briefs4memberus.TeamMemberRoleCreator,
		briefs4memberus.TeamMemberRoleOwner,
		briefs4memberus.TeamMemberRoleContributor,
	}
	if request.Type == "family" {
		roles = append(roles, briefs4memberus.TeamMemberRoleAdult)
	}

	if request.Type == "family" && request.Title == "" {
		request.Title = "Family"
	}
	teamDto := &models4teamus.TeamDto{
		TeamBrief: models4teamus.TeamBrief{
			Type:  request.Type,
			Title: request.Title,
		},
		WithUserIDs: dbmodels.WithUserIDs{
			UserIDs: []string{userID},
		},
		//WithMembers: models4memberus.WithMembers{
		//	Members: []*briefs4memberus.MemberBrief{&teamMember},
		//},
		NumberOf: map[string]int{
			"members": 1,
		},
	}
	teamDto.CountryID = user.Dto.CountryID
	if request.Type == "work" {
		zero := 0
		hundred := 100
		teamDto.Metrics = []*models4teamus.TeamMetric{
			{ID: "cc", Title: "Code coverage", Type: "int", Mode: "TeamIDs", Min: &zero, Max: &hundred},
			{ID: "bb", Title: "Build is broken", Type: "bool", Mode: "TeamIDs", Bool: &models4teamus.BoolMetric{
				True:  &models4teamus.BoolMetricVal{Label: "Yes", Color: "danger"},
				False: &models4teamus.BoolMetricVal{Label: "No", Color: "success"},
			}},
			{ID: "wfh", Title: "Working From Home", Type: "bool", Mode: "personal", Bool: &models4teamus.BoolMetric{
				True:  &models4teamus.BoolMetricVal{Label: "Yes", Color: "tertiary"},
				False: &models4teamus.BoolMetricVal{Label: "No", Color: "secondary"},
			}},
		}
	}

	if err = teamDto.Validate(); err != nil {
		return response, fmt.Errorf("teamDto reacord is not valid: %w", err)
	}

	var teamID string
	title := request.Title
	if request.Type == "family" {
		title = ""
	}
	teamID, err = getUniqueTeamID(ctx, tx, title)
	if err != nil {
		return response, fmt.Errorf("failed to get an unique ContactID for a new teamDto: %w", err)
	}
	teamKey := dal.NewKeyWithID(dal4teamus.TeamsCollection, teamID)

	teamRecord := dal.NewRecordWithData(teamKey, teamDto)
	if err = tx.Insert(ctx, teamRecord); err != nil {
		return response, fmt.Errorf("failed to insert a new teamDto record: %w", err)
	}

	teamContactus := dal4contactus.NewContactusTeamContext(teamID)

	teamMember := user.Dto.ContactBrief
	teamMember.UserID = userID
	teamMember.Roles = roles
	if teamMember.Gender == "" {
		teamMember.Gender = "unknown"
	}
	if user.Dto.Defaults != nil && len(user.Dto.Defaults.ShortNames) > 0 {
		teamMember.ShortTitle = user.Dto.Defaults.ShortNames[0].Name
	}
	//if len(teamMember.Emails) == 0 && len(user.Emails) > 0 {
	//	teamMember.Emails = user.Emails
	//}
	//if len(teamMember.Phones) == 0 && len(user.Dto.Phones) > 0 {
	//	teamMember.Phones = user.Dto.Phones
	//}
	teamContactus.Data.AddContact(memberID, &teamMember)

	if err := tx.Insert(ctx, teamContactus.Record); err != nil {
		return response, fmt.Errorf("failed to insert a new teamDto contactus record: %w", err)
	}

	teamInfo := models4userus.UserTeamBrief{
		TeamBrief: models4teamus.TeamBrief{
			Type:  request.Type,
			Title: request.Title,
		},
		//MemberType: briefs4memberus.TeamMemberTypeCreator,
		Roles: roles,
	}

	if user.Dto.Teams == nil {
		user.Dto.Teams = make(map[string]*models4userus.UserTeamBrief, 1)
	}
	updates := user.Dto.SetTeamBrief(teamID, &teamInfo)
	if err = user.Dto.Validate(); err != nil {
		return response, fmt.Errorf("user record is not valid after adding new team info: %v", err)
	}
	if user.Record.Exists() {
		if err = tx.Update(ctx, user.Key, updates); err != nil {
			return response, fmt.Errorf("failed to update user record with a new teamDto info: %w", err)
		}
	} else {
		if err = tx.Insert(ctx, user.Record); err != nil {
			return response, fmt.Errorf("failed to insert new user record: %w", err)
		}
	}

	teamMember.Roles = roles
	if _, err = CreateMemberRecordFromBrief(ctx, tx, teamID, memberID, teamMember); err != nil {
		return response, fmt.Errorf("failed to create member's record: %w", err)
	}

	response.Team.ID = teamID
	response.Team.Dto.Type = teamInfo.Type
	response.Team.Dto.Title = teamInfo.Title
	return
}

func getUniqueTeamID(ctx context.Context, getter dal.ReadSession, title string) (teamID string, err error) {
	if title == "" || briefs4memberus.IsKnownTeamMemberRole(title, []string{}) {
		teamID = random.ID(5)
	} else {
		teamID = strings.Replace(slug.Make(title), "-", "", -1)
	}
	const maxAttemptsCount = 9
	for i := 0; i <= maxAttemptsCount; i++ {
		if i == maxAttemptsCount {
			return "", errors.New("too many attempts to get an unique team ContactID")
		}
		teamID = strings.ToLower(teamID)
		teamKey := dal.NewKeyWithID(dal4teamus.TeamsCollection, teamID)
		teamRecord := dal.NewRecordWithData(teamKey, nil)
		if err = getter.Get(ctx, teamRecord); dal.IsNotFound(err) {
			return teamID, nil
		} else if err != nil {
			return teamID, err
		}
		if i == 0 && title != "" {
			teamID += "_"
		}
		teamID += random.ID(1)
	}
	return teamID, nil
}
