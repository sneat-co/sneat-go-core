package facade4memberus

import (
	"context"
	"errors"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/modules/contactus/briefs4contactus"
	"github.com/sneat-co/sneat-go-core/modules/contactus/const4contactus"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dal4contactus"
	"github.com/sneat-co/sneat-go-core/modules/contactus/models4contactus"
	"github.com/sneat-co/sneat-go-core/modules/memberus/briefs4memberus"
	"github.com/sneat-co/sneat-go-core/modules/memberus/dal4memberus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dal4teamus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/facade4teamus"
	"github.com/sneat-co/sneat-go-core/modules/userus/models4userus"
	"github.com/strongo/validation"
)

// CreateMember adds members to a team
func CreateMember(
	ctx context.Context,
	user facade.User,
	request dal4contactus.CreateMemberRequest,
) (
	response dal4contactus.CreateTeamMemberResponse,
	err error,
) {
	if err = request.Validate(); err != nil {
		err = fmt.Errorf("bad request for facademember.CreateMember(): %w", err)
		return
	}

	err = dal4teamus.CreateTeamItem(ctx, user, "members", request.TeamRequest,
		const4contactus.ModuleID,
		func(ctx context.Context, tx dal.ReadwriteTransaction, params *dal4teamus.ModuleTeamWorkerParams[*models4contactus.ContactusTeamDto]) (err error) {
			team := params.Team
			contactusTeam := dal4contactus.NewContactusTeamContext(params.Team.ID)
			if err := tx.Get(ctx, contactusTeam.Record); err != nil {
				return fmt.Errorf("failed to get contactus team record: %w", err)
			}

			if len(contactusTeam.Data.Contacts) == 0 {
				return errors.New("team has no members")
			}
			contactID, userMember := contactusTeam.Data.GetContactBriefByUserID(params.UserID)
			if userMember == nil {
				return errors.New("user does not belong to the team: " + params.UserID)
			}
			switch userMember.AgeGroup {
			case "", dbmodels.AgeGroupUnknown:
				switch request.Relationship {
				case dbmodels.RelationshipSpouse, dbmodels.RelationshipChild:
					userMember.AgeGroup = dbmodels.AgeGroupAdult
					userMemberKey := dal4memberus.NewMemberKey(request.TeamID, contactID)
					if err = tx.Update(ctx, userMemberKey, []dal.Update{
						{
							Field: "ageGroup",
							Value: userMember.AgeGroup,
						},
					}); err != nil {
						return fmt.Errorf("failed to update member record: %w", err)
					}
				}
			}
			memberBrief := request.ContactBrief
			memberBrief.ShortTitle = getShortTitle(request.Title, contactusTeam.Data.Contacts)
			if team.Data.Type == "family" {
				memberBrief.Roles = []string{
					briefs4memberus.TeamMemberRoleContributor,
				}
			}

			if memberBrief.Name.First != "" && briefs4memberus.IsUniqueShortTitle(memberBrief.Name.First, contactusTeam.Data.Contacts, briefs4memberus.TeamMemberRoleTeamMember) {
				memberBrief.ShortTitle = memberBrief.Name.First
			} else if memberBrief.Name.Full != "" {
				memberBrief.ShortTitle = getShortTitle(memberBrief.Name.Full, contactusTeam.Data.Contacts)
			}

			//if request.Emails != "" {
			//	memberBrief.Avatar = &dbprofile.Avatar{
			//		Gravatar: fmt.Sprintf("%x", md5.Sum([]byte(strings.ToLower(request.Email)))),
			//	}
			//}

			//if memberBrief.Name.First != "" && memberBrief.Name.Last != "" {
			//
			//}
			contactID, err = dbmodels.GenerateIDFromNameOrRandom(*memberBrief.Name, contactusTeam.Data.ContactIDs())
			if err != nil {
				return fmt.Errorf("failed to generate new member ContactID: %w", err)
			}

			var from string
			memberFoundByID := false
			for _, m := range contactusTeam.Data.Contacts {
				if m.UserID == params.UserID {
					memberFoundByID = true
					from = m.GetTitle()
					if from == "" {
						from = "userID=" + params.UserID
					}
				}
			}
			if !memberFoundByID {
				err = validation.NewErrBadRequestFieldValue("userID", "user does not belong to the team: userID="+params.UserID)
				return
			}
			if from == "" {
				err = validation.NewErrBadRequestFieldValue("userID", "team member has no title: userID="+params.UserID)
				return
			}
			{ // Update team record
				params.TeamUpdates = append(params.TeamUpdates,
					contactusTeam.Data.AddContact(contactID, &memberBrief),
				)
			}

			response.Member, err = facade4teamus.CreateMemberRecordFromBrief(ctx, tx, request.TeamID, contactID, memberBrief)
			if err != nil {
				return fmt.Errorf("failed to create member's record: %w", err)
			}

			if err = txUpdateMemberGroup(ctx, tx, params.Started, params.Team.Data, params.Team.Key, params.TeamUpdates); err != nil {
				return fmt.Errorf("failed to update team record: %w", err)
			}
			if contactID == "" {
				panic("contactID is empty")
			}
			response.Member.ID = contactID
			return nil
		},
	)
	if err != nil {
		err = fmt.Errorf("failed to create new team member: %w", err)
		return response, err
	}

	if response.Member.Data == nil {
		err = errors.New("response.ContactID.Data == nil")
		//log.Errorf(ctx, "response.Data == nil")
	}
	return response, err
}

func getShortTitle(title string, members map[string]*briefs4contactus.ContactBrief) string {
	shortNames := models4userus.GetShortNames(title)
	for _, short := range shortNames {
		isUnique := true
		for _, m := range members {
			if m.ShortTitle == short.Name {
				isUnique = false
				break
			}
		}
		if isUnique {
			return short.Name
		}
	}
	return ""
}
