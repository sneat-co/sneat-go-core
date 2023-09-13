package api4invitus

import (
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"github.com/sneat-co/sneat-go-core/modules/invitus/facade4invitus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dto4teamus"
	"net/http"
	"strings"
)

// httpGetPersonal is an API endpoint that returns personal invite data
func httpGetPersonal(w http.ResponseWriter, r *http.Request) {
	verifyOptions := verify.Request(verify.MinimumContentLength(apicore.MinJSONRequestSize), verify.MaximumContentLength(10*apicore.KB), verify.AuthenticationRequired(false))
	ctx, user, err := apicore.VerifyRequestAndCreateUserContext(w, r, verifyOptions)
	if err != nil {
		return
	}
	q := r.URL.Query()
	request := facade4invitus.GetPersonalInviteRequest{
		TeamRequest: dto4teamus.TeamRequest{
			TeamID: strings.TrimSpace(q.Get("teamID")),
		},
		InviteID: strings.TrimSpace(q.Get("inviteID")),
	}
	response, err := facade4invitus.GetPersonal(ctx, user, request)
	apicore.ReturnJSON(ctx, w, r, http.StatusOK, err, response)
}
