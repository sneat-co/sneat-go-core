package apicore

import (
	"fmt"
	"github.com/sneat-co/sneat-go-core/facade"
	"net/http"
)

func validateContentLength(r *http.Request, min int64, max int64) error {
	if min > 0 {
		if r.ContentLength < min {
			return fmt.Errorf("%w: 'Content-Length' should be greater then %v, got %v", facade.ErrBadRequest, min, r.ContentLength)
		}
	}
	if max >= 0 && r.ContentLength > max {
		return fmt.Errorf("%w: 'Content-Length' should be less then %v, got %v", facade.ErrBadRequest, max, r.ContentLength)
	}
	return nil
}
