package dbmodels

import "testing"

func TestWithMultiSpaceAssetIDs_Validate(t *testing.T) {
	tests := []struct {
		name     string
		assetIDs []string
		wantErr  bool
	}{
		{
			name:     "empty asset IDs",
			assetIDs: []string{},
			wantErr:  true,
		},
		{
			name:     "first element not asterisk",
			assetIDs: []string{"space1:asset1"},
			wantErr:  true,
		},
		{
			name:     "valid with asterisk only",
			assetIDs: []string{"*"},
			wantErr:  false,
		},
		{
			name:     "valid with asterisk and asset",
			assetIDs: []string{"*", "space1:asset1"},
			wantErr:  false,
		},
		{
			name:     "empty string after asterisk",
			assetIDs: []string{"*", ""},
			wantErr:  true,
		},
		{
			name:     "whitespace string after asterisk",
			assetIDs: []string{"*", "  "},
			wantErr:  true,
		},
		{
			name:     "invalid format - no colon",
			assetIDs: []string{"*", "space1asset1"},
			wantErr:  true,
		},
		{
			name:     "invalid format - empty spaceID",
			assetIDs: []string{"*", ":asset1"},
			wantErr:  true,
		},
		{
			name:     "invalid format - empty assetID",
			assetIDs: []string{"*", "space1:"},
			wantErr:  true,
		},
		{
			name:     "valid multiple assets",
			assetIDs: []string{"*", "space1:asset1", "space2:asset2"},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := WithMultiSpaceAssetIDs{AssetIDs: tt.assetIDs}
			err := v.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("WithMultiSpaceAssetIDs.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWithMultiSpaceAssetIDs_HasAssetID(t *testing.T) {
	tests := []struct {
		name     string
		assetIDs []string
		checkID  string
		want     bool
		panics   bool
	}{
		{
			name:     "empty list",
			assetIDs: []string{},
			checkID:  "space1:asset1",
			want:     false,
		},
		{
			name:     "asset found",
			assetIDs: []string{"*", "space1:asset1", "space2:asset2"},
			checkID:  "space1:asset1",
			want:     true,
		},
		{
			name:     "asset not found",
			assetIDs: []string{"*", "space1:asset1"},
			checkID:  "space2:asset2",
			want:     false,
		},
		{
			name:     "asterisk panics",
			assetIDs: []string{"*"},
			checkID:  "*",
			panics:   true,
		},
		{
			name:     "empty string panics",
			assetIDs: []string{"*"},
			checkID:  "",
			panics:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panics {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("WithMultiSpaceAssetIDs.HasAssetID() did not panic")
					}
				}()
			}
			v := WithMultiSpaceAssetIDs{AssetIDs: tt.assetIDs}
			if got := v.HasAssetID(tt.checkID); got != tt.want {
				t.Errorf("WithMultiSpaceAssetIDs.HasAssetID() = %v, want %v", got, tt.want)
			}
		})
	}
}
