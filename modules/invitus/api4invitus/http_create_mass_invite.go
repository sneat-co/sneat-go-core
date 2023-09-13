package api4invitus

import (
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"github.com/sneat-co/sneat-go-core/modules/invitus/facade4invitus"
	"net/http"
)

var createMassInvite = facade4invitus.CreateMassInvite

// httpPostCreateMassInvite is an API endpoint to create a mass-invite
func httpPostCreateMassInvite(w http.ResponseWriter, r *http.Request) {
	verifyOptions := verify.Request(verify.MinimumContentLength(apicore.MinJSONRequestSize), verify.MaximumContentLength(10*apicore.KB), verify.AuthenticationRequired(true))
	ctx, err := apicore.VerifyRequest(w, r, verifyOptions)
	if err != nil {
		return
	}
	var request facade4invitus.CreateMassInviteRequest
	if err = apicore.DecodeRequestBody(w, r, &request); err != nil {
		return
	}
	response, err := createMassInvite(ctx, request)
	apicore.ReturnJSON(ctx, w, r, http.StatusCreated, err, response)
}
