package module

import (
	"fmt"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/strongo/delaying"
)

// Module is the interface for config definition that all module must implement.
type Module interface {
	ID() coretypes.ModuleID
	Register(args RegistrationArgs)
}

type RegistrationArgs interface {
	Handle() HTTPHandleFunc
	MustRegisterDelayFunc() func(key string, i any) delaying.Delayer
}

var _ RegistrationArgs = (*registrationArgs)(nil)

type registrationArgs struct {
	handle                HTTPHandleFunc
	mustRegisterDelayFunc func(key string, i any) delaying.Delayer
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

var _ Module = (*config)(nil)

type config struct {
	id             coretypes.ModuleID
	registerRoutes func(handle HTTPHandleFunc)
	registerDelays func(mustRegisterFunc func(key string, i any) delaying.Delayer)
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

func (m *config) ID() coretypes.ModuleID {
	return m.id
}

type Option func(m *config)

func NewModule(id coretypes.ModuleID, options ...Option) Module {
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

func RegisterDelays(registerDelays func(mustRegisterFunc func(key string, i any) delaying.Delayer)) Option {
	return func(m *config) {
		m.registerDelays = registerDelays
	}
}
