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
			t.Fatalf("Expected nil error for valid Iana and empty OffsetMinutes, got: %v", err)
		}
	})

	t.Run("valid_iana_valid_utcoffset_should_pass", func(t *testing.T) {
		v := Timezone{
			Iana:          "America/New_York",
			OffsetMinutes: -300, // -5 hours = -300 minutes
		}
		if err := v.Validate(); err != nil {
			t.Fatalf("Expected nil error for valid Iana and valid OffsetMinutes, got: %v", err)
		}
	})

	t.Run("valid_iana_invalid_utcoffset_should_fail", func(t *testing.T) {
		v := Timezone{
			Iana:          "America/New_York",
			OffsetMinutes: 7, // Not divisible by 15
		}
		if err := v.Validate(); err == nil {
			t.Fatal("Expected to get error for invalid OffsetMinutes")
		}
	})

	// Test cases for UTC and GMT as valid Iana values
	t.Run("utc_as_valid_iana_should_pass", func(t *testing.T) {
		v := Timezone{
			Iana:          "UTC",
			OffsetMinutes: 0,
		}
		if err := v.Validate(); err != nil {
			t.Fatalf("Expected nil error for UTC as Iana, got: %v", err)
		}
	})

	t.Run("gmt_as_valid_iana_should_pass", func(t *testing.T) {
		v := Timezone{
			Iana:          "GMT",
			OffsetMinutes: 0,
		}
		if err := v.Validate(); err != nil {
			t.Fatalf("Expected nil error for GMT as Iana, got: %v", err)
		}
	})

	t.Run("invalid_iana_without_slash_should_fail", func(t *testing.T) {
		v := Timezone{
			Iana:          "InvalidTimezone", // Not UTC or GMT and no slash
			OffsetMinutes: 0,
		}
		if err := v.Validate(); err == nil {
			t.Fatal("Expected to get error for invalid Iana without slash")
		}
	})

	// Test cases for number of slashes in Iana field
	t.Run("iana_with_one_slash_should_pass", func(t *testing.T) {
		v := Timezone{
			Iana:          "America/New_York", // One slash
			OffsetMinutes: 0,
		}
		if err := v.Validate(); err != nil {
			t.Fatalf("Expected nil error for Iana with one slash, got: %v", err)
		}
	})

	t.Run("iana_with_multiple_slashes_should_fail", func(t *testing.T) {
		v := Timezone{
			Iana:          "America/North/New_York", // Two slashes
			OffsetMinutes: 0,
		}
		if err := v.Validate(); err == nil {
			t.Fatal("Expected to get error for Iana with multiple slashes")
		}
	})

	// Test cases for OffsetMinutes being an increment of 15
	t.Run("offset_minutes_divisible_by_15_should_pass", func(t *testing.T) {
		testCases := []int{0, 15, 30, 45, 60, -15, -30, -45, -60}
		for _, offset := range testCases {
			v := Timezone{
				Iana:          "America/New_York",
				OffsetMinutes: offset,
			}
			if err := v.Validate(); err != nil {
				t.Fatalf("Expected nil error for OffsetMinutes %d, got: %v", offset, err)
			}
		}
	})

	t.Run("offset_minutes_not_divisible_by_15_should_fail", func(t *testing.T) {
		testCases := []int{1, 7, 14, 16, 29, -1, -7, -14, -16, -29}
		for _, offset := range testCases {
			v := Timezone{
				Iana:          "America/New_York",
				OffsetMinutes: offset,
			}
			if err := v.Validate(); err == nil {
				t.Fatalf("Expected error for OffsetMinutes %d, but got nil", offset)
			}
		}
	})
}

