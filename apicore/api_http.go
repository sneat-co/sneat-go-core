package apicore

import (
	"context"
	"github.com/datatug/datatug/packages/server/endpoints"
	"github.com/getsentry/sentry-go"
	"github.com/sneat-co/sneat-go/src/core/httpserver"
	"log"
	"net/http"
)

var _ endpoints.Handler = Execute

// Execute is very similar to HandleAuthenticatedRequestWithBody() // TODO: consider code unification & reuse
var Execute = func(
	w http.ResponseWriter,
	r *http.Request,
	request endpoints.RequestDTO,
	verifyOptions endpoints.VerifyRequestOptions,
	successStatusCode int,
	getContext endpoints.ContextProvider,
	handler endpoints.Worker,
) {
	log.Printf("apicore.Execute(successStatusCode=%v)", successStatusCode)

	_, err := VerifyRequest(w, r, verifyOptions)
	if err != nil {
		log.Printf("failed to verify request: %v", err)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodDelete {
		if err = DecodeRequestBody(w, r, request); err != nil {
			log.Printf("failed to decode request body: %v", err)
			return
		}
	}
	var ctx context.Context
	ctx, err = getContext(r)
	if err != nil {
		log.Printf("failed to get request context: %v", err)
		httpserver.HandleError(err, "Execute", w)
		return
	}
	response, err := handler(ctx)
	if err != nil {
		sentry.CaptureException(err)
		err = nil
	}
	ReturnJSON(ctx, w, successStatusCode, err, response)
}
