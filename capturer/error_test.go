package capturer

import (
	"errors"
	"testing"
)

func Test_capturedError_Error(t *testing.T) {
	type fields struct {
		error error
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "should_pass",
			want: "captured error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &capturedError{
				error: tt.fields.error,
			}
			if got := v.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_capturedError_Unwrap(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "return_wrapped_error",
			err:  errors.New("test error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &capturedError{
				error: tt.err,
			}
			if err := v.Unwrap(); err != tt.err {
				t.Errorf("Unwrap() error = %v, wantErr %v", err, tt.err)
			}
		})
	}
}
