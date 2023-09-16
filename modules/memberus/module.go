package memberus

import (
	"github.com/sneat-co/sneat-go-core/modules"
	"github.com/sneat-co/sneat-go-core/modules/memberus/api4memberus"
	"github.com/sneat-co/sneat-go-core/modules/memberus/const4memberus"
)

func Module() modules.Module {
	return modules.NewModule(const4memberus.ModuleID, api4memberus.RegisterHttpRoutes)
}
