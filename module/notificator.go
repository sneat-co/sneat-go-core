package module

type NotificationArgs struct {
	UserID   string `json:"userID"`
	SubjType string `json:"subjectType"`
	SubjID   string `json:"subjectID"`
	EventID  string `json:"eventID"`
}
