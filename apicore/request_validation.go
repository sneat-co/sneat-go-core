package apicore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/httpserver"
	"github.com/sneat-co/sneat-go-core/sneatauth"
	"io"
	"log"
	"net/http"
	"strings"
)

var errBadOrigin = errors.New("bad origin")

// KB = 1024 bytes
const KB = 1024

// MinJSONRequestSize - non-empty json can't be less then 2 bytes, e.g. "{}"
const MinJSONRequestSize = 2

// DefaultMaxJSONRequestSize is set as 7 kilobytes what usually should be enough
const DefaultMaxJSONRequestSize = 7 * KB

type VerifyAuthenticatedRequestAndDecodeBodyFunc = func(
	w http.ResponseWriter, r *http.Request,
	options verify.RequestOptions,
	request facade.Request,
) (ctx context.Context, userContext facade.User, err error)

// VerifyAuthenticatedRequestAndDecodeBody decodes & verifies an HTTP request
var VerifyAuthenticatedRequestAndDecodeBody = func(
	w http.ResponseWriter, r *http.Request,
	options verify.RequestOptions,
	request facade.Request,
) (ctx context.Context, userContext facade.User, err error) {
	ctx, userContext, err = VerifyRequestAndCreateUserContext(w, r, options)
	if err != nil {
		return
	}
	if err = DecodeRequestBody(w, r, request); err != nil {
		return
	}
	return ctx, userContext, err
}

var NewAuthContext func(r *http.Request) (facade.AuthContext, error)

// VerifyRequestAndCreateUserContext runs common checks
var VerifyRequestAndCreateUserContext = func(
	w http.ResponseWriter, r *http.Request, options verify.RequestOptions,
) (ctx context.Context, userContext facade.User, err error) {
	if r == nil {
		panic("request is nil")
	}
	if w == nil {
		panic("response writer is nil")
	}
	if options == nil {
		panic("options is nil")
	}
	const from = "VerifyRequestAndCreateUserContext"
	var authContext facade.AuthContext
	if authContext, err = NewAuthContext(r); err != nil {
		httpserver.HandleError(err, from, w, r)
		return
	}
	if ctx, err = VerifyRequest(w, r, options); err != nil {
		httpserver.HandleError(err, from, w, r)
		return
	}
	if token := sneatauth.AuthTokenFromContext(ctx); token != nil {
		userContext = facade.AuthUser{ID: token.UID}
	}
	if userContext, err = authContext.User(r.Context(), options.AuthenticationRequired()); err != nil {
		httpserver.HandleError(err, from, w, r)
		return
	}
	return
}

//type verifyRequestOptions struct {
//	minimumContentLength   int64
//	maximumContentLength   int64
//	authenticationRequired bool
//	processUserID          func(uid string) error
//}
//
//var _ RequestOptions = (*verifyRequestOptions)(nil)
//
//func (v verifyRequestOptions) MinimumContentLength() int64 {
//	return v.minimumContentLength
//}
//
//func (v verifyRequestOptions) MaximumContentLength() int64 {
//	return v.maximumContentLength
//}
//
//func (v verifyRequestOptions) AuthenticationRequired() bool {
//	return v.authenticationRequired
//}
//
//func RequestOptions(opts) RequestOptions {
//
//}

var GetAuthTokenFromHttpRequest func(r *http.Request) (token *sneatauth.Token, err error)

// VerifyRequest runs common checks
var VerifyRequest = func(w http.ResponseWriter, r *http.Request, options verify.RequestOptions) (ctx context.Context, err error) {
	ctx = r.Context()
	if !httpserver.AccessControlAllowOrigin(w, r) {
		err = errBadOrigin
		return
	}
	if err = validateContentLength(r, options.MinimumContentLength(), options.MaximumContentLength()); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if GetAuthTokenFromHttpRequest == nil {
		panic("GetAuthTokenFromHttpRequest is nil")
	}

	var token *sneatauth.Token
	if token, err = GetAuthTokenFromHttpRequest(r); err != nil {
		err = fmt.Errorf("failed to get auth token from HTTP request: %w", err)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	if token == nil && options.AuthenticationRequired() {
		err = fmt.Errorf("authentication required")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	return
}

// DecodeRequestBody decodes body of HTTP request into a provide struct
func DecodeRequestBody(w http.ResponseWriter, r *http.Request, request facade.Request) (err error) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete && r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		err = fmt.Errorf("unsupported method: %v", r.Method)
		_, _ = fmt.Fprint(w, err.Error())
		return err
	}
	if r.ContentLength > 0 {
		var reader io.Reader = r.Body
		log.Println("HOST: " + r.Host)
		if strings.HasPrefix(r.Host, "localhost:") {
			var body []byte
			if body, err = io.ReadAll(r.Body); err != nil {
				_, _ = fmt.Fprintf(w, "Failed to read request body: %v", err)
				return err
			}
			log.Println("REQUEST BODY:\n", string(body))
			reader = bytes.NewReader(body)
		}
		if err = json.NewDecoder(reader).Decode(request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintf(w, "Failed to decode request body as JSON: %v", err)
			return err
		}
		if err = request.Validate(); err != nil {
			err = fmt.Errorf("request struct decoded from HTTP request body failed initial validation %T: %v", request, err)
			httpserver.HandleError(err, "DecodeRequestBody", w, r)
			return err
		}
	}
	return nil
}
