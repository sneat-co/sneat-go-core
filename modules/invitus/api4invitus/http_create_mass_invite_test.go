package api4invitus

import (
	"bytes"
	"context"
	"encoding/json"
	"firebase.google.com/go/v4/auth"
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/modules/invitus/facade4invitus"
	"github.com/sneat-co/sneat-go-core/modules/invitus/models4invitus"
	"github.com/sneat-co/sneat-go-core/sneatfb"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateMassInvite(t *testing.T) {
	const teamID = "unit-test"
	var invite models4invitus.MassInvite
	invite.Type = "mass"
	invite.Channel = "email"
	invite.Roles = []string{
		"contributor",
		"test-role1",
	}
	invite.From = models4invitus.InviteFrom{
		InviteContact: models4invitus.InviteContact{
			Channel:  "email",
			Address:  "from@example.com",
			Title:    "From Title",
			MemberID: "f1",
		},
	}
	//invite.To = &models4invitus.InviteTo{
	//	Channel:      "email",
	//	Address:      "to@example.com",
	//	Title:        "To Title",
	//	ToTeamMemberID: "t1",
	//}
	invite.TeamID = teamID
	invite.Team.Type = "family"
	invite.Team.Title = "Unit Test"
	invite.Created.Client.HostOrApp = "unit-test"
	invite.Created.Client.RemoteAddr = "127.0.0.1"
	invite.Created.At = time.Now()
	invite.From.UserID = "u1"
	invite.Status = "active"

	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	if err := encoder.Encode(facade4invitus.CreateMassInviteRequest{Invite: invite}); err != nil {
		t.Fatal(err)
	}
	//t.Log(buffer.String())

	req, err := http.NewRequest("POST", "/api4meetingus/create-invite", buffer)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://localhost:3000")

	createMassInvite = func(ctx context.Context, request facade4invitus.CreateMassInviteRequest) (response facade4invitus.CreateMassInviteResponse, err error) {
		response.ID = "test-id"
		return
	}
	apicore.NewContextWithToken = func(r *http.Request, authRequired bool) (ctx context.Context, err error) {
		return sneatfb.NewContextWithFirebaseToken(r.Context(), &auth.Token{UID: "unit-test-user"}), nil
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(httpPostCreateMassInvite)
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

	var response facade4invitus.CreateMassInviteResponse
	if err = json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err, responseBody)
	}
	if response.ID == "" {
		t.Fatal("Response is missing ContactID of created invite")
	}
}
