package api4memberus

import (
	"context"
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dal4contactus"
	"github.com/sneat-co/sneat-go-core/modules/memberus/facade4memberus"
	"net/http"
)

var createMember = facade4memberus.CreateMember

// httpPostCreateMember is an API endpoint that adds a members to a team
func httpPostCreateMember(w http.ResponseWriter, r *http.Request) {
	var request dal4contactus.CreateMemberRequest
	handler := func(ctx context.Context, userCtx facade.User) (interface{}, error) {
		return createMember(ctx, userCtx, request)
	}
	verifyOptions := verify.Request(verify.MinimumContentLength(apicore.MinJSONRequestSize), verify.MaximumContentLength(10*apicore.KB), verify.AuthenticationRequired(true))
	apicore.HandleAuthenticatedRequestWithBody(w, r, &request, handler, http.StatusCreated, verifyOptions)
}
