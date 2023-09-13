package api4contactus

import (
	"context"
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dto4contactus"
	"github.com/sneat-co/sneat-go-core/modules/contactus/facade4contactus"
	"net/http"
)

var setContactAddress = facade4contactus.SetContactsAddress

func httpSetContactAddress(w http.ResponseWriter, r *http.Request) {
	var request dto4contactus.SetContactAddressRequest
	handler := func(ctx context.Context, userCtx facade.User) (interface{}, error) {
		return nil, setContactAddress(ctx, userCtx, request)
	}
	verifyOptions := verify.Request(verify.MinimumContentLength(apicore.MinJSONRequestSize), verify.MaximumContentLength(10*apicore.KB), verify.AuthenticationRequired(true))
	apicore.HandleAuthenticatedRequestWithBody(w, r, &request, handler, http.StatusCreated, verifyOptions)
}
