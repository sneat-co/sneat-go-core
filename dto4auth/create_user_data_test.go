package dto4auth

import (
	"strings"
	"testing"

	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/strongo/strongoapp/appuser"
	"github.com/strongo/validation"
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

	t.Run("known_gender", func(t *testing.T) {
		for _, g := range append([]string{""}, dbmodels.Genders...) {
			d := DataToCreateUser{
				AuthAccount:  validAccount,
				RemoteClient: validRemoteClient,
				Gender:       g,
			}
			if err := d.Validate(); err != nil {
				t.Errorf("gender %q: expected valid data, got error: %v", g, err)
			}
		}
	})

	t.Run("unknown_gender", func(t *testing.T) {
		d := DataToCreateUser{
			AuthAccount:  validAccount,
			RemoteClient: validRemoteClient,
			Gender:       "<script>",
		}
		err := d.Validate()
		if err == nil {
			t.Fatal("expected error for a gender outside dbmodels.Genders")
		}
		if !validation.IsBadRequestError(err) {
			t.Errorf("expected a bad-request error, got %T: %v", err, err)
		}
		if !strings.Contains(err.Error(), "gender") {
			t.Errorf("error does not name the gender field: %v", err)
		}
		if strings.Contains(err.Error(), "<script>") {
			t.Errorf("error echoes the refused value: %v", err)
		}
	})
}
