package slugs

// options holds the per-call settings a caller can attach to Claim and
// Rename via Option. There is deliberately no package-level registry mapping
// a Namespace to its extra reserved words or default claimant: this package
// keeps no mutable state of its own, so nothing here can leak between
// unrelated namespaces or callers, and a caller that always claims into the
// same namespace can just always pass the same options.
type options struct {
	reservedWords []string
	claimedBy     string
}

// Option configures a single call to Claim or Rename.
type Option func(*options)

// WithReservedWords extends the base reserved-word list
// (REQ:reserved-words) for this call only. This is how the list becomes
// "extensible per namespace": the caller who knows what a given namespace's
// own routes look like (e.g. Bookius reserving a word that means nothing to
// a global space-slug namespace) supplies it here, scoped to exactly the
// calls that claim into that namespace.
func WithReservedWords(words ...string) Option {
	return func(o *options) {
		o.reservedWords = append(o.reservedWords, words...)
	}
}

// WithClaimedBy attaches a caller-defined identity to a claim, stored as
// ClaimedBy and returned by Resolve/Enumerate. This package never interprets
// or enforces it — see the Feature's open question on whether only the
// holder should be able to release a claim, which is deliberately left to
// callers for now.
func WithClaimedBy(claimedBy string) Option {
	return func(o *options) {
		o.claimedBy = claimedBy
	}
}

func newOptions(opts []Option) options {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
