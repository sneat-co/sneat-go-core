package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_filter(t *testing.T) {
	type args[T any] struct {
		items []T
		f     func(t T) bool
	}
	type testCase[T any] struct {
		name        string
		args        args[T]
		wantResult  []T
		wantRemoved []T
	}
	tests := []testCase[string]{
		{
			name: "remove_single_item",
			args: args[string]{
				items: []string{"a", "b", "c", "d", "e", "f"},
				f: func(t string) bool {
					return t != "c"
				},
			},
			wantResult:  []string{"a", "b", "d", "e", "f"},
			wantRemoved: []string{"c"},
		},
		{
			name: "remove_2_middle_items",
			args: args[string]{
				items: []string{"a", "b", "c", "d", "e", "f"},
				f: func(t string) bool {
					return t < "c" || t > "d"
				},
			},
			wantResult:  []string{"a", "b", "e", "f"},
			wantRemoved: []string{"c", "d"},
		},
		{
			name: "does_not_remove_anything",
			args: args[string]{
				items: []string{"a", "b", "c", "d", "e", "f"},
				f: func(t string) bool {
					return true
				},
			},
			wantResult:  []string{"a", "b", "c", "d", "e", "f"},
			wantRemoved: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotRemoved := Filter(tt.args.items, tt.args.f)
			assert.Equalf(t, tt.wantResult, gotResult, "Filter(%v, %v)", tt.args.items, tt.args.f)
			assert.Equalf(t, tt.wantRemoved, gotRemoved, "Filter(%v, %v)", tt.args.items, tt.args.f)
		})
	}
}
