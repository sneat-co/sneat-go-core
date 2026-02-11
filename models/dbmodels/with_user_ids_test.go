package dbmodels

import (
	"github.com/dal-go/dalgo/update"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithUserIDs_Validate(t *testing.T) {
	tests := []struct {
		name    string
		userIDs []string
		wantErr bool
	}{
		{
			name:    "empty user IDs",
			userIDs: []string{},
			wantErr: true,
		},
		{
			name:    "nil user IDs",
			userIDs: nil,
			wantErr: true,
		},
		{
			name:    "valid single user ID",
			userIDs: []string{"user1"},
			wantErr: false,
		},
		{
			name:    "valid multiple user IDs",
			userIDs: []string{"user1", "user2"},
			wantErr: false,
		},
		{
			name:    "empty string in user IDs",
			userIDs: []string{"user1", ""},
			wantErr: true,
		},
		{
			name:    "whitespace string in user IDs",
			userIDs: []string{"user1", "  "},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &WithUserIDs{UserIDs: tt.userIDs}
			err := v.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("WithUserIDs.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWithUserIDs_AddUserID(t *testing.T) {
	type fields struct {
		UserIDs []string
	}
	type args struct {
		uid string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantUpdates []update.Update
	}{
		{
			name:        "nil",
			fields:      fields{UserIDs: nil},
			args:        args{uid: "user1"},
			wantUpdates: []update.Update{update.ByFieldName("userIDs", []string{"user1"})},
		},
		{
			name:        "existing",
			fields:      fields{UserIDs: []string{"user1"}},
			args:        args{uid: "user1"},
			wantUpdates: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &WithUserIDs{
				UserIDs: tt.fields.UserIDs,
			}
			assert.Equalf(t, tt.wantUpdates, v.AddUserID(tt.args.uid), "AddUserID(%v)", tt.args.uid)
		})
	}
}

func TestWithUserIDs_HasUserID(t *testing.T) {
	type fields struct {
		UserIDs []string
	}
	type args struct {
		uid string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "empty",
			fields: fields{UserIDs: nil},
			args:   args{uid: "user1"},
			want:   false,
		},
		{
			name:   "not found",
			fields: fields{UserIDs: []string{"user2"}},
			args:   args{uid: "user1"},
			want:   false,
		},
		{
			name:   "found",
			fields: fields{UserIDs: []string{"user1"}},
			args:   args{uid: "user1"},
			want:   true,
		},
		{
			name:   "found_2nd_from_3",
			fields: fields{UserIDs: []string{"user1", "user2", "user3"}},
			args:   args{uid: "user2"},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &WithUserIDs{
				UserIDs: tt.fields.UserIDs,
			}
			assert.Equalf(t, tt.want, v.HasUserID(tt.args.uid), "HasUserID(%v)", tt.args.uid)
		})
	}
}
