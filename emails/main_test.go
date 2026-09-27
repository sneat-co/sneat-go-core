package emails

import (
	"context"
	"errors"
	"testing"
)

func TestInit(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("panic expected")
		}
	}()
	Init(nil)
}

type mockClient struct{}

func (m mockClient) Send(ctx context.Context, email Email) (Sent, error) {
	return nil, nil
}

func TestSend(t *testing.T) {
	t.Run("should_panic", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("panic expected")
			}
		}()
		_, _ = Send(context.Background(), Email{})
	})
	t.Run("should_pass", func(t *testing.T) {
		Init(mockClient{})
		_, err := Send(context.Background(), Email{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestSendEmailError(t *testing.T) {
	baseErr := errors.New("underlying error")
	err := NewSendEmailError("failed to send", baseErr)
	if err.Error() != "failed to send: underlying error" {
		t.Errorf("unexpected Error(): %v", err.Error())
	}
	var sendErr *SendEmailError
	if !errors.As(err, &sendErr) || sendErr.Unwrap() != baseErr {
		t.Errorf("unexpected Unwrap(): %v", sendErr.Unwrap())
	}
}
