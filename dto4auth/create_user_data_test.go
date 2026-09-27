package dto4auth

import (
	"testing"

	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/strongo/strongoapp/appuser"
)

func TestDataToCreateUser_Validate(t *testing.T) {
	validAccount := appuser.AccountKey{
		App:      "app",
		Provider: "google",
		ID:       "user123",
	}
	validRemoteClient := dbmodels.RemoteClientInfo{
		HostOrApp:  "app",
		RemoteAddr: "127.0.0.1",
	}

	t.Run("invalid_auth_account", func(t *testing.T) {
		d := DataToCreateUser{}
		if err := d.Validate(); err == nil {
			t.Fatal("expected error for invalid auth account")
		}
	})

	t.Run("invalid_remote_client", func(t *testing.T) {
		d := DataToCreateUser{
			AuthAccount:  validAccount,
			RemoteClient: dbmodels.RemoteClientInfo{},
		}
		if err := d.Validate(); err == nil {
			t.Fatal("expected error for invalid remote client")
		}
	})

	t.Run("valid", func(t *testing.T) {
		d := DataToCreateUser{
			AuthAccount:  validAccount,
			RemoteClient: validRemoteClient,
		}
		if err := d.Validate(); err != nil {
			t.Fatalf("expected valid data, got error: %v", err)
		}
	})
}
