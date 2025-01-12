package email2writer

import (
	"context"
	"fmt"
	"github.com/sneat-co/sneat-go-core/emails"
	"io"
	"strconv"
	"time"
)

// NewClient creates a new email client that only logs to dev console
func NewClient(w func() (io.StringWriter, error)) emails.Client {
	if w == nil {
		panic("non nil string writer factory expected")
	}
	return email2writer{w: w}
}

type email2writer struct {
	w func() (io.StringWriter, error)
}

type sent struct {
	t time.Time
}

func (v sent) MessageID() string {
	return strconv.FormatInt(v.t.UnixNano(), 16)
}

func (v email2writer) Send(_ context.Context, email emails.Email) (emails.Sent, error) {
	w, err := v.w()
	if err != nil {
		return nil, err
	}
	const separatorLine = "===================="
	_, _ = w.WriteString("EMAIL\n")
	_, _ = w.WriteString(fmt.Sprintf("\t   From: %s\n", email.From))
	_, _ = w.WriteString(fmt.Sprintf("\t     To: %s\n", email.To))
	_, _ = w.WriteString(fmt.Sprintf("\tSubject: %s\n", email.Subject))
	_, _ = w.WriteString(fmt.Sprintf("\t   Text:%s\n%s\n%s\n", separatorLine, email.Text, separatorLine))
	_, _ = w.WriteString(fmt.Sprintf("\t   HTML:%s\n%s\n%s\n", separatorLine, email.HTML, separatorLine))
	return sent{t: time.Now()}, nil
}
