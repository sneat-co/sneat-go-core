package dbmodels

import "errors"
import "strings"

// SpaceItemIDSeparatorChar is a separator character between space ID and item ID
const SpaceItemIDSeparatorChar = '_'
const SpaceItemIDSeparator = "_"

// NewSpaceItemID returns space item ID as a concatenation of space ID and item ID
func NewSpaceItemID(spaceID, id string) SpaceItemID {
	return SpaceItemID(spaceID + SpaceItemIDSeparator + id)
}

type SpaceItemID string

func (v SpaceItemID) SpaceID() string {
	return string(v[:strings.IndexByte(string(v), SpaceItemIDSeparatorChar)])
}

func (v SpaceItemID) ItemID() string {
	return string(v[strings.IndexByte(string(v), SpaceItemIDSeparatorChar)+1:])
}

func (v SpaceItemID) Validate() error {
	s := string(v)
	if s == "" {
		return errors.New("space item ID is empty")
	}
	separatorIndex := strings.IndexByte(s, SpaceItemIDSeparatorChar)
	if separatorIndex < 0 {
		return errors.New("space item ID is missing separator char")
	}
	if separatorIndex == 0 {
		return errors.New("space item ID is missing space ID")
	}
	if separatorIndex == len(s)-1 {
		return errors.New("space item ID is missing item ID")
	}
	if strings.IndexByte(s[separatorIndex+1:], SpaceItemIDSeparatorChar) >= 0 {
		return errors.New("space item ID has too many separator chars")
	}
	return nil
}
