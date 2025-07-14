package facade

import "github.com/strongo/analytics"

type UserAnalytics interface {
	Send(msg analytics.Message)
}

var _ UserAnalytics = (*userAnalytics)(nil)

type userAnalytics struct {
	send func(msg analytics.Message)
}

func (v userAnalytics) Send(msg analytics.Message) {
	v.send(msg)
}

func NewUserAnalytics(send func(msg analytics.Message)) UserAnalytics {
	return userAnalytics{send: send}
}
