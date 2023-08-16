package apicore

import (
	"context"
	"github.com/sneat-co/sneat-go-core/httpserver"
	"github.com/sneat-co/sneat-go-core/monitoring"
	"log"
	"net/http"
)

// RequestDTO defines an interface that should be implemented by request DTO struct
type RequestDTO interface {
	Validate() error
}

// ResponseDTO common interface for response objects
type ResponseDTO interface {
	// Validate validates response
	Validate() error
}

// VerifyRequestOptions - options for request verification
type VerifyRequestOptions interface { // TODO: move to shared Sneat package
	MinimumContentLength() int64
	MaximumContentLength() int64
	AuthenticationRequired() bool
}

type ContextProvider = func(r *http.Request) (context.Context, error)

type Worker = func(ctx context.Context) (responseDTO ResponseDTO, err error)

// Execute is very similar to HandleAuthenticatedRequestWithBody() // TODO: consider code unification & reuse
var Execute = func(
	w http.ResponseWriter,
	r *http.Request,
	request RequestDTO,
	verifyOptions VerifyRequestOptions,
	successStatusCode int,
	getContext ContextProvider,
	handler Worker,
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
		httpserver.HandleError(err, "Execute", w, r)
		return
	}
	response, err := handler(ctx)
	if err != nil {
		monitoring.CaptureException(err)
		err = nil
	}
	ReturnJSON(ctx, w, r, successStatusCode, err, response)
}
