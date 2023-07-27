package validate

import (
	"testing"
	"time"
)

func TestDateString(t *testing.T) {

	type expected struct {
		error bool
		date  time.Time
	}
	var data = map[string]expected{
		"":           {error: true},
		"2020-12-31": {error: false, date: time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)},
	}
	for s, expects := range data {
		d, err := DateString(s)
		if expects.error && err == nil {
			t.Error("Expected error, got nil")
		}
		if !expects.error && err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !expects.error && d != expects.date {
			t.Errorf("Expected date %v, got %v", expects.date, d)
		}
	}
}
