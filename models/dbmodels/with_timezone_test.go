package dbmodels

import (
	"github.com/dal-go/dalgo/update"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTimezone_Validate(t *testing.T) {
	t.Run("empty_timezone_should_fail", func(t *testing.T) {
		v := Timezone{}
		if err := v.Validate(); err == nil {
			t.Fatal("Expected to get error for empty timezone record")
		}
	})

	t.Run("nil_timezone_should_pass", func(t *testing.T) {
		var v *Timezone
		if err := v.Validate(); err != nil {
			t.Fatalf("Expected nil error for nil timezone, got: %v", err)
		}
	})

	t.Run("valid_iana_empty_utcoffset_should_pass", func(t *testing.T) {
		v := Timezone{
			Iana: "America/New_York",
		}
		if err := v.Validate(); err != nil {
			t.Fatalf("Expected nil error for valid Iana and empty UtcOffset, got: %v", err)
		}
	})

	t.Run("valid_iana_valid_utcoffset_should_pass", func(t *testing.T) {
		v := Timezone{
			Iana:      "America/New_York",
			UtcOffset: "-05:00",
		}
		if err := v.Validate(); err != nil {
			t.Fatalf("Expected nil error for valid Iana and valid UtcOffset, got: %v", err)
		}
	})

	t.Run("valid_iana_invalid_utcoffset_should_fail", func(t *testing.T) {
		v := Timezone{
			Iana:      "America/New_York",
			UtcOffset: "invalid",
		}
		if err := v.Validate(); err == nil {
			t.Fatal("Expected to get error for invalid UtcOffset")
		}
	})
}

func TestWithTimezone_SetTimezone(t *testing.T) {
	t.Run("nil_timezone", func(t *testing.T) {
		v := &WithTimezone{}
		updates := v.SetTimezone("America/New_York", "-05:00")

		assert.NotNil(t, v.Timezone, "Timezone should not be nil after setting")
		assert.Equal(t, "America/New_York", v.Timezone.Iana, "Iana should be set correctly")
		assert.Equal(t, "-05:00", v.Timezone.UtcOffset, "UtcOffset should be set correctly")
		assert.Len(t, updates, 1, "Should return 1 update")

		expectedTimezone := &Timezone{
			Iana:      "America/New_York",
			UtcOffset: "-05:00",
		}
		expectedUpdates := []update.Update{update.ByFieldName("timezone", expectedTimezone)}
		assert.Equal(t, expectedUpdates, updates, "Updates should match expected")
	})

	t.Run("existing_timezone_different_values", func(t *testing.T) {
		v := &WithTimezone{
			Timezone: &Timezone{
				Iana:      "Europe/London",
				UtcOffset: "+00:00",
			},
		}
		updates := v.SetTimezone("America/New_York", "-05:00")

		assert.Equal(t, "America/New_York", v.Timezone.Iana, "Iana should be updated")
		assert.Equal(t, "-05:00", v.Timezone.UtcOffset, "UtcOffset should be updated")
		assert.Len(t, updates, 1, "Should return 1 update")

		expectedTimezone := &Timezone{
			Iana:      "America/New_York",
			UtcOffset: "-05:00",
		}
		expectedUpdates := []update.Update{update.ByFieldName("timezone", expectedTimezone)}
		assert.Equal(t, expectedUpdates, updates, "Updates should match expected")
	})

	t.Run("existing_timezone_same_values", func(t *testing.T) {
		v := &WithTimezone{
			Timezone: &Timezone{
				Iana:      "America/New_York",
				UtcOffset: "-05:00",
			},
		}
		updates := v.SetTimezone("America/New_York", "-05:00")

		assert.Equal(t, "America/New_York", v.Timezone.Iana, "Iana should remain the same")
		assert.Equal(t, "-05:00", v.Timezone.UtcOffset, "UtcOffset should remain the same")
		assert.Len(t, updates, 0, "Should return no updates when values are the same")
		assert.Nil(t, updates, "Updates should be nil when no changes are made")
	})

	t.Run("change_only_iana", func(t *testing.T) {
		v := &WithTimezone{
			Timezone: &Timezone{
				Iana:      "Europe/London",
				UtcOffset: "-05:00",
			},
		}
		updates := v.SetTimezone("America/New_York", "-05:00")

		assert.Equal(t, "America/New_York", v.Timezone.Iana, "Iana should be updated")
		assert.Equal(t, "-05:00", v.Timezone.UtcOffset, "UtcOffset should remain the same")
		assert.Len(t, updates, 1, "Should return 1 update")

		expectedTimezone := &Timezone{
			Iana:      "America/New_York",
			UtcOffset: "-05:00",
		}
		expectedUpdates := []update.Update{update.ByFieldName("timezone", expectedTimezone)}
		assert.Equal(t, expectedUpdates, updates, "Updates should match expected")
	})

	t.Run("change_only_utcOffset", func(t *testing.T) {
		v := &WithTimezone{
			Timezone: &Timezone{
				Iana:      "America/New_York",
				UtcOffset: "+00:00",
			},
		}
		updates := v.SetTimezone("America/New_York", "-05:00")

		assert.Equal(t, "America/New_York", v.Timezone.Iana, "Iana should remain the same")
		assert.Equal(t, "-05:00", v.Timezone.UtcOffset, "UtcOffset should be updated")
		assert.Len(t, updates, 1, "Should return 1 update")

		expectedTimezone := &Timezone{
			Iana:      "America/New_York",
			UtcOffset: "-05:00",
		}
		expectedUpdates := []update.Update{update.ByFieldName("timezone", expectedTimezone)}
		assert.Equal(t, expectedUpdates, updates, "Updates should match expected")
	})
}
