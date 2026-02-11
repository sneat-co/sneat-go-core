package dbmodels

import "testing"

func TestWithOptionalRelatedAs_GetRelatedAs(t *testing.T) {
	v := WithOptionalRelatedAs{RelatedAs: "parent"}
	if got := v.GetRelatedAs(); got != "parent" {
		t.Errorf("GetRelatedAs() = %v, want %v", got, "parent")
	}
}

func TestWithOptionalRelatedAs_Equal(t *testing.T) {
	tests := []struct {
		name string
		v1   WithOptionalRelatedAs
		v2   WithOptionalRelatedAs
		want bool
	}{
		{
			name: "both empty",
			v1:   WithOptionalRelatedAs{},
			v2:   WithOptionalRelatedAs{},
			want: true,
		},
		{
			name: "equal values",
			v1:   WithOptionalRelatedAs{RelatedAs: "parent"},
			v2:   WithOptionalRelatedAs{RelatedAs: "parent"},
			want: true,
		},
		{
			name: "different values",
			v1:   WithOptionalRelatedAs{RelatedAs: "parent"},
			v2:   WithOptionalRelatedAs{RelatedAs: "child"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v1.Equal(tt.v2); got != tt.want {
				t.Errorf("WithOptionalRelatedAs.Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithOptionalRelatedAs_Validate(t *testing.T) {
	tests := []struct {
		name      string
		relatedAs string
		wantErr   bool
	}{
		{
			name:      "empty related as",
			relatedAs: "",
			wantErr:   false,
		},
		{
			name:      "valid related as",
			relatedAs: "parent",
			wantErr:   false,
		},
		{
			name:      "leading space",
			relatedAs: " parent",
			wantErr:   true,
		},
		{
			name:      "trailing space",
			relatedAs: "parent ",
			wantErr:   true,
		},
		{
			name:      "leading and trailing spaces",
			relatedAs: " parent ",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := WithOptionalRelatedAs{RelatedAs: tt.relatedAs}
			err := v.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("WithOptionalRelatedAs.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWithRequiredRelatedAs_Validate(t *testing.T) {
	tests := []struct {
		name      string
		relatedAs string
		wantErr   bool
	}{
		{
			name:      "empty related as",
			relatedAs: "",
			wantErr:   true,
		},
		{
			name:      "valid related as",
			relatedAs: "parent",
			wantErr:   false,
		},
		{
			name:      "leading space",
			relatedAs: " parent",
			wantErr:   true,
		},
		{
			name:      "trailing space",
			relatedAs: "parent ",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := WithRequiredRelatedAs{
				WithOptionalRelatedAs: WithOptionalRelatedAs{RelatedAs: tt.relatedAs},
			}
			err := v.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("WithRequiredRelatedAs.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWithRequiredRelatedAs_Equal(t *testing.T) {
	tests := []struct {
		name string
		v1   WithRequiredRelatedAs
		v2   WithRequiredRelatedAs
		want bool
	}{
		{
			name: "equal values",
			v1: WithRequiredRelatedAs{
				WithOptionalRelatedAs: WithOptionalRelatedAs{RelatedAs: "parent"},
			},
			v2: WithRequiredRelatedAs{
				WithOptionalRelatedAs: WithOptionalRelatedAs{RelatedAs: "parent"},
			},
			want: true,
		},
		{
			name: "different values",
			v1: WithRequiredRelatedAs{
				WithOptionalRelatedAs: WithOptionalRelatedAs{RelatedAs: "parent"},
			},
			v2: WithRequiredRelatedAs{
				WithOptionalRelatedAs: WithOptionalRelatedAs{RelatedAs: "child"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v1.Equal(tt.v2); got != tt.want {
				t.Errorf("WithRequiredRelatedAs.Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}
