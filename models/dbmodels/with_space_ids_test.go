package dbmodels

import (
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithSingleSpaceID(t *testing.T) {
	type args struct {
		spaceID coretypes.SpaceID
	}
	tests := []struct {
		name string
		args args
		want WithSpaceIDs
	}{
		{
			name: "ok",
			args: args{
				spaceID: "space1",
			},
			want: WithSpaceIDs{SpaceIDs: []coretypes.SpaceID{"space1"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, WithSingleSpaceID(tt.args.spaceID), "WithSingleSpaceID(%v)", tt.args.spaceID)
		})
	}
}

func TestWithSpaceIDs_Validate(t *testing.T) {
	tests := []struct {
		name     string
		spaceIDs []coretypes.SpaceID
		wantErr  bool
	}{
		{
			name:     "empty space IDs",
			spaceIDs: []coretypes.SpaceID{},
			wantErr:  true,
		},
		{
			name:     "nil space IDs",
			spaceIDs: nil,
			wantErr:  true,
		},
		{
			name:     "valid single space ID",
			spaceIDs: []coretypes.SpaceID{"space123"},
			wantErr:  false,
		},
		{
			name:     "valid multiple space IDs",
			spaceIDs: []coretypes.SpaceID{"space1", "space2"},
			wantErr:  false,
		},
		{
			name:     "empty string in space IDs",
			spaceIDs: []coretypes.SpaceID{"space1", ""},
			wantErr:  true,
		},
		{
			name:     "whitespace string in space IDs",
			spaceIDs: []coretypes.SpaceID{"space1", "  "},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := WithSpaceIDs{SpaceIDs: tt.spaceIDs}
			err := v.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("WithSpaceIDs.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWithSpaceIDs_JoinSpaceIDs(t *testing.T) {
	tests := []struct {
		name     string
		spaceIDs []coretypes.SpaceID
		sep      string
		want     string
	}{
		{
			name:     "empty list",
			spaceIDs: []coretypes.SpaceID{},
			sep:      ",",
			want:     "",
		},
		{
			name:     "single space ID",
			spaceIDs: []coretypes.SpaceID{"space1"},
			sep:      ",",
			want:     "space1",
		},
		{
			name:     "multiple space IDs with comma",
			spaceIDs: []coretypes.SpaceID{"space1", "space2", "space3"},
			sep:      ",",
			want:     "space1,space2,space3",
		},
		{
			name:     "multiple space IDs with pipe",
			spaceIDs: []coretypes.SpaceID{"space1", "space2"},
			sep:      "|",
			want:     "space1|space2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := WithSpaceIDs{SpaceIDs: tt.spaceIDs}
			if got := v.JoinSpaceIDs(tt.sep); got != tt.want {
				t.Errorf("WithSpaceIDs.JoinSpaceIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}
