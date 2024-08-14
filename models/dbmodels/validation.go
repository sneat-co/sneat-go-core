package dbmodels

import (
	"fmt"
	"github.com/sneat-co/sneat-go-core"
	"github.com/strongo/validation"
	"slices"
	"strings"
)

func ValidateWithIdsAndBriefs[R core.Validatable](idsField, briefsField string, ids []string, briefs map[string]R) error {
	if len(ids) == 0 {
		return validation.NewErrRecordIsMissingRequiredField(idsField)
	}
	if ids[0] != "*" {
		return validation.NewErrBadRecordFieldValue(idsField, "first element should be '*'")
	}
	for _, id := range ids[1:] {
		if _, ok := briefs[id]; !ok {
			return validation.NewErrRecordIsMissingRequiredField(fmt.Sprintf("%s[%s]", briefsField, id))
		}
	}
	for id, r := range briefs {
		field := func() string {
			return fmt.Sprintf("%s[%s]", briefsField, id)
		}
		if !slices.Contains(ids[1:], id) {
			return validation.NewErrBadRecordFieldValue(field(), "id is not in "+idsField)
		}
		//if r == nil {
		//	return validation.NewErrRecordIsMissingRequiredField(field())
		//}
		if err := r.Validate(); err != nil {
			return validation.NewErrBadRecordFieldValue(field(), err.Error())
		}
	}
	return nil
}

// ValidateRequiredName validates required names
//func ValidateRequiredName(v *Name) error {
//	if strings.TrimSpace(v.First) == "" && strings.TrimSpace(v.Last) == "" && strings.TrimSpace(v.Full) == "" && strings.TrimSpace(v.Nick) == "" {
//		return validation.NewErrBadRecordFieldValue("first|last|full|nick", "at least one of names should be specified")
//	}
//	return nil
//}

// ValidateTitle validates title
func ValidateTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return validation.NewErrRecordIsMissingRequiredField("title")
	}
	if strings.TrimSpace(title) != title {
		return validation.NewErrBadRecordFieldValue("name", "title should not start or end with whitespace characters")
	}
	return nil

}
