package httpserver

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/sneat-co/sneat-go-core/capturer"
	"github.com/sneat-co/sneat-go-core/monitoring"
	"github.com/strongo/log"
	"github.com/strongo/validation"
	"io"
	"net/http"
	"reflect"
)

type errorDetails struct {
	Message string `json:"message"`
	From    string `json:"from,omitempty"`
	Type    string `json:"type"`
}

func (e errorDetails) String() string {
	return fmt.Sprintf("From %s: %s: %s", e.From, e.Type, e.Message)
}

type errorResponse struct {
	Error errorDetails `json:"error"`
}

// HandleError handles error and returns appropriate HTTP status code and error details as JSON
var HandleError = func(ctx context.Context, err error, from string, w http.ResponseWriter, r *http.Request) {
	if ctx == nil {
		ctx = r.Context()
	}
	if flag.Lookup("test.v") == nil { // do not log errors during tests
		log.Errorf(ctx, "HandleError: from=%s; error: %s", from, err)
	}
	if isCaptured, e := capturer.IsCapturedError(err); isCaptured {
		err = e
	} else {
		_ = monitoring.CaptureException(ctx, err)
	}
	AccessControlAllowOrigin(w, r)
	if IsUnauthorizedError(err) {
		w.WriteHeader(http.StatusUnauthorized)
	} else if validation.IsBadRequestError(err) {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Add("Content-Type", "application/json")
	responseBody := errorResponse{
		Error: errorDetails{
			From:    from,
			Message: err.Error(),
			Type:    reflect.TypeOf(err).String(),
		},
	}

	if content, err := json.Marshal(responseBody); err != nil {
		err = fmt.Errorf("failed to encode response to JSON: %w", err)
		log.Errorf(ctx, "HandleError: %v", err)
		_ = monitoring.CaptureException(ctx, err)
		//w.WriteHeader(500) // TODO: Ask at StackOverflow: Does it make sense?
		_, _ = io.WriteString(w, "Failed to encode error as JSON: ")
		_, _ = io.WriteString(w, err.Error())
		return
	} else {
		_, err = w.Write(content)
		if err != nil {
			log.Errorf(ctx, "HandleError: failed to write response body: %v", err)
		}
	}
}
