package dbmodels

import (
	"fmt"
	"github.com/sneat-co/sneat-go/src/core"
	"github.com/sneat-co/sneat-go/src/core/validate"
	"github.com/strongo/slice"
	"github.com/strongo/validation"
	"strings"
)

func ValidateSetSliceField(field string, v []string, isRecordID bool) error {
	count := len(v)
	for i, s := range v {
		if s2 := strings.TrimSpace(s); s2 == "" {
			return validation.NewErrRecordIsMissingRequiredField(fmt.Sprintf("%v[%v]", field, i))
		} else if s2 != s {
			return validation.NewErrBadRecordFieldValue(fmt.Sprintf("%v[%v]", field, i), "starts or ends with spaces")
		}
		if i < count {
			if slice.Contains(v[i+1:], s) {
				return validation.NewErrBadRecordFieldValue(field,
					fmt.Sprintf("duplicate value at indexes %d & %d: %s", i, slice.Index(v[i+1:], s), s))
			}
		}
		if isRecordID {
			if err := validate.RecordID(s); err != nil {
				return validation.NewErrBadRecordFieldValue(fmt.Sprintf("%s[%v]", field, i), err.Error())
			}
		}
	}
	return nil
}

func ValidateWithIdsAndBriefs[K string | TeamItemID, R core.Validatable](idsField, briefsField string, ids []K, briefs map[K]R) error {
	for id, r := range briefs {
		if !slice.Contains(ids[1:], id) {
			return validation.NewErrBadRecordFieldValue(fmt.Sprintf("%s[%s]", briefsField, id), "id is not in "+idsField)
		}
		if err := r.Validate(); err != nil {
			return validation.NewErrBadRecordFieldValue(fmt.Sprintf("%s[%s]", briefsField, id), err.Error())
		}
	}
	for _, id := range ids[1:] {
		if _, ok := briefs[id]; !ok {
			return validation.NewErrRecordIsMissingRequiredField(fmt.Sprintf("%s[%s]", briefsField, id))
		}
	}
	return nil
}

// ValidateRequiredName validates required names
func ValidateRequiredName(v *Name) error {
	if strings.TrimSpace(v.First) == "" && strings.TrimSpace(v.Last) == "" && strings.TrimSpace(v.Full) == "" {
		return validation.NewErrBadRecordFieldValue("first|last|full", "at least one of names should be specified")
	}
	return nil
}