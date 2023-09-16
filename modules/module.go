package modules

// Module is the interface for module definition that all modules must implement.
type Module interface {
	ID() string
	Register(args ModuleRegistrationArgs)
}

type ModuleRegistrationArgs interface {
	Handle() HTTPHandleFunc
}

var _ ModuleRegistrationArgs = (*moduleRegistrationArgs)(nil)

type moduleRegistrationArgs struct {
	handle HTTPHandleFunc
}

func (a moduleRegistrationArgs) Handle() HTTPHandleFunc {
	return a.handle
}

func NewModuleRegistrationArgs(handle HTTPHandleFunc) ModuleRegistrationArgs {
	return &moduleRegistrationArgs{handle: handle}
}

var _ Module = (*module)(nil)

type module struct {
	id             string
	registerRoutes func(handle HTTPHandleFunc)
}

func (m *module) Register(args ModuleRegistrationArgs) {
	handle := args.Handle()
	m.registerRoutes(handle)
}

func (m *module) ID() string {
	return m.id
}

func NewModule(id string, registerRoutes func(handle HTTPHandleFunc)) Module {
	return &module{id: id, registerRoutes: registerRoutes}
}
