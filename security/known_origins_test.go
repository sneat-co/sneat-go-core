package security

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKnownHosts(t *testing.T) {
	assert.Equal(t, []string{"localhost:4200", "local.sneat.ws"}, knownHosts)
}

func TestIsSupportedOrigin(t *testing.T) {
	for i, s := range []string{"", "http://localhost:8100", "https://local.sneat.ws"} {
		if err := VerifyOrigin(s); err != nil {
			t.Errorf("string #%d [%v] should be a supported origin", i, s)
		}
	}
}
