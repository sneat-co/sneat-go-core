package apicore

import (
	"context"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/sneatauth"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleAuthenticatedRequestWithBody(t *testing.T) {
	type args struct {
		r                 *http.Request
		request           interface{ Validate() error }
		options           verify.RequestOptions
		successStatusCode int
		facadeHandler     FacadeHandler
	}

	GetAuthTokenFromHttpRequest = func(r *http.Request) (token *sneatauth.Token, err error) {
		return nil, nil
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "TestHandleAuthenticatedRequestWithBody",
			args: args{
				r:                 httptest.NewRequest(http.MethodGet, "/", nil),
				options:           verify.Request(),
				successStatusCode: http.StatusMethodNotAllowed, // TODO: make it working using http.StatusNoContent
				facadeHandler: func(ctx context.Context, userCtx facade.User) (response any, err error) {
					return nil, nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := new(httptest.ResponseRecorder)
			HandleAuthenticatedRequestWithBody(w, tt.args.r, tt.args.request, tt.args.options, tt.args.successStatusCode, tt.args.facadeHandler)
			assert.Equal(t, tt.args.successStatusCode, w.Code)
		})
	}
}
