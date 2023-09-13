package briefs4memberus

type TeamMemberRole = string

const (
	TeamMemberRoleTeamMember = "team_member"

	TeamMemberRoleAdult = "adult"

	TeamMemberRoleChild = "child"

	// TeamMemberRoleScrumMaster role of a scrum master
	TeamMemberRoleScrumMaster TeamMemberRole = "scrum_master"

	// TeamMemberRoleCreator role of a creator
	TeamMemberRoleCreator TeamMemberRole = "creator"

	// TeamMemberRoleOwner role of an owner
	TeamMemberRoleOwner TeamMemberRole = "owner"

	// TeamMemberRoleContributor role of a contributor
	TeamMemberRoleContributor TeamMemberRole = "contributor"

	// TeamMemberRoleSpectator role of spectator
	TeamMemberRoleSpectator TeamMemberRole = "spectator"

	// TeamMemberRoleExcluded if team members is excluded
	TeamMemberRoleExcluded TeamMemberRole = "excluded"
)

// TeamMemberKnownRoles defines known roles
var TeamMemberKnownRoles = []TeamMemberRole{
	TeamMemberRoleCreator,
	TeamMemberRoleContributor,
	TeamMemberRoleSpectator,
	TeamMemberRoleExcluded,
	TeamMemberRoleScrumMaster,
}

// IsKnownTeamMemberRole checks if role has valid value
func IsKnownTeamMemberRole(role TeamMemberRole, teamRoles []TeamMemberRole) bool {
	for _, r := range TeamMemberKnownRoles {
		if r == role {
			return true
		}
	}
	if teamRoles == nil {
		return true
	}
	for _, r := range teamRoles {
		if r == role {
			return true
		}
	}
	return false
}
