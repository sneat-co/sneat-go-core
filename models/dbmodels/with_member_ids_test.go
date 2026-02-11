package dbmodels

import "testing"

func TestWithMemberIDs_Validate(t *testing.T) {
	tests := []struct {
		name      string
		memberIDs []string
		wantErr   bool
	}{
		{
			name:      "nil member IDs",
			memberIDs: nil,
			wantErr:   false,
		},
		{
			name:      "empty member IDs",
			memberIDs: []string{},
			wantErr:   false,
		},
		{
			name:      "valid member IDs",
			memberIDs: []string{"member1", "member2"},
			wantErr:   false,
		},
		{
			name:      "empty string in member IDs",
			memberIDs: []string{"member1", ""},
			wantErr:   true,
		},
		{
			name:      "whitespace string in member IDs",
			memberIDs: []string{"member1", "  "},
			wantErr:   true,
		},
		{
			name:      "first member ID is empty",
			memberIDs: []string{""},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := WithMemberIDs{MemberIDs: tt.memberIDs}
			err := v.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("WithMemberIDs.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWithMemberIDs_HasMemberID(t *testing.T) {
	tests := []struct {
		name      string
		memberIDs []string
		checkID   string
		want      bool
	}{
		{
			name:      "empty list",
			memberIDs: []string{},
			checkID:   "member1",
			want:      false,
		},
		{
			name:      "member found",
			memberIDs: []string{"member1", "member2"},
			checkID:   "member1",
			want:      true,
		},
		{
			name:      "member not found",
			memberIDs: []string{"member1", "member2"},
			checkID:   "member3",
			want:      false,
		},
		{
			name:      "member found in middle",
			memberIDs: []string{"member1", "member2", "member3"},
			checkID:   "member2",
			want:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := WithMemberIDs{MemberIDs: tt.memberIDs}
			if got := v.HasMemberID(tt.checkID); got != tt.want {
				t.Errorf("WithMemberIDs.HasMemberID() = %v, want %v", got, tt.want)
			}
		})
	}
}
