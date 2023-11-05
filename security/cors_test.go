package security

import (
	core "github.com/sneat-co/sneat-go-core"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKnownHosts(t *testing.T) {
	assert.Equal(t, []string{"localhost:4200"}, core.KnownHosts)
}

func TestIsSupportedOrigin(t *testing.T) {
	for i, s := range []string{"", "http://localhost:8100"} {
		if !IsSupportedOrigin(s) {
			t.Errorf("string #%d [%v] should be a supported origin", i, s)
		}
	}
}
