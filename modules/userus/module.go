package userus

import (
	"github.com/sneat-co/sneat-go-core/modules"
	"github.com/sneat-co/sneat-go-core/modules/userus/api4userus"
	"github.com/sneat-co/sneat-go-core/modules/userus/const4userus"
)

func Module() modules.Module {
	return modules.NewModule(const4userus.ModuleID, api4userus.RegisterHttpRoutes)
}
