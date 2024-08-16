package apicore

import (
	"context"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"github.com/sneat-co/sneat-go-core/facade"
	"net/http"
)

// FacadeHandler defines a function that handles a request
type FacadeHandler = func(ctx context.Context, userCtx facade.UserContext) (response any, err error)

// HandleAuthenticatedRequestWithBody is very similar to Execute - consider code unification & reuse
func HandleAuthenticatedRequestWithBody(
	w http.ResponseWriter,
	r *http.Request,
	request interface{ Validate() error },
	options verify.RequestOptions,
	successStatusCode int,
	facadeHandler FacadeHandler,
) {
	ctx, userContext, err := VerifyAuthenticatedRequestAndDecodeBody(w, r, options, request)
	if err != nil {
		return // No need to write to response as the error has been handled inside the called function
	}
	var response any
	response, err = facadeHandler(ctx, userContext)
	ReturnJSON(ctx, w, r, successStatusCode, err, response)
}
