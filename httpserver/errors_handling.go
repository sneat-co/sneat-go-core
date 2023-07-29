package httpserver

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/sneat-co/sneat-go/src/core/capturer"
	"github.com/strongo/validation"
	"io"
	"log"
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
	Error errorDetails
}

// HandleError handles error and returns appropriate HTTP status code and error details as JSON
func HandleError(err error, from string, w http.ResponseWriter, r *http.Request) {
	if flag.Lookup("test.v") == nil { // do not log errors during tests
		log.Printf("ERROR: HandleError: from=%s; error: %v", from, err)
	}
	if isCaptured, e := capturer.IsCapturedError(err); isCaptured {
		err = e
	} else {
		_ = sentry.CaptureException(err)
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
		log.Printf("ERROR: HandleError: %v", err)
		_ = sentry.CaptureException(err)
		//w.WriteHeader(500) // TODO: Ask at StackOverflow: Does it make sense?
		_, _ = io.WriteString(w, "Failed to encode error as JSON: ")
		_, _ = io.WriteString(w, err.Error())
		return
	} else {
		_, err = w.Write(content)
		if err != nil {
			log.Printf("ERROR: HandleError: failed to write response body: %v", err)
		}
	}
}
