package email2writer

import (
	"github.com/sneat-co/sneat-go-core/emails"
	"io"
	"testing"
)

func TestNewClient(t *testing.T) {
	var client emails.Client

	if client = NewClient(func() (io.StringWriter, error) {
		return nil, nil
	}); client == nil {
		t.Errorf("NewClient should return non nil email client")
		return
	}
}
