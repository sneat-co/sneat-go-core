package dbmodels

import (
	"errors"
	"fmt"
	"slices"
)

var knownAuthProviderIDs = []string{
	"password",
	"firebase",
	"google.com",
	"apple.com",
	"microsoft.com",
}

func ValidateAuthProviderID(v string) error {
	if v == "" {
		return errors.New("is empty string")
	}
	if slices.Contains(knownAuthProviderIDs, v) {
		return nil
	}
	return fmt.Errorf("supported auth providers=%+v, got: %s", knownAuthProviderIDs, v)
}
