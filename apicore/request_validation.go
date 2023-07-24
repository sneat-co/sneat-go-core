package apicore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/datatug/datatug/packages/server/endpoints"
	facade2 "github.com/sneat-co/sneat-go/src/core/facade"
	"github.com/sneat-co/sneat-go/src/core/httpserver"
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

// NewFirestoreContext creates a context with a Firebase ContactID token
var NewFirestoreContext = func(r *http.Request, authRequired bool) (firestoreContext *facade2.FirestoreContext, err error) {
	var ctx context.Context
	if ctx, err = ContextWithFirebaseToken(r, authRequired); err != nil {
		return nil, err
	}
	return facade2.NewContextWithFirestoreClient(ctx)
}

type VerifyAuthenticatedRequestAndDecodeBodyFunc = func(
	w http.ResponseWriter, r *http.Request,
	options endpoints.VerifyRequestOptions,
	request facade2.Request,
) (ctx context.Context, userContext facade2.User, err error)

// VerifyAuthenticatedRequestAndDecodeBody decodes & verifies an HTTP request
var VerifyAuthenticatedRequestAndDecodeBody = func(
	w http.ResponseWriter, r *http.Request,
	options endpoints.VerifyRequestOptions,
	request facade2.Request,
) (ctx context.Context, userContext facade2.User, err error) {
	ctx, userContext, err = VerifyRequestAndCreateUserContext(w, r, options)
	if err != nil {
		return
	}
	if err = DecodeRequestBody(w, r, request); err != nil {
		return
	}
	return ctx, userContext, err
}

// VerifyRequestAndCreateUserContext runs common checks
var VerifyRequestAndCreateUserContext = func(
	w http.ResponseWriter, r *http.Request, options endpoints.VerifyRequestOptions,
) (ctx context.Context, userContext facade2.User, err error) {
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
	var authContext facade2.AuthContext
	if authContext, err = NewAuthContext(r); err != nil {
		httpserver.HandleError(err, from, w)
		return
	}
	if userContext, err = authContext.User(r.Context(), options.AuthenticationRequired()); err != nil {
		httpserver.HandleError(err, from, w)
		return
	}
	if ctx, err = VerifyRequest(w, r, options); err != nil {
		httpserver.HandleError(err, from, w)
		return
	}
	if UserContextProvider != nil {
		userContext = UserContextProvider()
		return
	}
	return
}

// UserContextProvider defines signature foe a function that provides user context
var UserContextProvider func() facade2.User

//type verifyRequestOptions struct {
//	minimumContentLength   int64
//	maximumContentLength   int64
//	authenticationRequired bool
//	processUserID          func(uid string) error
//}
//
//var _ endpoints.VerifyRequestOptions = (*verifyRequestOptions)(nil)
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
//func VerifyRequestOptions(opts) endpoints.VerifyRequestOptions {
//
//}

// VerifyRequest runs common checks
var VerifyRequest = func(w http.ResponseWriter, r *http.Request, options endpoints.VerifyRequestOptions) (ctx context.Context, err error) {
	ctx = r.Context()
	if !AccessControlAllowOrigin(w, r) {
		err = errBadOrigin
		return
	}
	if err = validateContentLength(r, options.MinimumContentLength(), options.MaximumContentLength()); err != nil {
		return
	}

	if ctx, err = ContextWithFirebaseToken(r, options.AuthenticationRequired()); err != nil {
		err = fmt.Errorf("failed to create context wuth firestore client: %w", err)
		return
	}
	return
}

// DecodeRequestBody decodes body of HTTP request into a provide struct
func DecodeRequestBody(w http.ResponseWriter, r *http.Request, request facade2.Request) (err error) {
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
			err = fmt.Errorf("Request struct decoded from HTTP request body failed initial validation %T: %v", request, err)
			httpserver.HandleError(err, "DecodeRequestBody", w)
			return err
		}
	}
	return nil
}
