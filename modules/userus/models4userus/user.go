package models4userus

import (
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"strings"
)

// UsersCollection a name of the user's db table
const UsersCollection = "users"

type ShortName struct {
	Name string `json:"name" firestore:"name"`
	Type string `json:"type" firestore:"type"`
}

// UserDefaults keeps user's defaults
type UserDefaults struct {
	ShortNames []ShortName `json:"shortNames,omitempty" firestore:"shortNames,omitempty"`
}

// CleanTitle cleans title from spaces
func CleanTitle(title string) string {
	title = strings.TrimSpace(title)
	for strings.Contains(title, "  ") {
		title = strings.Replace(title, "  ", " ", -1)
	}
	return title
}

// GetShortNames returns short names from a title
func GetShortNames(title string) (shortNames []ShortName) {
	title = CleanTitle(title)
	names := strings.Split(title, " ")
	shortNames = make([]ShortName, 0, len(names))
NAMES:
	for _, s := range names {
		name := strings.TrimSpace(s)
		if name == "" {
			continue
		}
		for _, sn := range shortNames {
			if sn.Name == name {
				continue NAMES
			}
		}
		shortNames = append(shortNames, ShortName{
			Name: name,
			Type: "unknown",
		})
	}
	return shortNames
}

type User = record.DataWithID[string, *UserDto]

func NewUser(id string) User {
	data := new(UserDto)
	key := dal.NewKeyWithID(UsersCollection, id)
	return record.NewDataWithID(id, key, data)
}
