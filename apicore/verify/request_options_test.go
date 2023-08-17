package verify

import "testing"

func TestRequest(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		o := Request()
		if *o.(*requestOptions) != (requestOptions{}) {
			t.Error("Request() should return empty RequestOptions")
		}
	})
	t.Run("minContentLength", func(t *testing.T) {
		o := Request(MinimumContentLength(12))
		if o.MinimumContentLength() != 12 {
			t.Error("MinimumContentLength() should return 12")
		}
	})
	t.Run("maxContentLength", func(t *testing.T) {
		o := Request(MaximumContentLength(12))
		if o.MaximumContentLength() != 12 {
			t.Error("MaximumContentLength() should return 12")
		}
	})
}
