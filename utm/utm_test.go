package utm

import "testing"

func TestUtm(t *testing.T) {
	var utm Params
	if !utm.IsEmpty() {
		t.Error("utm.IsEmpty() should return true")
	}
	utm = Params{
		Source:   "google",
		Medium:   "telegram",
		Campaign: "receipt",
	}
	if utm.IsEmpty() {
		t.Error("utm.IsEmpty() should return false")
	}
	if utm.String() != "utm_source=google&utm_medium=telegram&utm_campaign=receipt" {
		t.Errorf("utm.String() should return 'utm_source=google&utm_medium=telegram&utm_campaign=receipt', but it returns '%v'", utm.String())
	}
	if utm.ShortString() != "utm=google;telegram;receipt" {
		t.Errorf("utm.ShortString() should return 'utm=google;telegram;receipt', but it returns '%v'", utm.ShortString())
	}
}

func TestUtmString_panics(t *testing.T) {
	testPanics(t, func(utm Params) string { return utm.String() })
}

func TestUtmShortString_panics(t *testing.T) {
	testPanics(t, func(utm Params) string { return utm.ShortString() })
}

func testPanics(t *testing.T, f func(p Params) string) {
	t.Run("utm.Source is not provided", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("utm.ShortString() should panic if utm.Source is not provided")
			}
		}()
		utm := Params{
			Medium:   "telegram",
			Campaign: "receipt",
		}
		_ = f(utm)
	})
	t.Run("utm.Medium is not provided", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("utm.ShortString() should panic if utm.Medium is not provided")
			}
		}()
		utm := Params{
			Source:   "google",
			Campaign: "receipt",
		}
		_ = f(utm)
	})
	t.Run("utm.Campaign is not provided", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("utm.ShortString() should panic if utm.Campaign is not provided")
			}
		}()
		utm := Params{
			Source: "google",
			Medium: "telegram",
		}
		_ = f(utm)
	})
}
