package module

import (
	"fmt"
	"github.com/strongo/delaying"
)

// Module is the interface for config definition that all module must implement.
type Module interface {
	ID() string
	Register(args RegistrationArgs)
}

type RegistrationArgs interface {
	Handle() HTTPHandleFunc
	MustRegisterDelayFunc() func(key string, i any) delaying.Function
}

var _ RegistrationArgs = (*registrationArgs)(nil)

type registrationArgs struct {
	handle                HTTPHandleFunc
	mustRegisterDelayFunc func(key string, i any) delaying.Function
}

func (a registrationArgs) Handle() HTTPHandleFunc {
	return a.handle
}

func (a registrationArgs) MustRegisterDelayFunc() func(key string, i any) delaying.Function {
	return a.mustRegisterDelayFunc
}

func NewModuleRegistrationArgs(handle HTTPHandleFunc, mustRegisterDelayFunc func(key string, i any) delaying.Function) RegistrationArgs {
	return &registrationArgs{handle: handle, mustRegisterDelayFunc: mustRegisterDelayFunc}
}

var _ Module = (*config)(nil)

type config struct {
	id             string
	registerRoutes func(handle HTTPHandleFunc)
	registerDelays func(mustRegisterFunc func(key string, i any) delaying.Function)
}

func (m *config) Register(args RegistrationArgs) {

	if m.registerRoutes != nil {
		handle := args.Handle()
		if handle == nil {
			panic(fmt.Sprintf("can not register module as HTTP handle has not been provided (moduleID=%s)", m.id))
		}
		m.registerRoutes(handle)
	}

	if m.registerDelays != nil {
		mustRegisterDelayFunc := args.MustRegisterDelayFunc()
		if mustRegisterDelayFunc == nil {
			panic(fmt.Sprintf("can not register module as mustRegisterDelayFunc has not been provided (moduleID=%s)", m.id))
		}
		m.registerDelays(mustRegisterDelayFunc)
	}
}

func (m *config) ID() string {
	return m.id
}

type Option func(m *config)

func NewModule(id string, options ...Option) Module {
	m := &config{id: id}
	for _, option := range options {
		option(m)
	}
	return m
}
func RegisterRoutes(registerRoutes func(handle HTTPHandleFunc)) Option {
	return func(m *config) {
		m.registerRoutes = registerRoutes
	}
}

func RegisterDelays(registerDelays func(mustRegisterFunc func(key string, i any) delaying.Function)) Option {
	return func(m *config) {
		m.registerDelays = registerDelays
	}
}