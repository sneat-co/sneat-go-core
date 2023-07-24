package httpmock

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

// NewPostJSONRequest creates a new POST JSON request
func NewPostJSONRequest(method, target string, body interface{}) *http.Request {
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	if err := encoder.Encode(body); err != nil {
		panic(err)
	}
	return httptest.NewRequest(method, target, buffer)
}