func TestWithTimezone_SetTimezone(t *testing.T) {
	t.Run("nil_timezone", func(t *testing.T) {
		v := &WithTimezone{}
		updates := v.SetTimezone("America/New_York", -300) // -5 hours = -300 minutes

		assert.NotNil(t, v.Timezone, "Timezone should not be nil after setting")
		assert.Equal(t, "America/New_York", v.Timezone.Iana, "Iana should be set correctly")
		assert.Equal(t, -300, v.Timezone.OffsetMinutes, "OffsetMinutes should be set correctly")
		assert.Len(t, updates, 1, "Should return 1 update")

		expectedTimezone := &Timezone{
			Iana:          "America/New_York",
			OffsetMinutes: -300,
		}
		expectedUpdates := []update.Update{update.ByFieldName("timezone", expectedTimezone)}
		assert.Equal(t, expectedUpdates, updates, "Updates should match expected")
	})

	t.Run("existing_timezone_different_values", func(t *testing.T) {
		v := &WithTimezone{
			Timezone: &Timezone{
				Iana:          "Europe/London",
				OffsetMinutes: 0, // +0 hours = 0 minutes
			},
		}
		updates := v.SetTimezone("America/New_York", -300) // -5 hours = -300 minutes

		assert.Equal(t, "America/New_York", v.Timezone.Iana, "Iana should be updated")
		assert.Equal(t, -300, v.Timezone.OffsetMinutes, "OffsetMinutes should be updated")
		assert.Len(t, updates, 1, "Should return 1 update")

		expectedTimezone := &Timezone{
			Iana:          "America/New_York",
			OffsetMinutes: -300,
		}
		expectedUpdates := []update.Update{update.ByFieldName("timezone", expectedTimezone)}
		assert.Equal(t, expectedUpdates, updates, "Updates should match expected")
	})

	t.Run("existing_timezone_same_values", func(t *testing.T) {
		v := &WithTimezone{
			Timezone: &Timezone{
				Iana:          "America/New_York",
				OffsetMinutes: -300, // -5 hours = -300 minutes
			},
		}
		updates := v.SetTimezone("America/New_York", -300)

		assert.Equal(t, "America/New_York", v.Timezone.Iana, "Iana should remain the same")
		assert.Equal(t, -300, v.Timezone.OffsetMinutes, "OffsetMinutes should remain the same")
		assert.Len(t, updates, 0, "Should return no updates when values are the same")
		assert.Nil(t, updates, "Updates should be nil when no changes are made")
	})

	t.Run("change_only_iana", func(t *testing.T) {
		v := &WithTimezone{
			Timezone: &Timezone{
				Iana:          "Europe/London",
				OffsetMinutes: -300, // -5 hours = -300 minutes
			},
		}
		updates := v.SetTimezone("America/New_York", -300)

		assert.Equal(t, "America/New_York", v.Timezone.Iana, "Iana should be updated")
		assert.Equal(t, -300, v.Timezone.OffsetMinutes, "OffsetMinutes should remain the same")
		assert.Len(t, updates, 1, "Should return 1 update")

		expectedTimezone := &Timezone{
			Iana:          "America/New_York",
			OffsetMinutes: -300,
		}
		expectedUpdates := []update.Update{update.ByFieldName("timezone", expectedTimezone)}
		assert.Equal(t, expectedUpdates, updates, "Updates should match expected")
	})

	t.Run("change_only_utcOffset", func(t *testing.T) {
		v := &WithTimezone{
			Timezone: &Timezone{
				Iana:          "America/New_York",
				OffsetMinutes: 0, // +0 hours = 0 minutes
			},
		}
		updates := v.SetTimezone("America/New_York", -300) // -5 hours = -300 minutes

		assert.Equal(t, "America/New_York", v.Timezone.Iana, "Iana should remain the same")
		assert.Equal(t, -300, v.Timezone.OffsetMinutes, "OffsetMinutes should be updated")
		assert.Len(t, updates, 1, "Should return 1 update")

		expectedTimezone := &Timezone{
			Iana:          "America/New_York",
			OffsetMinutes: -300,
		}
		expectedUpdates := []update.Update{update.ByFieldName("timezone", expectedTimezone)}
		assert.Equal(t, expectedUpdates, updates, "Updates should match expected")
	})
}
