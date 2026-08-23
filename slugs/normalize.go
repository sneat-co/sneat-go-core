package slugs

import (
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"
)

// lowerCaser performs Unicode-aware lowercasing rather than strings.ToLower,
// which is ASCII-correct but not guaranteed correct for every script. It is
// package-level and safe for concurrent use: cases.Caser values carry no
// mutable state of their own.
var lowerCaser = cases.Lower(language.Und)

// Normalize puts a raw, human-typed slug into the exact form this package
// stores and claims, per REQ:normalisation. Every exported function that
// accepts a raw slug (Claim, Resolve, Release, Rename) calls this
// internally, so calling it yourself is only necessary to preview the result
// to a user before they commit to it — e.g. showing "your venue will be at
// /venue/st-marys-hall" as they type "St Mary's Hall".
//
// The steps, each of which is load-bearing for a specific failure mode
// described in the Feature's Problem section:
//
//  1. Unicode NFC normalisation, so two different byte sequences that render
//     identically (a precomposed "é" vs. "e" + a combining acute accent)
//     collapse to the same string before anything else looks at it. Without
//     this step two claims that are visually identical could both succeed.
//  2. Unicode-aware lowercasing, so "St-Marys" and "st-marys" are the same
//     slug rather than a private collision bug or a public phishing vector.
//  3. Every rune that is not a letter or digit — spaces, punctuation, runs of
//     hyphens, anything else — becomes a single separating hyphen, and runs
//     of separators collapse to one. This is what makes
//     "St--Marys Village Hall " and "st-marys-village-hall" the same slug.
//  4. Leading and trailing hyphens are trimmed, since a separator run at
//     either end has nothing to separate.
//
// Normalize does not strip or fold non-Latin letters, and does not attempt
// confusable/homoglyph detection — see the Feature's "Not Doing".
func Normalize(s string) string {
	s = norm.NFC.String(s)
	s = lowerCaser.String(s)

	var b strings.Builder
	b.Grow(len(s))
	lastWasSeparator := true // suppresses a leading hyphen the same way the loop suppresses runs
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastWasSeparator = false
			continue
		}
		if !lastWasSeparator {
			b.WriteByte('-')
			lastWasSeparator = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}
