package validate

import "testing"

func TestIsValidateTime(t *testing.T) {

	testCases := []struct {
		Name           string
		Input          string
		ExpectedOutput string
	}{
		{Name: "empty", Input: "", ExpectedOutput: "time field should be 5 characters long in HH:MM format"},
		{Name: "00:00", Input: "00:00", ExpectedOutput: ""},
		{Name: "00:01", Input: "00:01", ExpectedOutput: ""},
		{Name: "12:34", Input: "12:34", ExpectedOutput: ""},
		{Name: "23:59", Input: "23:59", ExpectedOutput: ""},
		{Name: "24:00", Input: "24:00", ExpectedOutput: ""},
		{Name: "34:56", Input: "34:56", ExpectedOutput: invalidTimeNumbers.Error()},
		{Name: "12345", Input: "12345", ExpectedOutput: shouldBeHHMMFormat.Error()},
		{Name: "H1:34", Input: "H1:34", ExpectedOutput: shouldBeHHMMFormat.Error()},
		{Name: "1H:34", Input: "1H:34", ExpectedOutput: shouldBeHHMMFormat.Error()},
		{Name: "12:M4", Input: "12:M4", ExpectedOutput: shouldBeHHMMFormat.Error()},
		{Name: "12:3M", Input: "12:3M", ExpectedOutput: shouldBeHHMMFormat.Error()},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			err := IsValidateTime(test.Input)
			if test.ExpectedOutput == "" && err == nil || test.ExpectedOutput == err.Error() {
				return
			}
			t.Errorf("%v: expected [%v] got [%v]", test.Input, test.ExpectedOutput, err)
		})
	}
}
