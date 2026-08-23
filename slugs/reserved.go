package slugs

// baseReservedWords is the fleet-wide floor: slugs no namespace may ever
// claim, because a route or a well-known path is more likely to collide with
// one of these than a real-world name is. REQ:reserved-words names this exact
// list as the minimum; a caller extends it per namespace via
// WithReservedWords rather than this package trying to guess every
// namespace's own routes (e.g. Bookius will want its own words that mean
// nothing here).
//
// Each entry is stored already normalised, since isReserved compares against
// an already-normalised candidate slug.
var baseReservedWords = map[string]struct{}{
	"new":        {},
	"edit":       {},
	"admin":      {},
	"api":        {},
	"static":     {},
	"assets":     {},
	"robots":     {},
	"sitemap":    {},
	"favicon":    {},
	"well-known": {},
}

// isReserved reports whether normalisedSlug — which must already have been
// through Normalize — is on the base reserved list or on extra, the
// per-namespace additions a caller supplied via WithReservedWords. extra is
// normalised here rather than requiring the caller to pre-normalise it, so
// `WithReservedWords("Default")` and `WithReservedWords("default")` behave
// identically.
func isReserved(normalisedSlug string, extra []string) bool {
	if _, ok := baseReservedWords[normalisedSlug]; ok {
		return true
	}
	for _, word := range extra {
		if Normalize(word) == normalisedSlug {
			return true
		}
	}
	return false
}
