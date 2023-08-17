package validate

import "testing"

func TestRequestID(t *testing.T) {
	type args struct {
		id        string
		fieldName string
	}
	type test struct {
		name        string
		args        args
		isValid     bool
		shouldPanic bool
	}
	for _, test := range []test{
		{
			name:        "empty",
			args:        args{},
			shouldPanic: true,
		},
		{
			name:    "spaces",
			args:    args{" ", "id"},
			isValid: false,
		},
		{
			name:    "number",
			args:    args{"123", "id"},
			isValid: true,
		},
		{
			name:    "letters",
			args:    args{"abc", "id"},
			isValid: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("The code did not panic")
					}
				}()
			}
			err := RequestID(test.args.id, test.args.fieldName)
			if !test.isValid && err == nil {
				t.Fatal("Expected error, got nil")
			} else if test.isValid && err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
		})
	}
}

func TestRecordID(t *testing.T) {
	type args struct {
		id string
	}
	type test struct {
		name        string
		args        args
		isValid     bool
		shouldPanic bool
	}
	for _, test := range []test{
		{
			name:    "empty",
			args:    args{},
			isValid: false,
		},
		{
			name:    "spaces",
			args:    args{" "},
			isValid: false,
		},
		{
			name:    "number",
			args:    args{"123"},
			isValid: true,
		},
		{
			name:    "letters",
			args:    args{"abc"},
			isValid: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("The code did not panic")
					}
				}()
			}
			err := RecordID(test.args.id)
			if !test.isValid && err == nil {
				t.Fatal("Expected error, got nil")
			} else if test.isValid && err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
		})
	}
}
