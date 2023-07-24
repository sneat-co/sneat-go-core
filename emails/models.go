package emails

// Email record
type Email struct {
	From    string   `json:"from" firestore:"from"`                           // From source email
	To      []string `json:"to" firestore:"to"`                               // To destination email(s)
	Subject string   `json:"subject" firestore:"subject"`                     // Subject text to send
	Text    string   `json:"text,omitempty" firestore:"text,omitempty"`       // Text is the text body representation
	HTML    string   `json:"html,omitempty" firestore:"html,omitempty"`       // HTMLBody is the HTML body representation
	ReplyTo []string `json:"replyTo,omitempty" firestore:"replyTo,omitempty"` // Reply-To email(s)
}
