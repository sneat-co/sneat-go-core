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
			name:          "simple error",
			err:           errors.New("simple error"),
			wantErrorType: "*errors.errorString",
		},
		{
			name:          "single wrap",
			err:           fmt.Errorf("single wrap: %s", errors.New("simple error")),
			wantErrorType: "*errors.errorString",
		},
		{
			name:          "double wrap",
			err:           fmt.Errorf("doble wrapped: %w", fmt.Errorf("wrapped error: %s", errors.New("simple error"))),
			wantErrorType: "*errors.errorString",
		},
		{
			name:              "wrapped validation error",
			err:               fmt.Errorf("wrapped: %w", validation.NewErrRecordIsMissingRequiredField("field1")),
			wantErrorType:     "validation.ErrBadFieldValue",
			wantRootErrorType: "*errors.errorString",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErrorType, gotRootErrorType := getErrorTypes(tt.err)
			if gotErrorType != tt.wantErrorType {
				t.Errorf("gotErrorType = %v, want %v", gotErrorType, tt.wantErrorType)
			}
			if gotRootErrorType != tt.wantRootErrorType {
				t.Errorf("gotRootErrorType = %v, want %v", gotErrorType, tt.wantErrorType)
			}
		})
	}
}
