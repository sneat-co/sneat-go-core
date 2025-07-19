package extension

import (
	"github.com/sneat-co/sneat-go-core/facade"
	"net/http"
)

// HTTPHandleFunc handles HTTP requests
type HTTPHandleFunc = func(method, path string, handler http.HandlerFunc)

// HandlerFuncWithUserContext handles HTTP requests with user context
type HandlerFuncWithUserContext func(http.ResponseWriter, *http.Request, facade.UserContext)

// Registerer registers HTTP handlers
type Registerer struct {
	HTTPHandle HTTPHandleFunc
	//HttpRouter          *httprouter.Router
	//WrapHttpHandlerFunc func(handler http.HandlerFunc) http.HandlerFunc
}
