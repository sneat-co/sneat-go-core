package dbmodels

import (
	"github.com/dal-go/dalgo/dal"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
		wantUpdates []dal.Update
	}{
		{
			name:        "nil",
			fields:      fields{UserIDs: nil},
			args:        args{uid: "user1"},
			wantUpdates: []dal.Update{{Field: "userIDs", Value: []string{"user1"}}},
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
