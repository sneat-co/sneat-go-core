package apicore

import (
	"context"
	"github.com/sneat-co/sneat-go-core/facade"
	"net/http"
)

// FacadeHandler TODO:?
type FacadeHandler = func(
	ctx context.Context,
	userCtx facade.User,
) (response interface{}, err error)

// HandleAuthenticatedRequestWithBody is very similar to Execute // TODO: consider code unification & reuse
func HandleAuthenticatedRequestWithBody(w http.ResponseWriter, r *http.Request,
	request interface{ Validate() error },
	facadeHandler FacadeHandler,
	successStatusCode int,
	options VerifyRequestOptions,
) {
	ctx, userContext, err := VerifyAuthenticatedRequestAndDecodeBody(w, r, options, request)
	if err != nil { // The error has been handled inside the function
		return
	}
	var response interface{}
	response, err = facadeHandler(ctx, userContext)
	ReturnJSON(ctx, w, r, successStatusCode, err, response)
}
