package api4teamus

import (
	"context"
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dto4teamus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/facade4teamus"
	"net/http"
)

var createTeam = facade4teamus.CreateTeam

// httpPostCreateTeam is an API endpoint that creates a new team
func httpPostCreateTeam(w http.ResponseWriter, r *http.Request) {
	var request dto4teamus.CreateTeamRequest
	handler := func(ctx context.Context, userCtx facade.User) (interface{}, error) {
		return createTeam(ctx, userCtx, request)
	}
	verifyOptions := verify.Request(verify.MinimumContentLength(apicore.MinJSONRequestSize), verify.MaximumContentLength(10*apicore.KB), verify.AuthenticationRequired(true))
	apicore.HandleAuthenticatedRequestWithBody(w, r, &request, handler, http.StatusCreated, verifyOptions)
}
