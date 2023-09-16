package briefs4memberus

import (
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/modules/contactus/briefs4contactus"
	"testing"
)

func TestMemberBase_Validate(t *testing.T) {
	type fields struct {
		ContactBase briefs4contactus.ContactBase
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "empty", fields: fields{}, wantErr: true,
		},
		{
			name: "should_pass", fields: fields{
				ContactBase: briefs4contactus.ContactBase{
					ContactBrief: briefs4contactus.ContactBrief{
						Gender:   dbmodels.GenderUnknown,
						Type:     briefs4contactus.ContactTypePerson,
						Title:    "test_title",
						AgeGroup: "unknown",
						WithUserID: dbmodels.WithUserID{
							UserID: "test_user_id",
						},
						WithRoles: dbmodels.WithRoles{
							Roles: []string{"contributor"},
						},
					},
					Status: briefs4contactus.ContactStatusActive,
				},
			}, wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := MemberBase{
				ContactBase: tt.fields.ContactBase,
			}
			if err := v.Validate(); (err != nil) != tt.wantErr {
				if tt.wantErr {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				} else {
					t.Errorf("Validate() unexpected error(s):\n%v", err)
				}
			}
		})
	}
}
