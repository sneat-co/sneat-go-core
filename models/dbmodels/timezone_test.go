package dbmodels

import "testing"

func TestTimezone_Validate(t *testing.T) {
	t.Run("should_fail", func(t *testing.T) {
		v := Timezone{}
		if err := v.Validate(); err == nil {
			t.Fatal("Expected to get error for empty timezone record")
		}
	})
}
