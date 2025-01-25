package dbprofile

import (
	"strings"
	"testing"
)

func TestAvatar_Validate(t *testing.T) {
	tests := []struct {
		name    string
		avatar  *Avatar
		wantErr string
	}{
		{
			name:    "empty",
			avatar:  &Avatar{},
			wantErr: "url|fileID|fileUniqueID",
		},
		{
			name: "with_url_and_provider",
			avatar: &Avatar{
				Provider: "provider1",
				URL:      "https://example.com/avatar.jpg",
			},
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.avatar.Validate()
			if err == nil && tt.wantErr == "" {
				return
			}
			if err == nil && tt.wantErr != "" {
				t.Errorf("Validate() returned no error, expected %v", tt.wantErr)
			}
			if err != nil && tt.wantErr == "" {
				t.Errorf("Validate() returned error when no error has been expected %v", err)
			}
			if err != nil && !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Validate() returned unexpected error %v, expected: %s", err, tt.wantErr)
			}
		})
	}
}

func TestAvatar_Equal(t *testing.T) {
	tests := []struct {
		name    string
		avatar1 *Avatar
		avatar2 *Avatar
		want    bool
	}{
		{
			name:    "nil",
			avatar1: nil,
			avatar2: nil,
			want:    true,
		},
		{
			name:    "empty",
			avatar1: &Avatar{},
			avatar2: &Avatar{},
			want:    true,
		},
		{
			name: "full",
			avatar1: &Avatar{
				Provider: "provider",
				URL:      "url",
				Width:    100,
				Height:   200,
				Size:     300,
			},
			avatar2: &Avatar{
				Provider: "provider",
				URL:      "url",
				Width:    100,
				Height:   200,
				Size:     300,
			},
			want: true,
		},
		{
			name: "provider",
			avatar1: &Avatar{
				Provider: "provider1",
				URL:      "url",
				Width:    100,
				Height:   200,
				Size:     300,
			},
			avatar2: &Avatar{
				Provider: "provider2",
				URL:      "url",
				Width:    100,
				Height:   200,
				Size:     300,
			},
			want: false,
		},
		{
			name: "url",
			avatar1: &Avatar{
				Provider: "provider",
				URL:      "url1",
				Width:    100,
				Height:   200,
				Size:     300,
			},
			avatar2: &Avatar{
				Provider: "provider",
				URL:      "url2",
				Width:    100,
				Height:   200,
				Size:     300,
			},
			want: false,
		},
		{
			name: "width",
			avatar1: &Avatar{
				Provider: "provider",
				URL:      "url",
				Width:    101,
				Height:   200,
				Size:     300,
			},
			avatar2: &Avatar{
				Provider: "provider",
				URL:      "url",
				Width:    102,
				Height:   200,
				Size:     300,
			},
			want: false,
		},
		{
			name: "height",
			avatar1: &Avatar{
				Provider: "provider",
				URL:      "url",
				Width:    100,
				Height:   201,
				Size:     300,
			},
			avatar2: &Avatar{
				Provider: "provider",
				URL:      "url",
				Width:    100,
				Height:   202,
				Size:     300,
			},
			want: false,
		},
		{
			name: "size",
			avatar1: &Avatar{
				Provider: "provider",
				URL:      "url",
				Width:    100,
				Height:   200,
				Size:     301,
			},
			avatar2: &Avatar{
				Provider: "provider",
				URL:      "url",
				Width:    100,
				Height:   200,
				Size:     302,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.avatar1.Equal(tt.avatar2); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}
