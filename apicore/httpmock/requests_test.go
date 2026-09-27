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

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on unmarshalable type")
		}
	}()
	NewPostJSONRequest(http.MethodPost, "https://target", make(chan int))
}
