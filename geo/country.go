package geo

type CountryAlpha2 = string

// IsValidCountryAlpha2 checks if the given value is a valid country alpha2 code
func IsValidCountryAlpha2(v CountryAlpha2) bool {
	return len(v) == 2
}
