package api4userus

import (
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"github.com/sneat-co/sneat-go-core/modules/userus/dto4userus"
	"github.com/sneat-co/sneat-go-core/modules/userus/facade4userus"
	"github.com/sneat-co/sneat-go-core/modules/userus/models4userus"
	"net/http"
)

var initUserRecord = facade4userus.InitUserRecord

// httpInitUserRecord sets user title
func httpInitUserRecord(w http.ResponseWriter, r *http.Request) {
	verifyOptions := verify.Request(verify.AuthenticationRequired(false), verify.MinimumContentLength(apicore.MinJSONRequestSize), verify.MaximumContentLength(10*apicore.KB))
	ctx, userContext, err := apicore.VerifyRequestAndCreateUserContext(w, r, verifyOptions)
	if err != nil {
		return
	}
	var request dto4userus.InitUserRecordRequest
	if err = apicore.DecodeRequestBody(w, r, &request); err != nil {
		return
	}
	request.RemoteClient = apicore.GetRemoteClientInfo(r)
	var user models4userus.UserContext
	user, err = initUserRecord(ctx, userContext, request)
	apicore.ReturnJSON(ctx, w, r, http.StatusOK, err, user.Dto)
}
