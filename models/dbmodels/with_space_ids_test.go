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
