package dbmodels

import (
	"fmt"
	"github.com/sneat-co/sneat-go-core/validate"
	"github.com/strongo/slice"
	"github.com/strongo/strongoapp/with"
	"github.com/strongo/validation"
	"regexp"
	"strings"
)

var reTeamDate = regexp.MustCompile(`\w+:\d{4}-[01]\d-[0-3]\d`)

// WithTeamDates holds date properties for indexed queries
type WithTeamDates struct {
	WithTeamIDs
	with.DatesFields
	TeamDates []string `json:"teamDates" firestore:"teamDates"`
}

// Validate returns error if not valid
func (v *WithTeamDates) Validate() error {
	if err := v.WithTeamIDs.Validate(); err != nil {
		return err
	}
	if err := v.DatesFields.Validate(); err != nil {
		return err
	}
	v.populateTeamDatesField()
	if len(v.TeamDates) != len(v.TeamIDs)*len(v.Dates) {
		message := fmt.Sprintf("len(v.TeamDates) != len(v.TeamIDs) * len(v.DatesFields): %v != %v*%v: {teamIDs=%v, dates: %v, teamDates: %v}",
			len(v.TeamDates),
			len(v.TeamIDs),
			len(v.Dates),
			strings.Join(v.TeamIDs, ","),
			strings.Join(v.Dates, ","),
			strings.Join(v.TeamDates, ","),
		)
		return validation.NewErrBadRequestFieldValue("dates,teamDates", message)
	}
	if err := v.validateDates(); err != nil {
		return err
	}
	if err := v.validateTeamDates(); err != nil {
		return err
	}
	return nil
}

func (v *WithTeamDates) populateTeamDatesField() {
	v.TeamDates = make([]string, 0, len(v.TeamIDs)*len(v.Dates))
	for _, teamID := range v.TeamIDs {
		for _, date := range v.Dates {
			v.TeamDates = append(v.TeamDates, teamID+":"+date)
		}
	}
}

func (v WithTeamDates) validateDates() error {
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

		isInTeamDate := false
		for _, td := range v.TeamDates {
			if strings.HasSuffix(td, ":"+date) {
				isInTeamDate = true
			}
		}
		if !isInTeamDate {
			return validation.NewErrBadRecordFieldValue(fmt.Sprintf("dates[%v]", i), "the value is not in 'teamDates'")
		}
	}
	return nil
}

func (v WithTeamDates) validateTeamDates() error {
	field := func(i int) string {
		return fmt.Sprintf("teamDates[%v]", i)
	}
	for i, td := range v.TeamDates {
		for j, td2 := range v.TeamDates {
			if j != i && td2 == td {
				return validation.NewErrBadRecordFieldValue("teamDates", "duplicate value: "+td)
			}
		}
		if !reTeamDate.MatchString(td) {
			return validation.NewErrBadRecordFieldValue(field(i), "invalid format: "+td)
		}
		s := strings.Split(td, ":")
		teamID := s[0]
		date := s[1]
		if slice.Index(v.TeamIDs, teamID) < 0 {
			return validation.NewErrBadRecordFieldValue(
				field(i),
				fmt.Sprintf("reference teamID %v that is not in 'teamIDs' field: {teamIDs: %+v}", teamID, v.TeamIDs),
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
