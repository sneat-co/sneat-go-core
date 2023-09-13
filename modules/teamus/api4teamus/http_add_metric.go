package api4teamus

import (
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"github.com/sneat-co/sneat-go-core/modules/teamus/facade4teamus"
	"net/http"
	"strings"
)

// httpPostAddMetric is an API endpoint that adds a metric
func httpPostAddMetric(w http.ResponseWriter, r *http.Request) {
	verifyOptions := verify.Request(verify.MinimumContentLength(apicore.MinJSONRequestSize), verify.MaximumContentLength(10*apicore.KB), verify.AuthenticationRequired(true))
	ctx, userContext, err := apicore.VerifyRequestAndCreateUserContext(w, r, verifyOptions)
	if err != nil {
		return
	}
	var request facade4teamus.AddTeamMetricRequest
	if request.TeamID = r.URL.Query().Get("id"); strings.TrimSpace(request.TeamID) == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("team 'id' should be passed as query parameter"))
		return
	}
	if err = apicore.DecodeRequestBody(w, r, &request); err != nil {
		return
	}
	response, err := addMetric(ctx, userContext, request)
	apicore.ReturnJSON(ctx, w, r, http.StatusCreated, err, response)
}

var addMetric = facade4teamus.AddMetric
