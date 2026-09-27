package email2writer

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/sneat-co/sneat-go-core/emails"
)

func TestNewClient(t *testing.T) {
	t.Run("nil_panics", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on nil factory")
			}
		}()
		NewClient(nil)
	})

	t.Run("valid", func(t *testing.T) {
		client := NewClient(func() (io.StringWriter, error) {
			return nil, nil
		})
		if client == nil {
			t.Fatal("NewClient should return non nil email client")
		}
	})
}

func TestEmail2Writer_Send(t *testing.T) {
	t.Run("factory_error", func(t *testing.T) {
		client := NewClient(func() (io.StringWriter, error) {
			return nil, errors.New("factory failed")
		})
		_, err := client.Send(context.Background(), emails.Email{})
		if err == nil {
			t.Fatal("expected error from factory failure")
		}
	})

	t.Run("success", func(t *testing.T) {
		var sb strings.Builder
		client := NewClient(func() (io.StringWriter, error) {
			return &sb, nil
		})
		sentResult, err := client.Send(context.Background(), emails.Email{
			From:    "from@example.com",
			To:      []string{"to@example.com"},
			Subject: "Hello",
			Text:    "Body text",
			HTML:    "<p>Body HTML</p>",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sentResult.MessageID() == "" {
			t.Fatal("expected non-empty message ID")
		}
		if !strings.Contains(sb.String(), "From: from@example.com") {
			t.Fatalf("unexpected output: %s", sb.String())
		}
	})
}
