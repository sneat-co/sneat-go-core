package verify

// DefaultJsonWithAuthRequired is a request verification that requires authentication and expects JSON content
// with default min = MinJSONRequestSize and max = DefaultMaxJSONRequestSize for content length
var DefaultJsonWithAuthRequired = Request(
	AuthenticationRequired(true),
	MinimumContentLength(MinJSONRequestSize),
	MaximumContentLength(DefaultMaxJSONRequestSize),
)

// DefaultJsonWithNoAuthRequired is a request verification that does not require authentication and expects JSON content
// with default min = MinJSONRequestSize and max = DefaultMaxJSONRequestSize for content length
var DefaultJsonWithNoAuthRequired = Request(
	MinimumContentLength(MinJSONRequestSize),
	MaximumContentLength(DefaultMaxJSONRequestSize),
)

// NoContentAuthRequired is a request verification that requires authentication and expects no content
var NoContentAuthRequired = Request(
	AuthenticationRequired(true),
	MinimumContentLength(0),
	MaximumContentLength(0),
)
