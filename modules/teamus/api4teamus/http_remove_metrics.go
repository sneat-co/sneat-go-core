package api4teamus

import (
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dto4teamus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/facade4teamus"
	"net/http"
)

var removeMetrics = facade4teamus.RemoveMetrics

// httpPostRemoveMetrics is an API endpoint that removes a team metric
func httpPostRemoveMetrics(w http.ResponseWriter, r *http.Request) {
	verifyOptions := verify.Request(verify.AuthenticationRequired(false), verify.MinimumContentLength(apicore.MinJSONRequestSize), verify.MaximumContentLength(10*apicore.KB))
	ctx, userContext, err := apicore.VerifyRequestAndCreateUserContext(w, r, verifyOptions)
	if err != nil {
		return
	}
	var request dto4teamus.TeamMetricsRequest
	if err = apicore.DecodeRequestBody(w, r, &request); err != nil {
		return
	}
	err = removeMetrics(ctx, userContext, request)
	apicore.IfNoErrorReturnOK(ctx, w, r, err)
}
