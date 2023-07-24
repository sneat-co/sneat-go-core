package security

import "testing"

func TestIsSupportedOrigin(t *testing.T) {
	for i, s := range []string{"", "http://localhost:8100", "https://dailyscrum.app"} {
		if !IsSupportedOrigin(s) {
			t.Errorf("string #%d [%v] should be a supported origin", i, s)
		}
	}
}
