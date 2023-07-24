package core

import "testing"

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		name string
		v    string
		want bool
	}{
		{name: "empty", v: "", want: false},
		{name: "a", v: "a", want: true},
		{name: "z", v: "z", want: true},
		{name: "A", v: "A", want: true},
		{name: "Z", v: "Z", want: true},
		{name: "0", v: "0", want: true},
		{name: "9", v: "9", want: true},
		{name: "azAZ09", v: "azAZ09", want: true},
		{name: "space_in_middle", v: "a z", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAlphanumericOrUnderscore(tt.v); got != tt.want {
				t.Errorf("IsAlphanumericOrUnderscore() = %v, want %v", got, tt.want)
			}
		})
	}
}
