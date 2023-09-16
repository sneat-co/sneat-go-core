package models4userus

import (
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/modules/contactus/briefs4contactus"
	"testing"
	"time"
)

func TestUserDto_Validate(t *testing.T) {
	userDto := UserDto{
		ContactBase: briefs4contactus.ContactBase{
			ContactBrief: briefs4contactus.ContactBrief{
				Type:   briefs4contactus.ContactTypePerson,
				Gender: "unknown",
				Name: &dbmodels.Name{
					First: "Firstname",
					Last:  "Lastname",
				},
				AgeGroup: "unknown",
			},
			Status: "active",
		},
		Created: dbmodels.CreatedInfo{
			At: time.Now(),
			Client: dbmodels.RemoteClientInfo{
				HostOrApp:  "unit-test",
				RemoteAddr: "127.0.0.1",
			},
		},
	}
	userDto.CountryID = "--"
	t.Run("empty_record", func(t *testing.T) {
		if err := userDto.Validate(); err != nil {
			t.Fatalf("no error expected, got: %v", err)
		}
	})
}
