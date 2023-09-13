package facade4userus

import (
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/modules/userus/dto4userus"
	"testing"
)

func TestSetUserTitleRequest_Validate(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		request := dto4userus.InitUserRecordRequest{}
		if err := request.Validate(); err == nil {
			t.Fatal("expected error for empty request")
		}
	})
	t.Run("valid", func(t *testing.T) {
		request := dto4userus.InitUserRecordRequest{
			AuthProvider: "password",
			Email:        "a@example.com",
			Name: &dbmodels.Name{
				Full: "Test User",
			},
		}
		if err := request.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
