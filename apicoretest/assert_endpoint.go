package apicoretest

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type AssertOptions struct {
	AuthRequired bool
}

func TestEndpoint(
	t *testing.T,
	f func(w http.ResponseWriter, r *http.Request),
	o AssertOptions,
	r *http.Request,
) {
	if o.AuthRequired {
		assertAuthRequired(t, f, r)
	}
}

func assertAuthRequired(
	t *testing.T,
	f func(w http.ResponseWriter, r *http.Request),
	r *http.Request,
) {
	t.Run(fmt.Sprintf("StatusUnauthorized:%s:%s", r.Method, r.URL.Path), func(t *testing.T) {
		// Arrange
		w := httptest.NewRecorder()
		// Act
		f(w, r)
		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
