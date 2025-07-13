package dbmodels

import (
	"github.com/dal-go/dalgo/update"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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
		updates, err := v.SetTimezone("America/New_York")

		assert.NoError(t, err, "Should not return an error for valid timezone")
		assert.NotNil(t, v.Timezone, "Timezone should not be nil after setting")
		assert.Equal(t, "America/New_York", v.Timezone.Iana, "Iana should be set correctly")
		assert.NotZero(t, v.Timezone.OffsetMinutes, "OffsetMinutes should be set")
		assert.Len(t, updates, 1, "Should return 1 update")

		// Check that the update is of the correct type
		expectedUpdate := update.ByFieldName("timezone", v.Timezone)
		assert.Equal(t, expectedUpdate, updates[0], "Update should be of the correct type")
	})

	t.Run("existing_timezone_different_values", func(t *testing.T) {
		v := &WithTimezone{
			Timezone: &Timezone{
				Iana:          "Europe/London",
				OffsetMinutes: 0, // +0 hours = 0 minutes
			},
		}
		updates, err := v.SetTimezone("America/New_York")

		assert.NoError(t, err, "Should not return an error for valid timezone")
		assert.Equal(t, "America/New_York", v.Timezone.Iana, "Iana should be updated")
		assert.NotZero(t, v.Timezone.OffsetMinutes, "OffsetMinutes should be updated")
		assert.Len(t, updates, 1, "Should return 1 update")
	})

	t.Run("existing_timezone_same_values", func(t *testing.T) {
		v := &WithTimezone{
			Timezone: &Timezone{
				Iana:          "America/New_York",
				OffsetMinutes: -300, // -5 hours = -300 minutes
			},
		}

		// First get the current offset minutes for the timezone
		offsetMinutes, err := getOffsetMinutes("America/New_York", time.Now())
		assert.NoError(t, err, "Should not return an error for valid timezone")

		// Set the timezone with the current offset
		v.Timezone.OffsetMinutes = offsetMinutes

		updates, err := v.SetTimezone("America/New_York")

		assert.NoError(t, err, "Should not return an error for valid timezone")
		assert.Equal(t, "America/New_York", v.Timezone.Iana, "Iana should remain the same")
		assert.Equal(t, offsetMinutes, v.Timezone.OffsetMinutes, "OffsetMinutes should remain the same")
		assert.Empty(t, updates, "Should return no updates when values are the same")
	})

	t.Run("invalid_timezone_name", func(t *testing.T) {
		v := &WithTimezone{}
		updates, err := v.SetTimezone("Invalid/Timezone")

		assert.Error(t, err, "Should return an error for invalid timezone name")
		assert.Nil(t, updates, "Should not return any updates when error occurs")
		assert.Nil(t, v.Timezone, "Timezone should remain nil when error occurs")
	})

	t.Run("empty_timezone_name", func(t *testing.T) {
		v := &WithTimezone{}
		updates, err := v.SetTimezone("")

		assert.Error(t, err, "Should return an error for empty timezone name")
		assert.Nil(t, updates, "Should not return any updates when error occurs")
		assert.Nil(t, v.Timezone, "Timezone should remain nil when error occurs")
	})
}
