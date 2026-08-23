package extension

import (
	"strings"
	"testing"

	"github.com/sneat-co/sneat-go-core/coretypes"
)

func venueProfile() SpaceRegistrationProfile {
	return SpaceRegistrationProfile{
		SpaceType:     coretypes.SpaceTypeCompany,
		SlugNamespace: "gametable:space",
	}
}

func TestSpaceRegistrationProfile_Validate(t *testing.T) {
	tests := []struct {
		name    string
		profile SpaceRegistrationProfile
		wantErr string
	}{
		{
			name:    "company is registerable",
			profile: SpaceRegistrationProfile{SpaceType: coretypes.SpaceTypeCompany},
		},
		{
			name:    "club is registerable",
			profile: SpaceRegistrationProfile{SpaceType: coretypes.SpaceTypeClub, SlugNamespace: "sneatclub:space"},
		},
		{
			name:    "creator roles are allowed",
			profile: SpaceRegistrationProfile{SpaceType: coretypes.SpaceTypeGroup, CreatorRoles: []string{"member"}},
		},
		{
			name:    "unknown space type is rejected",
			profile: SpaceRegistrationProfile{SpaceType: "community-center"},
			wantErr: "unknown space type",
		},
		{
			name:    "empty space type is rejected",
			profile: SpaceRegistrationProfile{},
			wantErr: "unknown space type",
		},
		{
			// The platform owns system spaces; an extension must not mint one.
			name:    "system space type is rejected",
			profile: SpaceRegistrationProfile{SpaceType: coretypes.SpaceTypeSystem},
			wantErr: "cannot be registered by an extension",
		},
		{
			// A spot has no members, so there is no operator to register it.
			name:    "spot space type is rejected",
			profile: SpaceRegistrationProfile{SpaceType: coretypes.SpaceTypeSpot},
			wantErr: "cannot be registered by an extension",
		},
		{
			name:    "personal space type is rejected",
			profile: SpaceRegistrationProfile{SpaceType: coretypes.SpaceTypePersonal},
			wantErr: "cannot be registered by an extension",
		},
		{
			name:    "padded slug namespace is rejected",
			profile: SpaceRegistrationProfile{SpaceType: coretypes.SpaceTypeCompany, SlugNamespace: " gametable:space "},
			wantErr: "leading or trailing whitespace",
		},
		{
			name:    "blank creator role is rejected",
			profile: SpaceRegistrationProfile{SpaceType: coretypes.SpaceTypeCompany, CreatorRoles: []string{"  "}},
			wantErr: "empty creator role",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.profile.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() = nil, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Validate() = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestRegisterSpaceProfile(t *testing.T) {
	t.Run("declared profile is readable from the config and the registry", func(t *testing.T) {
		ResetSpaceProfilesForTest()
		ext := NewExtension("gametable", RegisterSpaceProfile(venueProfile()))

		got := ext.SpaceProfile()
		if got == nil {
			t.Fatal("SpaceProfile() = nil, want the declared profile")
		}
		if got.SpaceType != coretypes.SpaceTypeCompany {
			t.Errorf("SpaceType = %q, want %q", got.SpaceType, coretypes.SpaceTypeCompany)
		}

		fromRegistry, ok := LookupSpaceProfile("gametable")
		if !ok {
			t.Fatal("LookupSpaceProfile() found nothing for a declared extension")
		}
		if fromRegistry.SlugNamespace != "gametable:space" {
			t.Errorf("SlugNamespace = %q, want %q", fromRegistry.SlugNamespace, "gametable:space")
		}
	})

	t.Run("an extension without a profile registers no spaces", func(t *testing.T) {
		ResetSpaceProfilesForTest()
		ext := NewExtension("schoolus")

		if ext.SpaceProfile() != nil {
			t.Error("SpaceProfile() is set for an extension that declared none")
		}
		if _, ok := LookupSpaceProfile("schoolus"); ok {
			t.Error("LookupSpaceProfile() found a profile for an extension that declared none")
		}
	})

	t.Run("redeclaring the same profile is a no-op", func(t *testing.T) {
		// Building an extension twice — as tests and multiple composition roots
		// do — must not be an error.
		ResetSpaceProfilesForTest()
		NewExtension("gametable", RegisterSpaceProfile(venueProfile()))
		NewExtension("gametable", RegisterSpaceProfile(venueProfile()))

		if _, ok := LookupSpaceProfile("gametable"); !ok {
			t.Error("LookupSpaceProfile() lost the profile after a redeclaration")
		}
	})

	t.Run("redeclaring a different profile panics", func(t *testing.T) {
		// Two answers to "what is a venue Space?" is exactly the drift
		// decision 0006 exists to end, so it fails loudly at startup.
		ResetSpaceProfilesForTest()
		NewExtension("gametable", RegisterSpaceProfile(venueProfile()))

		defer func() {
			if recover() == nil {
				t.Error("redeclaring a different profile did not panic")
			}
		}()
		NewExtension("gametable", RegisterSpaceProfile(SpaceRegistrationProfile{
			SpaceType: coretypes.SpaceTypeClub,
		}))
	})

	t.Run("an invalid profile panics", func(t *testing.T) {
		ResetSpaceProfilesForTest()
		defer func() {
			if recover() == nil {
				t.Error("an invalid profile did not panic")
			}
		}()
		NewExtension("bad", RegisterSpaceProfile(SpaceRegistrationProfile{SpaceType: "nonsense"}))
	})

	t.Run("differing creator roles are a conflict", func(t *testing.T) {
		ResetSpaceProfilesForTest()
		NewExtension("x", RegisterSpaceProfile(SpaceRegistrationProfile{
			SpaceType:    coretypes.SpaceTypeCompany,
			CreatorRoles: []string{"member"},
		}))
		defer func() {
			if recover() == nil {
				t.Error("differing creator roles did not panic")
			}
		}()
		NewExtension("x", RegisterSpaceProfile(SpaceRegistrationProfile{
			SpaceType:    coretypes.SpaceTypeCompany,
			CreatorRoles: []string{"member", "owner"},
		}))
	})

	t.Run("a differing creator role value is a conflict", func(t *testing.T) {
		ResetSpaceProfilesForTest()
		NewExtension("y", RegisterSpaceProfile(SpaceRegistrationProfile{
			SpaceType:    coretypes.SpaceTypeCompany,
			CreatorRoles: []string{"member"},
		}))
		defer func() {
			if recover() == nil {
				t.Error("a differing creator role value did not panic")
			}
		}()
		NewExtension("y", RegisterSpaceProfile(SpaceRegistrationProfile{
			SpaceType:    coretypes.SpaceTypeCompany,
			CreatorRoles: []string{"owner"},
		}))
	})
}

func TestRegisterableSpaceTypes(t *testing.T) {
	ResetSpaceProfilesForTest()
	if got := RegisterableSpaceTypes(); len(got) != 0 {
		t.Fatalf("RegisterableSpaceTypes() = %v on an empty registry, want none", got)
	}

	NewExtension("gametable", RegisterSpaceProfile(SpaceRegistrationProfile{SpaceType: coretypes.SpaceTypeCompany}))
	NewExtension("communitycentrum", RegisterSpaceProfile(SpaceRegistrationProfile{SpaceType: coretypes.SpaceTypeCompany}))
	NewExtension("sneatclub", RegisterSpaceProfile(SpaceRegistrationProfile{SpaceType: coretypes.SpaceTypeClub}))

	got := RegisterableSpaceTypes()
	// Two products registering `company` yield one entry, sorted.
	want := []coretypes.SpaceType{coretypes.SpaceTypeClub, coretypes.SpaceTypeCompany}
	if len(got) != len(want) {
		t.Fatalf("RegisterableSpaceTypes() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("RegisterableSpaceTypes() = %v, want %v", got, want)
		}
	}
}
