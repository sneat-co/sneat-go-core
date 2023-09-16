package api4memberus

import (
	"context"
	"firebase.google.com/go/v4/auth"
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/apicore/httpmock"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/modules/contactus/briefs4contactus"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dal4contactus"
	"github.com/sneat-co/sneat-go-core/modules/contactus/models4contactus"
	"github.com/sneat-co/sneat-go-core/modules/memberus/briefs4memberus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dto4teamus"
	"github.com/sneat-co/sneat-go-core/sneatfb"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHttpAddMember(t *testing.T) {

	const teamID = "unit-test"
	request := dal4contactus.CreateMemberRequest{
		TeamRequest: dto4teamus.TeamRequest{
			TeamID: teamID,
		},
		Relationship: "spouse",
		MemberBase: briefs4memberus.MemberBase{
			ContactBase: briefs4contactus.ContactBase{
				ContactBrief: briefs4contactus.ContactBrief{
					Type:     briefs4contactus.ContactTypePerson,
					Gender:   "unknown",
					Title:    "Some new members",
					AgeGroup: "unknown",
					WithRoles: dbmodels.WithRoles{
						Roles: []string{briefs4memberus.TeamMemberRoleContributor},
					},
				},
				Status: "active",
				//WithRequiredCountryID: dbmodels.WithRequiredCountryID{
				//	CountryID: "--",
				//},
				Emails: []dbmodels.PersonEmail{
					{Type: "personal", Address: "someone@example.com"},
				},
			},
		}}
	request.CountryID = "IE"

	apicore.UserContextProvider = func() facade.User {
		return facade.NewUser("TestUserID")
	}

	//t.Log(buffer.String())

	req := httpmock.NewPostJSONRequest("POST", "/v0/team/create_member", request)
	req.Host = "localhost"
	req.Header.Set("Origin", "http://localhost:3000")

	createMember = func(ctx context.Context, userCtx facade.User, request dal4contactus.CreateMemberRequest) (response dal4contactus.CreateTeamMemberResponse, err error) {
		if request.TeamID != teamID {
			t.Fatalf("Expected teamID=%v, got: %v", teamID, request.TeamID)
		}
		response.Member.ID = "abc1"
		response.Member.Data = &models4contactus.ContactDto{
			ContactBase: briefs4contactus.ContactBase{
				ContactBrief: briefs4contactus.ContactBrief{
					Type:  briefs4contactus.ContactTypeCompany,
					Title: "Some company",
					WithOptionalCountryID: dbmodels.WithOptionalCountryID{
						CountryID: "IE",
					},
					WithRoles: dbmodels.WithRoles{
						Roles: []string{briefs4memberus.TeamMemberRoleContributor},
					},
				},
				Status: "active",
				//WithRequiredCountryID: dbmodels.WithRequiredCountryID{
				//	CountryID: "--",
				//},
			},
		}
		response.Member.Data = &models4contactus.ContactDto{
			ContactBase: response.Member.Data.ContactBase,
		}
		return
	}

	const uid = "unit-test-user"
	apicore.NewContextWithToken = func(r *http.Request, authRequired bool) (ctx context.Context, err error) {
		return sneatfb.NewContextWithFirebaseToken(r.Context(), &auth.Token{UID: uid}), nil
	}
	sneatfb.NewFirebaseAuthToken = func(ctx context.Context, fbIDToken func() (string, error), authRequired bool) (*auth.Token, error) {
		return &auth.Token{UID: uid}, nil
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(httpPostCreateMember)
	handler.ServeHTTP(rr, req)

	responseBody := rr.Body.String()

	if expected := http.StatusCreated; rr.Code != expected {
		t.Fatalf(
			"unexpected status: got (%v) expects (%v): %v",
			rr.Code,
			expected,
			responseBody,
		)
	}
}
