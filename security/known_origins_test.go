package security

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultKnownHosts(t *testing.T) {
	expected := []string{
		"localhost:4200",
		"local-app.sneat.ws",
	}
	assert.Equal(t, expected, knownHosts)
}

func TestVerifyOrigin(t *testing.T) {
	for i, o := range []string{
		"",
		"http://localhost:8100",
		"https://local-app.sneat.ws",
	} {
		t.Run("host: "+o, func(t *testing.T) {
			if err := VerifyOrigin(o); err != nil {
				t.Errorf("string #%d [%v] should be a supported origin", i, o)
			}
		})
	}
}
