package dbmodels

import (
	"errors"
	"github.com/sneat-co/sneat-go-core/coretypes"
)
import "strings"

// SpaceItemIDSeparatorChar is a separator character between space userID and item userID
const SpaceItemIDSeparatorChar = '_'
const SpaceItemIDSeparator = "_"

// NewSpaceItemID returns space item userID as a concatenation of space userID and item userID
func NewSpaceItemID(spaceID coretypes.SpaceID, id string) SpaceItemID {
	return SpaceItemID(string(spaceID) + SpaceItemIDSeparator + id)
}

type SpaceItemID string

func (v SpaceItemID) SpaceID() coretypes.SpaceID {
	return coretypes.SpaceID(v[:strings.IndexByte(string(v), SpaceItemIDSeparatorChar)])
}

func (v SpaceItemID) ItemID() string {
	return string(v[strings.IndexByte(string(v), SpaceItemIDSeparatorChar)+1:])
}

func (v SpaceItemID) Validate() error {
	s := string(v)
	if s == "" {
		return errors.New("space item userID is empty")
	}
	separatorIndex := strings.IndexByte(s, SpaceItemIDSeparatorChar)
	if separatorIndex < 0 {
		return errors.New("space item userID is missing separator char")
	}
	if separatorIndex == 0 {
		return errors.New("space item userID is missing space userID")
	}
	if separatorIndex == len(s)-1 {
		return errors.New("space item userID is missing item userID")
	}
	if strings.IndexByte(s[separatorIndex+1:], SpaceItemIDSeparatorChar) >= 0 {
		return errors.New("space item userID has too many separator chars")
	}
	return nil
}
