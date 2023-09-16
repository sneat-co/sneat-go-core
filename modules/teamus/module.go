package teamus

import (
	"github.com/sneat-co/sneat-go-core/modules"
	"github.com/sneat-co/sneat-go-core/modules/teamus/api4teamus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/const4teamus"
)

func Module() modules.Module {
	return modules.NewModule(const4teamus.ModuleID, api4teamus.RegisterHttpRoutes)
}
