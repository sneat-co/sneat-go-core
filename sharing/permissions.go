package sharing

import "github.com/strongo/strongoapp/with"

type Permission = string

type Permissions map[string]with.CreatedFields

const (
	PermittedToView Permission = "view"
	PermittedToEdit Permission = "edit"
)
