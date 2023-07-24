package validate

import (
	"fmt"
	"time"
)

const ISO8601_DATE_LAYOUT = "2006-01-02"

// ValidDateString checks if a string is in valid ISO "YYYY-MM-DD" format
func ValidDateString(s string) (date time.Time, err error) {
	if len(s) != 10 {
		return date, fmt.Errorf("date field should be 10 characters long in YYYY-MM-DD format, got string of %v chars: [%v]", len(s), s)
	}
	if date, err = time.Parse(ISO8601_DATE_LAYOUT, s); err != nil {
		return date, err
	}
	return date, nil
}
