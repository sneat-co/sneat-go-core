package slugs_test

import (
	"testing"

	"github.com/sneat-co/sneat-go-core/slugs"
)

// TestNormalize pins the exact rule REQ:normalisation states: Unicode-
// normalised, lowercased, runs of separators collapsed to one hyphen, none
// leading or trailing. The "St--Marys Village Hall " case is the literal
// example from AC:equivalent-inputs-normalise-together; the combining-accent
// case is what "Unicode-normalised" buys over a naive lowercase-and-hyphenate
// that only handles ASCII.
func TestNormalize(t *testing.T) {
	// precomposedE is "e with acute accent" as the single codepoint U+00E9.
	// decomposedE is the canonically-equivalent decomposition: plain "e"
	// (U+0065) followed by a standalone combining acute accent (U+0301).
	// They are different byte sequences that render identically. Both are
	// spelled with explicit \u escapes rather than a literal accented
	// character, precisely because "two byte sequences that look the same"
	// is the failure mode under test — relying on this source file's own
	// text encoding to carry a literal accented character untouched would
	// undermine the point of the test.
	const precomposedE = "\u00e9" // U+00E9 as a single precomposed codepoint
	const decomposedE = "e\u0301" // "e" (U+0065) + a standalone combining acute accent (U+0301)

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"already_normalised", "st-marys-village-hall", "st-marys-village-hall"},
		{"the_AC_example", "St--Marys Village Hall ", "st-marys-village-hall"},
		{"leading_and_trailing_separators", "  --New Center-- ", "new-center"},
		{"apostrophe_becomes_separator", "St Mary's Hall", "st-mary-s-hall"},
		{"empty_input", "", ""},
		{"only_separators", "   ---   ", ""},
		{"single_word", "Admin", "admin"},
		{
			name:  "precomposed_accent",
			input: "Caf" + precomposedE,
			want:  "caf" + precomposedE,
		},
		{
			// Canonically equivalent to the case above and must normalise to
			// the exact same stored string, or "Café" typed on two different
			// keyboards/input methods could claim two different documents
			// that render identically.
			name:  "decomposed_accent_matches_precomposed",
			input: "Caf" + decomposedE,
			want:  "caf" + precomposedE,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := slugs.Normalize(tt.input); got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
