package security

import (
	"github.com/stretchr/testify/assert"
	"slices"
	"testing"
)

func TestIsKnownHost(t *testing.T) {
	type args struct {
		host string
	}
	tests := []struct {
		name       string
		knownHosts []string
		args       args
		want       bool
	}{
		{
			name:       "emptyKnownHosts",
			knownHosts: []string{},
			args:       args{host: "localhost"},
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := IsKnownHost(tt.args.host)
			assert.Equalf(t, tt.want, actual, "IsKnownHost(%v), knownHosts: %+v", tt.args.host, tt.knownHosts)
		})
	}
}

func TestAddKnownHosts(t *testing.T) {
	type args struct {
		hosts []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "single",
			args: args{hosts: []string{"host1.example.com"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			knownHostsBackup := knownHosts[:]
			knownOriginsBackup := knownOrigins[:]
			AddKnownHosts(tt.args.hosts...)
			for _, host := range tt.args.hosts {
				assert.Truef(t, slices.Contains(knownHosts, host), "knownHosts should has %s, got: %+v", host, knownHosts)
				httpsOrigin := "https://" + host
				assert.Truef(t, slices.Contains(knownOrigins, httpsOrigin), "knownHosts should has %s, got: %+v", httpsOrigin, knownOrigins)
			}
			knownHosts = knownHostsBackup
			knownOrigins = knownOriginsBackup
		})
	}
}
