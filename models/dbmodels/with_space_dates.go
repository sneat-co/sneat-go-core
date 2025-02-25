package dbmodels

import (
	"fmt"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/sneat-co/sneat-go-core/validate"
	"github.com/strongo/slice"
	"github.com/strongo/strongoapp/with"
	"github.com/strongo/validation"
	"regexp"
	"strings"
)

var reSpaceDate = regexp.MustCompile(`\w+:\d{4}-[01]\d-[0-3]\d`)

// WithSpaceDates holds date properties for indexed queries
type WithSpaceDates struct {
	WithSpaceIDs
	with.DatesFields
	SpaceDates []string `json:"spaceDates" firestore:"spaceDates"`
}

// Validate returns error if not valid
func (v *WithSpaceDates) Validate() error {
	if err := v.WithSpaceIDs.Validate(); err != nil {
		return err
	}
	if err := v.DatesFields.Validate(); err != nil {
		return err
	}
	v.populateSpaceDatesField()
	if len(v.SpaceDates) != len(v.SpaceIDs)*len(v.Dates) {
		message := fmt.Sprintf("len(v.SpaceDates) != len(v.SpaceIDs) * len(v.DatesFields): %v != %v*%v: {spaceIDs=%v, dates: %v, spaceDates: %v}",
			len(v.SpaceDates),
			len(v.SpaceIDs),
			len(v.Dates),
			v.JoinSpaceIDs(","),
			strings.Join(v.Dates, ","),
			strings.Join(v.SpaceDates, ","),
		)
		return validation.NewErrBadRequestFieldValue("dates,spaceDates", message)
	}
	if err := v.validateDates(); err != nil {
		return err
	}
	if err := v.validateSpaceDates(); err != nil {
		return err
	}
	return nil
}

func (v *WithSpaceDates) populateSpaceDatesField() {
	v.SpaceDates = make([]string, 0, len(v.SpaceIDs)*len(v.Dates))
	for _, spaceID := range v.SpaceIDs {
		for _, date := range v.Dates {
			v.SpaceDates = append(v.SpaceDates, string(spaceID)+":"+date)
		}
	}
}

func (v *WithSpaceDates) validateDates() error {
	for i, date := range v.Dates {
		if strings.TrimSpace(date) == "" {
			return validation.NewErrRecordIsMissingRequiredField(fmt.Sprintf("dates[%v]", i))
		}
		if _, err := validate.DateString(date); err != nil {
			return validation.NewErrBadRecordFieldValue("date", err.Error())
		}

		for j, date2 := range v.Dates {
			if j != i && date2 == date {
				return validation.NewErrBadRecordFieldValue("dates", "duplicate value: "+date)
			}
		}

		isInSpaceDate := false
		for _, td := range v.SpaceDates {
			if strings.HasSuffix(td, ":"+date) {
				isInSpaceDate = true
			}
		}
		if !isInSpaceDate {
			return validation.NewErrBadRecordFieldValue(fmt.Sprintf("dates[%v]", i), "the value is not in 'spaceDates'")
		}
	}
	return nil
}

func (v *WithSpaceDates) validateSpaceDates() error {
	field := func(i int) string {
		return fmt.Sprintf("spaceDates[%v]", i)
	}
	for i, td := range v.SpaceDates {
		for j, td2 := range v.SpaceDates {
			if j != i && td2 == td {
				return validation.NewErrBadRecordFieldValue("spaceDates", "duplicate value: "+td)
			}
		}
		if !reSpaceDate.MatchString(td) {
			return validation.NewErrBadRecordFieldValue(field(i), "invalid format: "+td)
		}
		s := strings.Split(td, ":")
		spaceID := coretypes.SpaceID(s[0])
		date := s[1]
		if slice.Index(v.SpaceIDs, spaceID) < 0 {
			return validation.NewErrBadRecordFieldValue(
				field(i),
				fmt.Sprintf("reference spaceID %v that is not in 'spaceIDs' field: {spaceIDs: %+v}", spaceID, v.SpaceIDs),
			)
		}
		if slice.Index(v.Dates, date) < 0 {
			return validation.NewErrBadRecordFieldValue(
				field(i),
				fmt.Sprintf("reference date %v that is not in 'dates' field: {dates: %+v}", date, v.DatesFields),
			)
		}
	}
	return nil
}
