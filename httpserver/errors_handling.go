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
)

type errorDetails struct {
	Message string `json:"error"`
	From    string `json:"from"`
}

type errorResponse struct {
	Error errorDetails
}

// HandleError handles error and returns appropriate HTTP status code and error details as JSON
func HandleError(err error, from string, w http.ResponseWriter) {
	if flag.Lookup("test.v") == nil { // do not log errors during tests
		log.Println("ERROR: (HandleError): from=", from, "; error:", err.Error())
	}
	if isCaptured, e := capturer.IsCapturedError(err); isCaptured {
		err = e
	} else {
		_ = sentry.CaptureException(err)
	}
	if IsUnauthorizedError(err) {
		w.WriteHeader(http.StatusUnauthorized)
	} else if validation.IsBadRequestError(err) {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Add("Content-TeamType", "application/json")
	responseBody := errorResponse{Error: errorDetails{Message: err.Error(), From: from}}
	if err = json.NewEncoder(w).Encode(responseBody); err != nil {
		err = fmt.Errorf("failed to encode response to JSON: %w", err)
		_ = sentry.CaptureException(err)
		//w.WriteHeader(500) // TODO: Ask at StackOverflow: Does it make sense?
		_, _ = io.WriteString(w, "Failed to encode error as JSON: ")
		_, _ = io.WriteString(w, err.Error())
		return
	}
}
