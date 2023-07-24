package httpmock

import (
	"net/http"
	"testing"
)

func TestNewPostJsonRequest(t *testing.T) {
	request := NewPostJSONRequest(http.MethodPost, "https://target", nil)
	if request == nil {
		t.Fatal("request == nil")
	}
}
