package extension

import (
	"context"
	"fmt"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/strongo/delaying"
)

type OptionID string

// Config is the interface for extension configuration that every extension must implement.
type Config interface {
	internal()
	ID() coretypes.ExtID
	Register(args RegistrationArgs)
}

type RegistrationArgs interface {
	Handle() HTTPHandleFunc
	MustRegisterDelayFunc() func(key string, i any) delaying.Delayer
	CreateNotificationFunc() CreateNotificationFunc
}

var _ RegistrationArgs = (*registrationArgs)(nil)

type registrationArgs struct {
	handle                HTTPHandleFunc
	mustRegisterDelayFunc func(key string, i any) delaying.Delayer
	createNotification    CreateNotificationFunc
}

func (a registrationArgs) CreateNotificationFunc() CreateNotificationFunc {
	return a.createNotification
}

func (a registrationArgs) Handle() HTTPHandleFunc {
	return a.handle
}

func (a registrationArgs) MustRegisterDelayFunc() func(key string, i any) delaying.Delayer {
	return a.mustRegisterDelayFunc
}

func NewModuleRegistrationArgs(handle HTTPHandleFunc, mustRegisterDelayFunc func(key string, i any) delaying.Delayer) RegistrationArgs {
	return &registrationArgs{handle: handle, mustRegisterDelayFunc: mustRegisterDelayFunc}
}

var _ Config = (*config)(nil)

type CreateNotificationFunc func(ctx context.Context, args NotificationArgs) (m any, err error)

type BotProfileParams struct {
}

var _ Config = (*config)(nil)

// config implements Config interface
type config struct {
	id                   coretypes.ExtID
	registerRoutes       []func(handle HTTPHandleFunc)
	registerDelays       []func(mustRegisterFunc func(key string, i any) delaying.Delayer)
	registerNotificators []func(createNotification CreateNotificationFunc)
}

func (m *config) internal() {}

func (m *config) Register(args RegistrationArgs) {

	if len(m.registerRoutes) > 0 {
		handle := args.Handle()
		if handle == nil {
			panic(fmt.Sprintf("can not register module as HTTP handle has not been provided (moduleID=%s)", m.id))
		}
		for _, r := range m.registerRoutes {
			r(handle)
		}
	}

	if len(m.registerDelays) > 0 {
		mustRegisterDelayFunc := args.MustRegisterDelayFunc()
		if mustRegisterDelayFunc == nil {
			panic(fmt.Sprintf("can not register module as mustRegisterDelayFunc has not been provided (moduleID=%s)", m.id))
		}
		for _, r := range m.registerDelays {
			r(mustRegisterDelayFunc)
		}
	}

	for _, r := range m.registerNotificators {
		r(args.CreateNotificationFunc())
	}
}

func (m *config) ID() coretypes.ExtID {
	return m.id
}

type Option func(m Config)

func NewExtension(id coretypes.ExtID, options ...Option) Config {
	m := &config{id: id}
	for _, option := range options {
		option(m)
	}
	return m
}

func RegisterRoutes(registerRoutes func(handle HTTPHandleFunc)) Option {
	return func(m Config) {
		c := m.(*config)
		c.registerRoutes = append(c.registerRoutes, registerRoutes)
	}
}

func RegisterDelays(registerDelays func(mustRegisterFunc func(key string, i any) delaying.Delayer)) Option {
	return func(m Config) {
		c := m.(*config)
		c.registerDelays = append(c.registerDelays, registerDelays)
	}
}

func RegisterNotificator(registerNotificator func(createNotificationMessage CreateNotificationFunc)) Option {
	return func(m Config) {
		c := m.(*config)
		c.registerNotificators = append(c.registerNotificators, registerNotificator)
	}
}
