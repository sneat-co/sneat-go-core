package api4contactus

import (
	"github.com/sneat-co/sneat-go-core/modules"
	"net/http"
)

// RegisterHttpRoutes registers contact routes
func RegisterHttpRoutes(handle modules.HTTPHandleFunc) {
	handle(http.MethodPost, "/v0/contacts/create_contact", httpPostCreateContact)
	handle(http.MethodDelete, "/v0/contacts/delete_contact", httpDeleteContact)
	handle(http.MethodPost, "/v0/contacts/set_contacts_status", httpSetContactStatus)
	handle(http.MethodPost, "/v0/contacts/set_contact_address", httpSetContactAddress)
	handle(http.MethodPost, "/v0/contacts/set_contact_roles", httpSetContactRoles)
	////
	//handle(http.MethodGet, "/v0/team/join_info", api.GetTeamJoinInfo)
	//handle(http.MethodPost, "/v0/team/join_team", api.JoinTeam)
	//handle(http.MethodPost, "/v0/team/refuse_to_join_team", api.RefuseToJoinTeam)
	//handle(http.MethodPost, "/v0/team/leave_team", api.LeaveTeam)
	//handle(http.MethodPost, "/v0/team/create_member", api.AddMember)
	//handle(http.MethodPost, "/v0/team/add_metric", api.AddMetric)
	//handle(http.MethodPost, "/v0/team/remove_member", api.RemoveMember)
	//handle(http.MethodPost, "/v0/team/change_member_role", api.ChangeMemberRole)
	//handle(http.MethodPost, "/v0/team/remove_metrics", api.RemoveMetrics)
	//handle(http.MethodGet, "/v0/team", api.GetTeam)
}
