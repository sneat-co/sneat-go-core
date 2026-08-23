package extension

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/sneat-co/sneat-go-core/coretypes"
)

// SpaceRegistrationProfile declares what registering a Space means for one
// extension. It is the "domain-specific" half of unified space registration:
// the platform owns the operation, each extension owns this declaration.
//
// It is deliberately data, not behaviour. An extension states which kind of
// Space it registers, which public URL namespace it claims slugs in, and which
// roles its registering operator needs — and the platform's RegisterSpace
// executes all of that in one transaction. Nothing here is a callback, so a
// reviewer can answer "what does registering a venue do?" by reading one struct
// literal.
//
// See sneat-specs decision 0006 (unified space registration). The shape follows
// eventius' vertical registry: a declarative row per product, where adding a
// product means adding a row and nothing in the engine changes.
type SpaceRegistrationProfile struct {
	// SpaceType is the Space type this extension registers. It describes how
	// membership works, NOT what the Space is for — the extension id recorded
	// in SpaceDbo.Modules is what says the Space is a venue, a school or a
	// community centre.
	//
	// The rule from decision 0006: staff are members and the public are
	// customers => SpaceTypeCompany. A members' organisation, where membership
	// *is* the relationship, is SpaceTypeClub. Products do not mint their own
	// Space types.
	SpaceType coretypes.SpaceType

	// SlugNamespace is the slugs namespace this extension claims public slugs
	// in (e.g. "communitycentrum:space"). Empty means the extension has no
	// public URL scheme for its Spaces, and RegisterSpace never claims a slug
	// for it.
	SlugNamespace string

	// CreatorRoles are roles the registering user needs in addition to the ones
	// CreateSpace always grants a creator (member, creator, owner,
	// contributor). Usually empty: the default set already covers an operator
	// registering their own Space.
	CreatorRoles []string
}

// Validate reports whether the profile can be registered.
func (p SpaceRegistrationProfile) Validate() error {
	if !coretypes.IsValidSpaceType(p.SpaceType) {
		return fmt.Errorf("space registration profile has an unknown space type: %q", p.SpaceType)
	}
	switch p.SpaceType {
	case coretypes.SpaceTypeSystem, coretypes.SpaceTypeSpot, coretypes.SpaceTypePersonal:
		return fmt.Errorf("space type %q cannot be registered by an extension: it is provisioned by the platform or has no ordinary membership", p.SpaceType)
	}
	if p.SlugNamespace != strings.TrimSpace(p.SlugNamespace) {
		return fmt.Errorf("space registration slug namespace must not have leading or trailing whitespace: %q", p.SlugNamespace)
	}
	for _, role := range p.CreatorRoles {
		if strings.TrimSpace(role) == "" {
			return fmt.Errorf("space registration profile has an empty creator role")
		}
	}
	return nil
}

var (
	spaceProfilesMu sync.RWMutex
	spaceProfiles   = map[coretypes.ExtID]SpaceRegistrationProfile{}
)

// RegisterSpaceProfile declares this extension's space-registration profile.
//
// The profile is recorded in a process-wide registry at extension-construction
// time — before any request is served — so the platform can answer both
// "what does registering for extension X create?" and "which Space types can an
// extension register at all?" without importing any extension.
//
// Re-declaring the same profile for the same extension is a no-op, so
// constructing an extension twice (as tests do) is safe. Declaring a
// *different* profile for an id already registered panics: two answers to
// "what is a venue Space?" is the drift decision 0006 exists to end.
func RegisterSpaceProfile(profile SpaceRegistrationProfile) Option {
	return func(m Config) {
		c := m.(*config)
		if err := profile.Validate(); err != nil {
			panic(fmt.Sprintf("extension %s: %v", c.id, err))
		}
		c.spaceProfile = &profile

		spaceProfilesMu.Lock()
		defer spaceProfilesMu.Unlock()
		if existing, ok := spaceProfiles[c.id]; ok && !existing.equal(profile) {
			panic(fmt.Sprintf(
				"extension %s declares two different space registration profiles: %+v and %+v",
				c.id, existing, profile))
		}
		spaceProfiles[c.id] = profile
	}
}

func (p SpaceRegistrationProfile) equal(other SpaceRegistrationProfile) bool {
	if p.SpaceType != other.SpaceType || p.SlugNamespace != other.SlugNamespace {
		return false
	}
	if len(p.CreatorRoles) != len(other.CreatorRoles) {
		return false
	}
	for i, role := range p.CreatorRoles {
		if role != other.CreatorRoles[i] {
			return false
		}
	}
	return true
}

// LookupSpaceProfile returns the registration profile declared by an extension.
// The second result is false when the extension declared none, which means it
// does not register Spaces of its own — schoolus-style extensions that only
// ever operate inside a Space someone else registered.
func LookupSpaceProfile(extID coretypes.ExtID) (SpaceRegistrationProfile, bool) {
	spaceProfilesMu.RLock()
	defer spaceProfilesMu.RUnlock()
	profile, ok := spaceProfiles[extID]
	return profile, ok
}

// RegisterableSpaceTypes returns every Space type some extension declares it
// registers, sorted.
//
// This is what lets the platform decide "is this an ordinary Space an extension
// may act on?" from the registry instead of a hard-coded switch, so adding a
// product no longer means editing a file that product does not own.
func RegisterableSpaceTypes() []coretypes.SpaceType {
	spaceProfilesMu.RLock()
	defer spaceProfilesMu.RUnlock()

	seen := make(map[coretypes.SpaceType]bool, len(spaceProfiles))
	types := make([]coretypes.SpaceType, 0, len(spaceProfiles))
	for _, profile := range spaceProfiles {
		if seen[profile.SpaceType] {
			continue
		}
		seen[profile.SpaceType] = true
		types = append(types, profile.SpaceType)
	}
	sort.Slice(types, func(i, j int) bool { return types[i] < types[j] })
	return types
}

// ResetSpaceProfilesForTest clears the registry. Tests that assert on registry
// contents call it to stay independent of which extensions another test built.
func ResetSpaceProfilesForTest() {
	spaceProfilesMu.Lock()
	defer spaceProfilesMu.Unlock()
	spaceProfiles = map[coretypes.ExtID]SpaceRegistrationProfile{}
}
