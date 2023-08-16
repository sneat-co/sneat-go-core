package dbmodels

import "testing"

func TestGenerateIDFromNameOrRandom(t *testing.T) {
	type args struct {
		name        Name
		existingIDs []string
	}
	tests := []struct {
		name    string
		args    args
		wantId  string
		wantErr bool
	}{
		{
			name: "english_first_last",
			args: args{
				name: Name{
					First: "John",
					Last:  "Smith",
				},
				existingIDs: []string{},
			},
			wantErr: false,
			wantId:  "js",
		},
		{
			name: "non_english_first_last",
			args: args{
				name: Name{
					First: "Иван",
					Last:  "Петров",
				},
				existingIDs: []string{},
			},
			wantErr: false,
			wantId:  "ip",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotId, err := GenerateIDFromNameOrRandom(tt.args.name, tt.args.existingIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateIDFromNameOrRandom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotId != tt.wantId {
				t.Errorf("GenerateIDFromNameOrRandom() gotId = %v, want %v", gotId, tt.wantId)
			}
		})
	}
}
