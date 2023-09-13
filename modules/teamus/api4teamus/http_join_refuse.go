package api4teamus

import (
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"github.com/sneat-co/sneat-go-core/modules/memberus/facade4memberus"
	"net/http"
	"strconv"
)

var refuseToJoinTeam = facade4memberus.RefuseToJoinTeam

// httpPostRefuseToJoinTeam an API endpoint that records user refusal to join a team
func httpPostRefuseToJoinTeam(w http.ResponseWriter, r *http.Request) {
	verifyOptions := verify.Request(verify.AuthenticationRequired(false), verify.MinimumContentLength(apicore.MinJSONRequestSize), verify.MaximumContentLength(10*apicore.KB))
	ctx, userContext, err := apicore.VerifyRequestAndCreateUserContext(w, r, verifyOptions)
	if err != nil {
		return
	}
	q := r.URL.Query()
	var pin int
	if pin, err = strconv.Atoi(q.Get("pin")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("pin is expected to be an integer"))
		return
	}
	request := facade4memberus.RefuseToJoinTeamRequest{
		TeamID: q.Get("id"),
		Pin:    int32(pin),
	}
	err = refuseToJoinTeam(ctx, userContext, request)
	apicore.IfNoErrorReturnOK(ctx, w, r, err)
}
