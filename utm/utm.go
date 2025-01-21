package utm

import (
	"fmt"
	"net/url"
)

type Params struct {
	Source   string // (referrer: google, citysearch, newsletter4), in our case it's BotID
	Medium   string // In our case it's bot platform e.g. 'telegram', 'fbm', etc.
	Campaign string // Identify where link is placed. For example 'receipt'
}

func (utm Params) IsEmpty() bool {
	return utm.Source == "" && utm.Medium == "" && utm.Campaign == ""
}

func (utm Params) String() string {
	switch "" {
	case utm.Source:
		panic("utm.Source is not provided")
	case utm.Medium:
		panic("utm.Medium is not provided")
	case utm.Campaign:
		panic("utm.Campaign is not provided")
	}
	return fmt.Sprintf("utm_source=%v&utm_medium=%v&utm_campaign=%v",
		url.QueryEscape(utm.Source), url.QueryEscape(utm.Medium), url.QueryEscape(utm.Campaign))
}

func (utm Params) ShortString() string {
	switch "" {
	case utm.Source:
		panic("utm.Source is not provided")
	case utm.Medium:
		panic("utm.Medium is not provided")
	case utm.Campaign:
		panic("utm.Campaign is not provided")
	}
	return fmt.Sprintf("utm=%v;%v;%v",
		url.QueryEscape(utm.Source), url.QueryEscape(utm.Medium), url.QueryEscape(utm.Campaign))
}
