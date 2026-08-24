package coretypes

import (
	"strings"
	"testing"
)

func TestIsValidSpaceType(t *testing.T) {
	type args struct {
		v SpaceType
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"SpaceTypePersonal", args{SpaceTypePersonal}, true},
		{"SpaceTypeGroup", args{SpaceTypeGroup}, true},
		{"private-now-invalid", args{"private"}, false},
		{"SpaceTypeSystem", args{SpaceTypeSystem}, true},
		{"SpaceTypeSpot", args{SpaceTypeSpot}, true},
		{"EmptySpaceType", args{""}, false},
		{"InvalidSpaceType", args{"Foo"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidSpaceType(tt.args.v); got != tt.want {
				t.Errorf("IsValidSpaceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpaceTypeGroupFullRefRoundTrip(t *testing.T) {
	spaceRef := NewSpaceRef(SpaceTypeGroup, "circle-1")

	if got, want := spaceRef, SpaceRef("group!circle-1"); got != want {
		t.Errorf("NewSpaceRef() = %q, want %q", got, want)
	}
	if got, want := spaceRef.SpaceType(), SpaceTypeGroup; got != want {
		t.Errorf("SpaceType() = %q, want %q", got, want)
	}
	if got, want := spaceRef.SpaceID(), SpaceID("circle-1"); got != want {
		t.Errorf("SpaceID() = %q, want %q", got, want)
	}
	if got, want := spaceRef.UrlPath(), "group/circle-1"; got != want {
		t.Errorf("UrlPath() = %q, want %q", got, want)
	}
}

func TestSpotSpaceID(t *testing.T) {
	tests := []struct {
		name   string
		spotID string
		want   SpaceID
	}{
		{"simple", "acme-gym", SpaceID("spot~acme-gym")},
		{"with_digits", "venue42", SpaceID("spot~venue42")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SpotSpaceID(tt.spotID); got != tt.want {
				t.Errorf("SpotSpaceID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSpaceRef(t *testing.T) {
	type args struct {
		spaceType SpaceType
		spaceID   SpaceID
	}
	tests := []struct {
		name string
		args args
		want SpaceRef
	}{
		{"ShouldPass", args{SpaceTypePersonal, "foo"}, "personal!foo"},
		{"EmptySpaceType", args{"", "foo"}, ""},
		{"ShouldPass", args{SpaceTypePersonal, ""}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want == "" {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("NewSpaceRef() did not panic")
					}
				}()
			}
			if got := NewSpaceRef(tt.args.spaceType, tt.args.spaceID); got != tt.want {
				t.Errorf("NewSpaceRef() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpaceRef_SpaceID(t *testing.T) {
	tests := []struct {
		name string
		v    SpaceRef
		want SpaceID
	}{
		{"full", "private!foo", "foo"},
		{"no_spaceType", "!foo", "foo"},
		{"no_spaceID", "private!", ""},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.SpaceID(); got != tt.want {
				t.Errorf("SpaceID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpaceRef_SpaceType(t *testing.T) {
	tests := []struct {
		name string
		v    SpaceRef
		want SpaceType
	}{
		{"full", "personal!foo", "personal"},
		{"no_spaceID", "personal!", "personal"},
		{"no_spaceType", "!foo", ""},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.SpaceType(); got != tt.want {
				t.Errorf("SpaceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpaceRef_UrlPath(t *testing.T) {
	tests := []struct {
		name string
		v    SpaceRef
		want string
	}{
		{"ShouldPass", "personal!foo", "personal/foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.UrlPath(); got != tt.want {
				t.Errorf("UrlPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewWeakSpaceRef(t *testing.T) {
	type args struct {
		spaceType SpaceType
	}
	tests := []struct {
		name        string
		args        args
		want        SpaceRef
		expectPanic []string
	}{
		{
			name: "personal",
			args: args{SpaceTypePersonal},
			want: SpaceRef(SpaceTypePersonal),
		},
		{
			name: "family",
			args: args{SpaceTypeFamily},
			want: SpaceRef(SpaceTypeFamily),
		},
		{
			name:        "empty",
			args:        args{""},
			expectPanic: []string{"family", "personal"},
		},
		{
			name:        "group_requires_full_reference",
			args:        args{SpaceTypeGroup},
			expectPanic: []string{"family", "personal"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.expectPanic) > 0 {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("NewWeakSpaceRef() did not panic")
					} else {
						s := r.(string)
						for _, expected := range tt.expectPanic {
							if !strings.Contains(s, expected) {
								t.Errorf("expected '%s' to be in panic message", expected)
							}
						}
					}

				}()
			}
			if got := NewWeakSpaceRef(tt.args.spaceType); got != tt.want {
				t.Errorf("NewWeakSpaceRef() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWeakSpaceRef(t *testing.T) {
	tests := []struct {
		name     string
		spaceRef SpaceRef
		want     SpaceRef
	}{
		{"family", FamilyWeekSpaceRef, SpaceRef(SpaceTypeFamily)},
		{"personal", PersonalWeekSpaceRef, SpaceRef(SpaceTypePersonal)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.spaceRef != tt.want {
				t.Errorf("WeakSpaceRef() = %v, want %v", tt.spaceRef, tt.want)
			}
		})
	}
}

func TestCanonicalSpaceType(t *testing.T) {
	// The legacy "private" (renamed to "personal") must validate as stored
	// data but never as a request value.
	if got := CanonicalSpaceType("private"); got != SpaceTypePersonal {
		t.Errorf("CanonicalSpaceType(private) = %q, want %q", got, SpaceTypePersonal)
	}
	if !IsValidSpaceType(CanonicalSpaceType("private")) {
		t.Error("stored-data validation must accept legacy private")
	}
	if IsValidSpaceType("private") {
		t.Error("request validation must keep rejecting legacy private")
	}
	// Modern values pass through unchanged.
	if got := CanonicalSpaceType(SpaceTypeClub); got != SpaceTypeClub {
		t.Errorf("CanonicalSpaceType(club) = %q, want %q", got, SpaceTypeClub)
	}
	if got := CanonicalSpaceType("nonsense"); got != "nonsense" {
		t.Errorf("CanonicalSpaceType(nonsense) = %q, want unchanged", got)
	}
}
