package email2console

import (
	emails2 "github.com/sneat-co/sneat-go/src/core/emails"
	"github.com/strongo/random"
	"log"
)

// NewClient creates a new email client that only logs to dev console
func NewClient() emails2.Client {
	return email2Console{}
}

type email2Console struct {
}

type sent struct {
	messageID string
}

func (v sent) MessageID() string {
	return v.messageID
}

func (v email2Console) Send(email emails2.Email) (emails2.Sent, error) {
	const separatorLine = "\n=============================="
	log.Println("EMAIL")
	log.Println("\tFrom:", email.From)
	log.Println("\tTo:", email.To)
	log.Println("\tSubject:", email.Subject)
	if email.Text != "" {
		log.Println("\tText:", separatorLine, "\n", email.Text, separatorLine)
	}
	if email.HTML != "" {
		log.Println("\tHTML:", separatorLine, "\n", email.HTML, separatorLine)
	}
	return sent{random.ID(7)}, nil
}
