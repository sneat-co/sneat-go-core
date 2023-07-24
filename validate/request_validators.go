package validate

import (
	"errors"
	"github.com/strongo/validation"
	"net/mail"
	"strings"
)

// RequestID validates request ContactID
func RequestID(id, fieldName string) error { // TODO: move into github.com/strongo/validation/validate
	if v := strings.TrimSpace(id); v == "" {
		return validation.NewErrRequestIsMissingRequiredField("title")
	} else if v != id {
		return validation.NewErrBadRequestFieldValue(fieldName, "id should not start or end with whitespace characters")
	}
	return nil
}

// RecordID validates record ContactID
func RecordID(id string) error { // TODO: move into github.com/strongo/validation/validate
	if strings.TrimSpace(id) != id {
		return errors.New("id should not start or end with whitespace characters")
	}
	if strings.ContainsAny(id, " \t\r") {
		return errors.New("must not contain whitespace characters")
	}
	return nil
}

// RequestTitle validates request title
func RequestTitle(title, fieldName string) error { // TODO: move into github.com/strongo/validation/validate
	if v := strings.TrimSpace(title); v == "" {
		return validation.NewErrRequestIsMissingRequiredField("title")
	} else if v != title {
		return validation.NewErrBadRequestFieldValue(fieldName, "title should not start or end with whitespace characters")
	}
	return nil
}

// RecordTitle validates record title
func RecordTitle(title, fieldName string) error { // TODO: move into github.com/strongo/validation/validate
	if v := strings.TrimSpace(title); v == "" {
		return validation.NewErrRequestIsMissingRequiredField("title")
	} else if v != title {
		return validation.NewErrBadRecordFieldValue(fieldName, "title should not start or end with whitespace characters")
	}
	return nil
}

// RequiredEmail validates required email
func RequiredEmail(email, fieldName string) error { // TODO: move into github.com/strongo/validation/validate
	if v := strings.TrimSpace(email); v == "" {
		return validation.NewErrRequestIsMissingRequiredField(fieldName)
	}
	return validateEmail(email, fieldName)
}

// OptionalEmail validates optional email
func OptionalEmail(email, fieldName string) error { // TODO: move into github.com/strongo/validation/validate
	return validateEmail(email, fieldName)
}

func validateEmail(email, fieldName string) error {
	if v := strings.TrimSpace(email); v != email {
		return validation.NewErrBadRequestFieldValue(fieldName, "email should not start or end with whitespace characters")
	} else if _, err := mail.ParseAddress(email); err != nil {
		return validation.NewErrBadRequestFieldValue(fieldName, err.Error())
	}
	return nil
}
