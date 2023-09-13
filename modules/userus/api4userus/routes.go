package api4userus

import (
	"github.com/sneat-co/sneat-go-core/modules"
	"net/http"
)

// RegisterUserRoutes initiates users module
func RegisterUserRoutes(handle modules.HTTPHandleFunc) {
	handle(http.MethodPost, "/v0/users/init_user_record", httpInitUserRecord)
	handle(http.MethodPost, "/v0/users/set_user_country", httpSetUserCountry)
	//handle(http.MethodPost, "/v0/users/create_user", httpPostCreateUser)
}
