package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultKnownHosts(t *testing.T) {
	expected := []string{
		"localhost:4200",
		"local-app.sneat.ws",
	}
	assert.Equal(t, expected, knownHosts)
}

func TestVerifyOrigin(t *testing.T) {
	for _, origin := range []string{
		"",
		"http://localhost:8100",
		"https://localhost:4315",
		"http://listus-app.localhost:4315",
		"https://listus-app.localhost:4315",
		"https://SNEAT-API.LOCALHOST:4300",
		"https://local-app.sneat.ws",
	} {
		t.Run("allows "+origin, func(t *testing.T) {
			if err := VerifyOrigin(origin); err != nil {
				t.Errorf("%q should be a supported origin: %v", origin, err)
			}
		})
	}

	for _, origin := range []string{
		"listus-app.localhost:4315",
		"ftp://listus-app.localhost:4315",
		"https://listus-app.localhost.evil.example",
		"https://listus-app.localhost@evil.example",
		"https://evil-localhost:4315",
		"https://evil.example",
		"https://.localhost:4315",
		"https://two..dots.localhost:4315",
		"https://bad_name.localhost:4315",
		"https://-leading-hyphen.localhost:4315",
		"https://trailing-hyphen-.localhost:4315",
		"https://listus-app.localhost:4315/path",
		"https://listus-app.localhost:4315?query=value",
		"https://listus-app.localhost:0",
		"https://listus-app.localhost:65536",
		"https://listus-app.localhost:",
	} {
		t.Run("rejects "+origin, func(t *testing.T) {
			assert.ErrorIs(t, VerifyOrigin(origin), ErrBadOrigin)
		})
	}
}

func TestIsLocalhostHost(t *testing.T) {
	for _, host := range []string{
		"localhost",
		"localhost:4315",
		"listus-app.localhost",
		"sneat-api.localhost:4300",
		"SNEAT-API.LOCALHOST:4300",
	} {
		t.Run("allows "+host, func(t *testing.T) {
			assert.True(t, IsLocalhostHost(host))
		})
	}

	for _, host := range []string{
		"",
		"localhost.evil.example",
		"listus-app.localhost.evil.example",
		"evil-localhost:4315",
		".localhost:4315",
		"two..dots.localhost:4315",
		"bad_name.localhost:4315",
		"-leading-hyphen.localhost:4315",
		"trailing-hyphen-.localhost:4315",
		"listus-app.localhost:0",
		"listus-app.localhost:65536",
		"listus-app.localhost:not-a-port",
	} {
		t.Run("rejects "+host, func(t *testing.T) {
			assert.False(t, IsLocalhostHost(host))
		})
	}
}
