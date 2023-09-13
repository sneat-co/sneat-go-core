package models4memberus

import "testing"

func TestTeamMember_Validate(t *testing.T) {
	record := MemberDto{}
	t.Run("empty_record", func(t *testing.T) {
		if err := record.Validate(); err == nil {
			t.Fatal("error expected for empty record")
		}
	})
}
