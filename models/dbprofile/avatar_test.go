package dbprofile

import "testing"

func TestAvatar_Validate(t *testing.T) {
	avatar := Avatar{}
	t.Run("test_empty", func(t *testing.T) {
		if err := avatar.Validate(); err != nil {
			t.Fatal("error expected")
		}
	})
}
