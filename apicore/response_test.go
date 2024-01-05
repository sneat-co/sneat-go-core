package apicore

import (
	"context"
	"errors"
	"github.com/sneat-co/sneat-go-core/httpserver"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestReturnError(t *testing.T) {
	type args struct {
		in0 context.Context
		w   http.ResponseWriter
		r   *http.Request
		err error
	}
	tests := []struct {
		name        string
		args        args
		shouldPanic bool
	}{
		{
			name: "calls_httpserver_HandleError",
			args: args{
				err: errors.New("error 1"),
			},
		},
		{
			name:        "panics_on_nil_err",
			shouldPanic: true,
			args: args{
				err: nil,
			},
		},
	}
	var handleErrorBackup = httpserver.HandleError

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handleErrorCalled bool
			if tt.shouldPanic {
				defer func() {
					r := recover()
					assert.NotNil(t, r)
				}()
			} else {
				httpserver.HandleError = func(ctx context.Context, err error, from string, w http.ResponseWriter, r *http.Request) {
					handleErrorCalled = true
					assert.NotNil(t, err)
					assert.True(t, errors.Is(err, tt.args.err))
				}
			}

			ReturnError(tt.args.in0, tt.args.w, tt.args.r, tt.args.err)

			assert.Truef(t, handleErrorCalled, "must call httpserver.HandleError")
		})
		httpserver.HandleError = handleErrorBackup
	}
}
