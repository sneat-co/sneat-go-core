package modules

// Module is the interface for module definition that all modules must implement.
type Module interface {
	GetID() string
	Register(args RegistrationArgs)
}

type RegistrationArgs interface {
	Handle() HTTPHandleFunc
}

var _ RegistrationArgs = (*args)(nil)

type args struct {
	handle HTTPHandleFunc
}

func (a args) Handle() HTTPHandleFunc {
	return a.handle
}

func NewArgs(handle HTTPHandleFunc) RegistrationArgs {
	return &args{handle: handle}
}
