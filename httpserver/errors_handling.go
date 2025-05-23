package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/sneat-co/sneat-go-core/capturer"
	"github.com/sneat-co/sneat-go-core/monitoring"
	"github.com/strongo/logus"
	"github.com/strongo/validation"
	"io"
	"net/http"
	"reflect"
)

type errorDetails struct {
	Message     string `json:"message"`
	From        string `json:"from,omitempty"`
	Type        string `json:"type"`
	RootErrType string `json:"rootErrType,omitempty"`
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
	if flag.Lookup("test.v") == nil { // do not log errors during mock_module
		logus.Errorf(ctx, "HandleError: from=%s; error: %s", from, err)
	}

	if isCaptured, e := capturer.IsCapturedError(err); isCaptured {
		err = e
	} else {
		_ = monitoring.CaptureError(ctx, err)
	}

	var statusCode int
	switch {
	case validation.IsBadRequestError(err):
		statusCode = http.StatusBadRequest
	case IsUnauthorizedError(err):
		statusCode = http.StatusUnauthorized
	default:
		statusCode = http.StatusInternalServerError
	}

	w.WriteHeader(statusCode)

	AccessControlAllowOrigin(w, r)
	w.Header().Add("Content-Type", "application/json")

	responseBody := errorResponse{
		Error: errorDetails{
			From:    from,
			Message: err.Error(),
		},
	}
	responseBody.Error.Type, responseBody.Error.RootErrType = getErrorTypes(err)

	if content, err := json.Marshal(responseBody); err != nil {
		err = fmt.Errorf("failed to encode response to JSON: %w", err)
		logus.Errorf(ctx, "HandleError: %v", err)
		_ = monitoring.CaptureError(ctx, err)
		//w.WriteHeader(500) // TODO: Ask at StackOverflow: Does it make sense?
		_, _ = io.WriteString(w, "Failed to encode error as JSON: ")
		_, _ = io.WriteString(w, err.Error())
		return
	} else {
		_, err = w.Write(content)
		if err != nil {
			logus.Errorf(ctx, "HandleError: failed to write response body: %v", err)
		}
	}
}

func getErrorTypes(err error) (errorType, rootErrorType string) {
	errorType = reflect.TypeOf(err).String()

	const wrapErrorType = "*fmt.wrapError"

	for errorType == wrapErrorType {
		if err = errors.Unwrap(err); err == nil {
			break
		}
		errorType = reflect.TypeOf(err).String()
	}
	for {
		if err = errors.Unwrap(err); err == nil {
			break
		}
		if errType := reflect.TypeOf(err).String(); errType == "*errors.errorString" {
			break
		} else if errType != wrapErrorType {
			rootErrorType = errType
		}
	}
	return
}
