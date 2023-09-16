package api4memberus

import (
	"github.com/sneat-co/sneat-go-core/modules"
	"net/http"
)

// RegisterHttpRoutes registers member related routes
func RegisterHttpRoutes(handle modules.HTTPHandleFunc) {
	handle(http.MethodPost, "/v0/members/leave_team", httpPostLeaveTeam)
	handle(http.MethodPost, "/v0/members/create_member", httpPostCreateMember)
	handle(http.MethodPost, "/v0/members/remove_member", httpPostRemoveMember)
}
