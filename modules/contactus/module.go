package contactus

import (
	"github.com/sneat-co/sneat-go-core/modules"
	"github.com/sneat-co/sneat-go-core/modules/contactus/api4contactus"
	"github.com/sneat-co/sneat-go-core/modules/contactus/const4contactus"
)

func Module() modules.Module {
	return modules.NewModule(const4contactus.ModuleID, api4contactus.RegisterHttpRoutes)
}
