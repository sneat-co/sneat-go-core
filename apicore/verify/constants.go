package verify

// KB = 1024 bytes
const KB = 1024

// MinJSONRequestSize - non-empty json can't be less then 2 bytes, e.g. "{}"
const MinJSONRequestSize int64 = 2

// DefaultMaxJSONRequestSize is set as 7 kilobytes what usually should be enough
const DefaultMaxJSONRequestSize int64 = 7 * KB
