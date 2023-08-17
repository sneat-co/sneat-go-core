package apicore

import (
	"github.com/sneat-co/sneat-go-core/apicore/httpmock"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExecute(t *testing.T) {
	r := httpmock.NewPostJSONRequest(http.MethodPost, "http://localhost/", "string")

	w := httptest.NewRecorder()
	Execute(w, r, nil, verify.Request(verify.MinimumContentLength(1)), 0, nil, nil)
}
