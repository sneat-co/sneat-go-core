package validate

import "testing"

func TestRequestTitle(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "empty title",
			title:     "",
			fieldName: "title",
			wantErr:   true,
		},
		{
			name:      "whitespace title",
			title:     "  ",
			fieldName: "title",
			wantErr:   true,
		},
		{
			name:      "valid title",
			title:     "Test Title",
			fieldName: "title",
			wantErr:   false,
		},
		{
			name:      "title with leading space",
			title:     " Test",
			fieldName: "title",
			wantErr:   true,
		},
		{
			name:      "title with trailing space",
			title:     "Test ",
			fieldName: "title",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RequestTitle(tt.title, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequestTitle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRecordTitle(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "empty title",
			title:     "",
			fieldName: "title",
			wantErr:   true,
		},
		{
			name:      "whitespace title",
			title:     "  ",
			fieldName: "title",
			wantErr:   true,
		},
		{
			name:      "valid title",
			title:     "Test Title",
			fieldName: "title",
			wantErr:   false,
		},
		{
			name:      "title with leading space",
			title:     " Test",
			fieldName: "title",
			wantErr:   true,
		},
		{
			name:      "title with trailing space",
			title:     "Test ",
			fieldName: "title",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RecordTitle(tt.title, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecordTitle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRequiredEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "empty email",
			email:     "",
			fieldName: "email",
			wantErr:   true,
		},
		{
			name:      "whitespace email",
			email:     "  ",
			fieldName: "email",
			wantErr:   true,
		},
		{
			name:      "valid email",
			email:     "test@example.com",
			fieldName: "email",
			wantErr:   false,
		},
		{
			name:      "invalid email",
			email:     "not-an-email",
			fieldName: "email",
			wantErr:   true,
		},
		{
			name:      "email with leading space",
			email:     " test@example.com",
			fieldName: "email",
			wantErr:   true,
		},
		{
			name:      "email with trailing space",
			email:     "test@example.com ",
			fieldName: "email",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RequiredEmail(tt.email, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequiredEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOptionalEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "empty email",
			email:     "",
			fieldName: "email",
			wantErr:   true,
		},
		{
			name:      "valid email",
			email:     "test@example.com",
			fieldName: "email",
			wantErr:   false,
		},
		{
			name:      "invalid email",
			email:     "not-an-email",
			fieldName: "email",
			wantErr:   true,
		},
		{
			name:      "email with leading space",
			email:     " test@example.com",
			fieldName: "email",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := OptionalEmail(tt.email, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("OptionalEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
