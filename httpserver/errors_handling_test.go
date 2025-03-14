package httpserver

import (
	"errors"
	"fmt"
	"github.com/strongo/validation"
	"testing"
)

func Test_getErrorTypes(t *testing.T) {
	tests := []struct {
		name              string
		err               error
		wantErrorType     string
		wantRootErrorType string
	}{
		{
			name:              "simple error",
			err:               errors.New("simple error"),
			wantErrorType:     "*errors.errorString",
			wantRootErrorType: "",
		},
		{
			name:              "single wrap",
			err:               fmt.Errorf("single wrap: %s", errors.New("simple error")),
			wantErrorType:     "*errors.errorString",
			wantRootErrorType: "",
		},
		{
			name:              "double wrap",
			err:               fmt.Errorf("doble wrapped: %w", fmt.Errorf("wrapped error: %s", errors.New("simple error"))),
			wantErrorType:     "*errors.errorString",
			wantRootErrorType: "",
		},
		{
			name:              "validation error",
			err:               validation.NewErrRecordIsMissingRequiredField("field1"),
			wantErrorType:     "validation.ErrBadFieldValue",
			wantRootErrorType: "",
		},
		{
			name:              "wrapped validation error",
			err:               fmt.Errorf("wrapped: %w", validation.NewErrRecordIsMissingRequiredField("field1")),
			wantErrorType:     "validation.ErrBadFieldValue",
			wantRootErrorType: "",
		},
		{
			name:              "double wrapped validation error",
			err:               fmt.Errorf("wrapped2: %w", fmt.Errorf("wrapped1: %w", validation.NewErrRecordIsMissingRequiredField("field1"))),
			wantErrorType:     "validation.ErrBadFieldValue",
			wantRootErrorType: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErrorType, gotRootErrorType := getErrorTypes(tt.err)
			if gotErrorType != tt.wantErrorType {
				t.Errorf("gotErrorType = %v, want %v", gotErrorType, tt.wantErrorType)
			}
			if gotRootErrorType != tt.wantRootErrorType {
				t.Errorf("gotRootErrorType = %v, want %v", gotRootErrorType, tt.wantRootErrorType)
			}
		})
	}
}
