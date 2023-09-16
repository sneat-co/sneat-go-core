package invitus

import (
	"github.com/sneat-co/sneat-go-core/modules"
	"github.com/sneat-co/sneat-go-core/modules/invitus/api4invitus"
	"github.com/sneat-co/sneat-go-core/modules/invitus/const4invitus"
)

func Module() modules.Module {
	return modules.NewModule(const4invitus.ModuleID, api4invitus.RegisterHttpRoutes)
}
