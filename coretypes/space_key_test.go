package coretypes

import (
	"reflect"
	"testing"

	"github.com/dal-go/record"
)

func TestNewSpaceKey(t *testing.T) {
	type args struct {
		id SpaceID
	}
	t.Run("should_panic_on_empty_id", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("no expected panic")
			}
		}()
		NewSpaceKey("")
	})
	t.Run("should_panic_on_long_id", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("no expected panic")
			}
		}()
		NewSpaceKey("0123456789012345678901234567891")
	})
	t.Run("should_panic_on_bare_spot_prefix", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("no expected panic")
			}
		}()
		NewSpaceKey(SpaceID(SpotSpaceIDPrefix))
	})
	t.Run("should_panic_on_unknown_tilde_prefix", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("no expected panic")
			}
		}()
		NewSpaceKey("foo~bar")
	})
	tests := []struct {
		name string
		args args
		want *record.Key
	}{
		{
			name: "should_pass",
			args: args{id: "TestSpace"},
			want: record.NewKeyWithID(spacesCollection, "TestSpace"),
		},
		{
			name: "should_pass_spot_space_id",
			args: args{id: SpotSpaceID("abc123")},
			want: record.NewKeyWithID(spacesCollection, "spot~abc123"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSpaceKey(tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSpaceKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateSpaceID(t *testing.T) {
	for _, spaceID := range []SpaceID{"TestSpace", "family_1", SpotSpaceID("abc123")} {
		if err := ValidateSpaceID(spaceID); err != nil {
			t.Errorf("ValidateSpaceID(%q) = %v", spaceID, err)
		}
	}
	for _, spaceID := range []SpaceID{
		"",
		"family space",
		" family",
		"family@space",
		"0123456789012345678901234567891",
		SpaceID(SpotSpaceIDPrefix),
		"foo~bar",
	} {
		if err := ValidateSpaceID(spaceID); err == nil {
			t.Errorf("ValidateSpaceID(%q) = nil, want error", spaceID)
		}
	}
}

// TestNewSpaceModuleKey verifies the space-bound and spaceless module keys.
// Ported from sneat-core-modules/spaceus/dbo4spaceus's
// space_module_item_test.go.
func TestNewSpaceModuleKey(t *testing.T) {
	if got := NewSpaceModuleKey("space1", "contactus").String(); got != "spaces/space1/ext/contactus" {
		t.Errorf("NewSpaceModuleKey(space1) = %q", got)
	}
	if got := NewSpaceModuleKey("", "contactus").String(); got != "ext/contactus" {
		t.Errorf("NewSpaceModuleKey(empty) = %q", got)
	}
}
