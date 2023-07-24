package validate

import (
	"errors"
)

var (
	shouldBeHHMMFormat = errors.New("should be in HH:MM format")
	invalidTimeNumbers = errors.New("invalid time numbers")
)

// IsValidateTime checks if a string is in valid ISO "23:59" format
func IsValidateTime(s string) error {
	if len(s) != 5 {
		return errors.New("time field should be 5 characters long in HH:MM format")
	}
	if s[2] != ':' {
		return shouldBeHHMMFormat
	}
	if i := s[0] - '0'; i > 9 { // check for `|| i < 0` is not needed as `byte` is never < 0.
		return shouldBeHHMMFormat
	} else if i > 2 {
		return invalidTimeNumbers
	}
	if i := s[1] - '0'; i > 9 { // check for `|| i < 0` is not needed as `byte` is never < 0.
		return shouldBeHHMMFormat
	}
	if i := s[3] - '0'; i > 9 { // check for `|| i < 0` is not needed as `byte` is never < 0.
		return shouldBeHHMMFormat
	}
	if i := s[4] - '0'; i > 9 { // check for `|| i < 0` is not needed as `byte` is never < 0.
		return shouldBeHHMMFormat
	}
	return nil
}
