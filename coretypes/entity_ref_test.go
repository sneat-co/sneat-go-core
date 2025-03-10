package coretypes

import (
	"testing"
)

func TestEntityRef_ID(t *testing.T) {
	tests := []struct {
		name string
		v    EntityRef
		want string
	}{
		{
			name: "empty",
			v:    "",
			want: "",
		},
		{
			name: "without_space_id",
			v:    "contact:123",
			want: "123",
		},
		{
			name: "with_space_id",
			v:    "contact:123@space12",
			want: "123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.ID(); got != tt.want {
				t.Errorf("userID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntityRef_Kind(t *testing.T) {
	tests := []struct {
		name string
		v    EntityRef
		want string
	}{
		{
			name: "empty",
			v:    "",
			want: "",
		},
		{
			name: "without_space_id",
			v:    "contact:123",
			want: "contact",
		},
		{
			name: "with_space_id",
			v:    "contact:123@space1",
			want: "contact",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.Kind(); got != tt.want {
				t.Errorf("Kind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntityRef_SpaceID(t *testing.T) {
	tests := []struct {
		name string
		v    EntityRef
		want SpaceID
	}{
		{
			name: "empty",
			v:    "",
			want: "",
		},
		{
			name: "without_space_id",
			v:    "contact:123",
			want: "",
		},
		{
			name: "with_space_id",
			v:    "contact:123@space1",
			want: "space1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.SpaceID(); got != tt.want {
				t.Errorf("SpaceID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntityRef_Validate(t *testing.T) {
	tests := []struct {
		name    string
		v       EntityRef
		wantErr bool
	}{
		{
			name:    "empty",
			v:       "",
			wantErr: true,
		},
		{
			name:    "without_space_id",
			v:       "contact:123",
			wantErr: false,
		},
		{
			name:    "with_space_id",
			v:       "contact:123@space1",
			wantErr: false,
		},
		{
			name:    "empty_kind_without_space_id",
			v:       ":123",
			wantErr: true,
		},
		{
			name:    "empty_kind_with_space_id",
			v:       ":123@space1",
			wantErr: true,
		},
		{
			name:    "empty_id_without_space_id",
			v:       "contact:",
			wantErr: true,
		},
		{
			name:    "empty_id_with_space_id",
			v:       "contact:@space1",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.v.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewEntityRef(t *testing.T) {
	type args struct {
		kind string
		id   string
	}
	tests := []struct {
		name string
		args args
		want EntityRef
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewEntityRef(tt.args.kind, tt.args.id); got != tt.want {
				t.Errorf("NewEntityRef() = %v, want %v", got, tt.want)
			}
		})
	}
}
