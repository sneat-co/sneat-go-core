package dbmodels

import "testing"

func TestIsKnownRelationship(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"spouse", RelationshipSpouse, true},
		{"partner", RelationshipPartner, true},
		{"child", RelationshipChild, true},
		{"sibling", RelationshipSibling, true},
		{"parent", RelationshipParent, true},
		{"grandparent", RelationshipGrandparent, true},
		{"other", RelationshipOther, true},
		{"unknown", RelationshipUnknown, true},
		{"undisclosed", RelationshipUndisclosed, true},
		{"invalid", "invalid", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsKnownRelationship(tt.value); got != tt.want {
				t.Errorf("IsKnownRelationship(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestIsKnownAgeGroupOrEmpty(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"empty", "", true},
		{"unknown", AgeGroupUnknown, true},
		{"adult", AgeGroupAdult, true},
		{"child", AgeGroupChild, true},
		{"senior", AgeGroupSenior, true},
		{"undisclosed", AgeGroupUndisclosed, true},
		{"invalid", "invalid", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsKnownAgeGroupOrEmpty(tt.value); got != tt.want {
				t.Errorf("IsKnownAgeGroupOrEmpty(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestValidateAgeGroup(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		required bool
		wantErr  bool
	}{
		{"valid adult", AgeGroupAdult, false, false},
		{"valid child", AgeGroupChild, false, false},
		{"empty not required", "", false, false},
		{"empty required", "", true, true},
		{"whitespace required", "  ", true, true},
		{"invalid value", "invalid", false, true},
		{"invalid value required", "invalid", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAgeGroup(tt.value, tt.required)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAgeGroup(%v, %v) error = %v, wantErr %v", tt.value, tt.required, err, tt.wantErr)
			}
		})
	}
}
