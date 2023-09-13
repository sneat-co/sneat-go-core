package api4invitus

import (
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"github.com/sneat-co/sneat-go-core/modules/invitus/facade4invitus"
	"net/http"
)

// httpPostRejectPersonalInvite rejects personal invite
func httpPostRejectPersonalInvite(w http.ResponseWriter, r *http.Request) {
	verifyOptions := verify.Request(verify.MinimumContentLength(apicore.MinJSONRequestSize), verify.MaximumContentLength(10*apicore.KB), verify.AuthenticationRequired(false))
	ctx, userContext, err := apicore.VerifyRequestAndCreateUserContext(w, r, verifyOptions)
	if err != nil {
		return
	}
	request := facade4invitus.RejectPersonalInviteRequest{}
	if err = apicore.DecodeRequestBody(w, r, &request); err != nil {
		return
	}
	if err = facade4invitus.RejectPersonalInvite(ctx, userContext, request); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	apicore.IfNoErrorReturnOK(ctx, w, r, err)
}
