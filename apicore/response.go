package apicore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/httpserver"
	"github.com/sneat-co/sneat-go-core/monitoring"
	"github.com/strongo/logus"
	"io"
	"net/http"
)

// IfNoErrorReturnOK returns HTTP status OK and empty response body
func IfNoErrorReturnOK(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		httpserver.HandleError(ctx, err, "IfNoErrorReturnOK", w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	//_, _ = w.Write([]byte(`{"success"":true}`))
}

// IfNoErrorReturnCreatedOK returns HTTP status OK and empty response body
func IfNoErrorReturnCreatedOK(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		httpserver.HandleError(ctx, err, "IfNoErrorReturnOK", w, r)
		return
	}
	w.WriteHeader(http.StatusCreated)
	//_, _ = w.Write([]byte(`{"success"":true}`))
}

// ReturnStatus returns provided HTTP status and empty response body
func ReturnStatus(ctx context.Context, w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	if err != nil {
		httpserver.HandleError(ctx, err, "ReturnStatus", w, r)
		return
	}

	if statusCode == http.StatusOK {
		statusCode = http.StatusNoContent
	} else if statusCode >= 400 {
		panic("error code should be accompanied by an err")
	}
	w.WriteHeader(statusCode)
}

// ReturnError returns provided HTTP status and empty response body
func ReturnError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		err = errors.New("an attempt to return nil err")
	}
	httpserver.HandleError(ctx, err, "ReturnError", w, r)
}

// ReturnJSON returns response as JSON and sets header "Content-Role=application/json" if err=nil.
// If the response is validatable, it will be validated.
// If passed err is not nil, the response is ignored and an error content is returned as response body.
func ReturnJSON(ctx context.Context, w http.ResponseWriter, r *http.Request, successStatusCode int, err error, response interface{}) {
	if err != nil {
		if errors.Is(err, facade.ErrUnauthorized) || errors.Is(err, facade.ErrUnauthenticated) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprint(w, err.Error())
			return
		}
		httpserver.HandleError(ctx, err, "ReturnJSON", w, r)
		return
	}
	if v, ok := response.(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			httpserver.HandleError(ctx, fmt.Errorf("response is not valid: %w", err), "ReturnJSON", w, r)
			return
		}
	}
	logus.Debugf(ctx, "ReturnJSON(successStatusCode=%v, err=%v)", successStatusCode, err)
	w.WriteHeader(successStatusCode)
	if successStatusCode == http.StatusNoContent {
		if response != nil {
			panic(fmt.Sprintf("ReturnJSON: successStatusCode is http.StatusNoContent=204 but response is not nil: %T=%+v", response, response))
		}
		return
	}
	if response == nil && successStatusCode == http.StatusOK {
		panic("ReturnJSON: response is nil but successStatusCode is http.StatusOK=200, expected to be http.StatusNoContent=204")
	}
	w.Header().Add("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(response); err != nil {
		err = fmt.Errorf("failed to encode response to JSON: %w", err)
		monitoring.CaptureException(ctx, err)
		w.WriteHeader(500) // Ask StackOverflow: Does it make sense?
		_, _ = io.WriteString(w, "Failed to encode response: ")
		_, _ = io.WriteString(w, err.Error())
		return
	}
}
