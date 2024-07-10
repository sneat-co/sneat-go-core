package dbmodels

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSpaceItemID(t *testing.T) {
	type args struct {
		spaceID string
		id      string
	}
	tests := []struct {
		name string
		args args
		want SpaceItemID
	}{
		{
			name: "ok",
			args: args{
				spaceID: "space1",
				id:      "item1",
			},
			want: "space1_item1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewSpaceItemID(tt.args.spaceID, tt.args.id), "NewSpaceItemID(%v, %v)", tt.args.spaceID, tt.args.id)
		})
	}
}
