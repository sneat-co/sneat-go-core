package verify

var DefaultJsonWithAuthRequired = Request(
	AuthenticationRequired(true),
	MinimumContentLength(MinJSONRequestSize),
	MaximumContentLength(DefaultMaxJSONRequestSize),
)

var DefaultJsonWithNoAuthRequired = Request(
	MinimumContentLength(MinJSONRequestSize),
	MaximumContentLength(DefaultMaxJSONRequestSize),
)
