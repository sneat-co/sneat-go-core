package facade

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewErrNoAuthHeader(t *testing.T) {
	type args struct {
		headerName string
	}
	tests := []struct {
		name    string
		args    args
		expects string
	}{
		{"empty", args{""}, "unauthorized: authorization header is not provided"},
		{"with_header_name", args{"BEARER"}, "unauthorized: authorization header is not provided: BEARER"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewErrNoAuthHeader(tt.args.headerName)
			assert.NotNil(t, err)
			assert.True(t, errors.Is(err, ErrUnauthorized))
			assert.Equal(t, tt.expects, err.Error())
		})
	}
}
