package dbmodels

type Status = string

const (
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
	StatusDeleted  Status = "deleted"
	StatusDraft    Status = "draft"
)
