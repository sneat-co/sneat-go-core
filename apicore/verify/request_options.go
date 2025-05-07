package verify

// RequestOptions - options for request verification
type RequestOptions interface { // TODO: move to sharing Sneat package
	MinimumContentLength() int64
	MaximumContentLength() int64
	AuthenticationRequired() bool
}

type requestOptions struct {
	minContentLength int64
	maxContentLength int64
	authRequired     bool
}

func (r *requestOptions) MinimumContentLength() int64 {
	return r.minContentLength
}

func (r *requestOptions) MaximumContentLength() int64 {
	return r.maxContentLength
}

func (r *requestOptions) AuthenticationRequired() bool {
	return r.authRequired
}

func Request(options ...func(options2 *requestOptions)) RequestOptions {
	o := requestOptions{}
	for _, option := range options {
		option(&o)
	}
	return &o
}

func MinimumContentLength(v int64) func(*requestOptions) {
	return func(o *requestOptions) {
		o.minContentLength = v
	}
}

func MaximumContentLength(v int64) func(*requestOptions) {
	return func(o *requestOptions) {
		o.maxContentLength = v
	}
}

func AuthenticationRequired(v bool) func(*requestOptions) {
	return func(o *requestOptions) {
		o.authRequired = v
	}
}
