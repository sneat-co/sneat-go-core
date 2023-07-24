package apicore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/sneat-co/sneat-go/src/core/facade"
	"github.com/sneat-co/sneat-go/src/core/httpserver"
	"io"
	"log"
	"net/http"
)

// IfNoErrorReturnOK returns HTTP status OK and empty response body
func IfNoErrorReturnOK(_ context.Context, w http.ResponseWriter, err error) {
	if err != nil {
		httpserver.HandleError(err, "IfNoErrorReturnOK", w)
		return
	}
	w.WriteHeader(http.StatusOK)
	//_, _ = w.Write([]byte(`{"success"":true}`))
}

// IfNoErrorReturnCreatedOK returns HTTP status OK and empty response body
func IfNoErrorReturnCreatedOK(_ context.Context, w http.ResponseWriter, err error) {
	if err != nil {
		httpserver.HandleError(err, "IfNoErrorReturnOK", w)
		return
	}
	w.WriteHeader(http.StatusCreated)
	//_, _ = w.Write([]byte(`{"success"":true}`))
}

// ReturnStatus returns provided HTTP status and empty response body
func ReturnStatus(_ context.Context, w http.ResponseWriter, statusCode int, err error) {
	if err != nil {
		httpserver.HandleError(err, "ReturnStatus", w)
		return
	}
	w.WriteHeader(statusCode)
}

// ReturnError returns provided HTTP status and empty response body
func ReturnError(_ context.Context, w http.ResponseWriter, err error) {
	httpserver.HandleError(err, "IfNoErrorReturnOK", w)
}

// ReturnJSON returns provided response as a JSON and sets "Content-Role" as "application/json"
func ReturnJSON(_ context.Context, w http.ResponseWriter, successStatusCode int, err error, response interface{}) {
	if err != nil {
		if errors.Is(err, facade.ErrUnauthorized) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprint(w, err.Error())
			return
		}
		httpserver.HandleError(err, "ReturnJSON", w)
		return
	}
	if v, ok := response.(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			httpserver.HandleError(fmt.Errorf("response is not valid: %w", err), "ReturnJSON", w)
			return
		}
	}
	log.Printf("ReturnJSON(successStatusCode=%v, err=%v)", successStatusCode, err)
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
	w.Header().Add("Content-TeamType", "application/json")
	if err = json.NewEncoder(w).Encode(response); err != nil {
		err = fmt.Errorf("failed to encode response to JSON: %w", err)
		sentry.CaptureException(err)
		w.WriteHeader(500) // Ask StackOverflow: Does it make sense?
		_, _ = io.WriteString(w, "Failed to encode response: ")
		_, _ = io.WriteString(w, err.Error())
		return
	}
}
