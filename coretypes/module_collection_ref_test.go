package coretypes

import "testing"

func TestModuleCollectionRef_Validate(t *testing.T) {
	type fields struct {
		ModuleID   ModuleID
		Collection string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				ModuleID:   "test-module",
				Collection: "test-collection",
			},
			wantErr: false,
		},
		{
			name: "missing moduleID",
			fields: fields{
				ModuleID:   "",
				Collection: "test-collection",
			},
			wantErr: true,
		},
		{
			name: "missing collection",
			fields: fields{
				ModuleID:   "test-module",
				Collection: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &ModuleCollectionRef{
				ModuleID:   tt.fields.ModuleID,
				Collection: tt.fields.Collection,
			}
			if err := v.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
