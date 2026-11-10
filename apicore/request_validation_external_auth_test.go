package apicore

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/sneatauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testReservedUID = "system-reserved-user"

func setupTestExternalVerifier(t *testing.T, returnedToken *sneatauth.Token, returnedErr error) {
	origGetter := GetAuthTokenFromHttpRequest
	t.Cleanup(func() {
		GetAuthTokenFromHttpRequest = origGetter
	})

	policy, err := sneatauth.NewReservedExternalUIDPolicy([]string{testReservedUID})
	require.NoError(t, err)

	rawVerifier := func(r *http.Request, authRequired bool) (*sneatauth.Token, error) {
		return returnedToken, returnedErr
	}

	wrapped, err := sneatauth.WrapExternalTokenVerifier(rawVerifier, policy)
	require.NoError(t, err)

	GetAuthTokenFromHttpRequest = wrapped
}

type httpEntryDenialCase struct {
	name          string
	returnedToken *sneatauth.Token
	returnedErr   error
	authRequired  bool
}

var httpEntryDenialCases = []httpEntryDenialCase{
	{
		name:          "reserved UID denied regardless of claims (required auth)",
		returnedToken: &sneatauth.Token{UID: testReservedUID, IsAdmin: true, Issuer: "https://auth.provider"},
		returnedErr:   nil,
		authRequired:  true,
	},
	{
		name:          "reserved UID denied on optional auth",
		returnedToken: &sneatauth.Token{UID: testReservedUID},
		returnedErr:   nil,
		authRequired:  false,
	},
	{
		name:          "blank token UID rejected (required auth)",
		returnedToken: &sneatauth.Token{UID: ""},
		returnedErr:   nil,
		authRequired:  true,
	},
	{
		name:          "blank token UID rejected on optional auth",
		returnedToken: &sneatauth.Token{UID: ""},
		returnedErr:   nil,
		authRequired:  false,
	},
	{
		name:          "whitespace-only token UID rejected (required auth)",
		returnedToken: &sneatauth.Token{UID: "   "},
		returnedErr:   nil,
		authRequired:  true,
	},
	{
		name:          "whitespace-only token UID rejected on optional auth",
		returnedToken: &sneatauth.Token{UID: "   "},
		returnedErr:   nil,
		authRequired:  false,
	},
	{
		name:          "required auth with nil token",
		returnedToken: nil,
		returnedErr:   nil,
		authRequired:  true,
	},
	{
		name:          "underlying provider error",
		returnedToken: nil,
		returnedErr:   errors.New("upstream provider unavailable"),
		authRequired:  true,
	},
	{
		name:          "token accompanied by provider error preserves provider error precedence (required auth)",
		returnedToken: &sneatauth.Token{UID: "ordinary-user-token"},
		returnedErr:   errors.New("upstream provider error takes precedence"),
		authRequired:  true,
	},
	{
		name:          "token accompanied by provider error on optional auth preserves provider error precedence",
		returnedToken: &sneatauth.Token{UID: "ordinary-user-token"},
		returnedErr:   errors.New("upstream provider error takes precedence optional"),
		authRequired:  false,
	},
}

func TestVerifyRequest_DeniedCases(t *testing.T) {
	for _, tc := range httpEntryDenialCases {
		t.Run(tc.name, func(t *testing.T) {
			setupTestExternalVerifier(t, tc.returnedToken, tc.returnedErr)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
			options := verify.Request(verify.AuthenticationRequired(tc.authRequired))

			ctx, err := VerifyRequest(w, r, options)

			// 1. Must return facade.ErrUnauthorized classification
			require.Error(t, err)
			assert.True(t, errors.Is(err, facade.ErrUnauthorized), "expected facade.ErrUnauthorized, got: %v", err)

			// 2. Direct VerifyRequest must NOT write the authentication response
			assert.Equal(t, 0, w.Body.Len(), "VerifyRequest must not write to response body")
			assert.Equal(t, http.StatusOK, w.Code, "VerifyRequest must not set error HTTP status code on ResponseWriter")

			// 3. No token may be accepted into the returned context
			assert.Nil(t, sneatauth.AuthTokenFromContext(ctx), "no token must be present in context on denial")
		})
	}
}

func TestVerifyRequest_SuccessCases(t *testing.T) {
	t.Run("ordinary UID accepted and preserved in context", func(t *testing.T) {
		validToken := &sneatauth.Token{
			UID:     "ordinary-user-123",
			Issuer:  "https://auth.provider",
			IsAdmin: false,
		}
		setupTestExternalVerifier(t, validToken, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
		options := verify.Request(verify.AuthenticationRequired(true))

		ctx, err := VerifyRequest(w, r, options)
		require.NoError(t, err)
		assert.Equal(t, 0, w.Body.Len())

		acceptedToken := sneatauth.AuthTokenFromContext(ctx)
		require.NotNil(t, acceptedToken)
		assert.Equal(t, "ordinary-user-123", acceptedToken.UID)
		assert.Equal(t, "https://auth.provider", acceptedToken.Issuer)
	})

	t.Run("optional auth with nil token stays anonymous", func(t *testing.T) {
		setupTestExternalVerifier(t, nil, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
		options := verify.Request(verify.AuthenticationRequired(false))

		ctx, err := VerifyRequest(w, r, options)
		require.NoError(t, err)
		assert.Equal(t, 0, w.Body.Len())
		assert.Nil(t, sneatauth.AuthTokenFromContext(ctx))
	})
}

func TestVerifyRequestAndCreateUserContext_DeniedCases(t *testing.T) {
	for _, tc := range httpEntryDenialCases {
		t.Run(tc.name, func(t *testing.T) {
			setupTestExternalVerifier(t, tc.returnedToken, tc.returnedErr)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
			options := verify.Request(verify.AuthenticationRequired(tc.authRequired))

			ctxWithUser, err := VerifyRequestAndCreateUserContext(w, r, options)

			// 1. Must return facade.ErrUnauthorized classification
			require.Error(t, err)
			assert.True(t, errors.Is(err, facade.ErrUnauthorized), "expected facade.ErrUnauthorized, got: %v", err)

			// 2. Returns nil USER context
			assert.Nil(t, ctxWithUser, "must return nil user context on denial")

			// 3. Writes HTTP 401 via existing handler
			assert.Equal(t, http.StatusUnauthorized, w.Code, "must write HTTP 401 status")
			assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
			assert.Greater(t, w.Body.Len(), 0, "must write error JSON body")
		})
	}
}

func TestVerifyRequestAndCreateUserContext_SuccessCases(t *testing.T) {
	t.Run("ordinary user creates valid user context", func(t *testing.T) {
		validToken := &sneatauth.Token{
			UID: "ordinary-user-456",
		}
		setupTestExternalVerifier(t, validToken, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
		options := verify.Request(verify.AuthenticationRequired(true))

		ctxWithUser, err := VerifyRequestAndCreateUserContext(w, r, options)
		require.NoError(t, err)
		require.NotNil(t, ctxWithUser)
		require.NotNil(t, ctxWithUser.User())
		assert.Equal(t, "ordinary-user-456", ctxWithUser.User().GetUserID())
		assert.Equal(t, 0, w.Body.Len())
	})

	t.Run("optional auth with nil token returns nil user context without error", func(t *testing.T) {
		setupTestExternalVerifier(t, nil, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
		options := verify.Request(verify.AuthenticationRequired(false))

		ctxWithUser, err := VerifyRequestAndCreateUserContext(w, r, options)
		require.NoError(t, err)
		assert.Nil(t, ctxWithUser)
		assert.Equal(t, 0, w.Body.Len())
	})
}

func TestDirectWrapper_ProviderErrorIdentityBoundary(t *testing.T) {
	policy, err := sneatauth.NewReservedExternalUIDPolicy([]string{testReservedUID})
	require.NoError(t, err)

	reservedVerifier := func(r *http.Request, authRequired bool) (*sneatauth.Token, error) {
		return &sneatauth.Token{UID: testReservedUID}, nil
	}

	wrapper, err := sneatauth.WrapExternalTokenVerifier(reservedVerifier, policy)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)

	// Direct wrapper preserves provider errors.Is identity
	tok, directErr := wrapper(req, true)
	assert.Nil(t, tok)
	require.Error(t, directErr)
	assert.True(t, errors.Is(directErr, sneatauth.ErrReservedExternalUID), "direct wrapper must preserve ErrReservedExternalUID")
	assert.False(t, errors.Is(directErr, facade.ErrUnauthorized), "direct wrapper does not wrap with facade.ErrUnauthorized")

	// VerifyRequest wraps with facade.ErrUnauthorized using %v, which does NOT preserve underlying errors.Is
	origGetter := GetAuthTokenFromHttpRequest
	defer func() { GetAuthTokenFromHttpRequest = origGetter }()
	GetAuthTokenFromHttpRequest = wrapper

	w := httptest.NewRecorder()
	options := verify.Request(verify.AuthenticationRequired(true))
	ctx, apicoreErr := VerifyRequest(w, req, options)
	assert.Nil(t, sneatauth.AuthTokenFromContext(ctx))
	require.Error(t, apicoreErr)
	assert.True(t, errors.Is(apicoreErr, facade.ErrUnauthorized), "VerifyRequest must classify as facade.ErrUnauthorized")
	// Due to apicore's %v formatting, the underlying errors.Is(apicoreErr, sneatauth.ErrReservedExternalUID) is lost
	assert.False(t, errors.Is(apicoreErr, sneatauth.ErrReservedExternalUID), "apicore's %v formatting does not preserve underlying errors.Is")
	assert.True(t, strings.Contains(apicoreErr.Error(), "reserved external UID"), "apicore error string contains provider text")
}

func TestExternalAuth_DTOHeadersAndDALBypassCannotMintAuthority(t *testing.T) {
	origGetter := GetAuthTokenFromHttpRequest
	defer func() { GetAuthTokenFromHttpRequest = origGetter }()

	policy, err := sneatauth.NewReservedExternalUIDPolicy([]string{testReservedUID})
	require.NoError(t, err)

	// Even if verifier is tempted by request headers, it returns a token with reserved UID
	verifier := func(r *http.Request, authRequired bool) (*sneatauth.Token, error) {
		// Verify that external request headers/body do not bypass the wrapper check
		return &sneatauth.Token{
			UID:     testReservedUID,
			Issuer:  r.Header.Get("X-Requested-Issuer"),
			IsAdmin: r.Header.Get("X-Requested-Admin") == "true",
		}, nil
	}

	wrapper, err := sneatauth.WrapExternalTokenVerifier(verifier, policy)
	require.NoError(t, err)
	GetAuthTokenFromHttpRequest = wrapper

	// Client sends arbitrary privileged-looking headers
	req := httptest.NewRequest(http.MethodPost, "/api/resource", strings.NewReader(`{"role":"superuser","uid":"system-reserved-user"}`))
	req.Header.Set("X-System-UID", testReservedUID)
	req.Header.Set("X-Requested-Issuer", "trusted-authority")
	req.Header.Set("X-Requested-Admin", "true")
	req.Header.Set("Authorization", "Bearer forged-privileged-token")

	w := httptest.NewRecorder()
	options := verify.Request(verify.AuthenticationRequired(true), verify.MaximumContentLength(1024))

	// VerifyRequest rejects
	ctx, err := VerifyRequest(w, req, options)
	require.Error(t, err)
	assert.True(t, errors.Is(err, facade.ErrUnauthorized))
	assert.Nil(t, sneatauth.AuthTokenFromContext(ctx))

	// VerifyRequestAndCreateUserContext rejects and writes 401
	w2 := httptest.NewRecorder()
	ctxWithUser, err2 := VerifyRequestAndCreateUserContext(w2, req, options)
	require.Error(t, err2)
	assert.True(t, errors.Is(err2, facade.ErrUnauthorized))
	assert.Nil(t, ctxWithUser)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)

	// SpaceAccessAuthorizedByCaller is a DAL contract opt-out, not authentication.
	// 1. An empty parent gains no authenticated token.
	baseCtx := req.Context()
	dalBypassEmpty := facade.SpaceAccessAuthorizedByCaller(baseCtx)
	require.NotNil(t, dalBypassEmpty)
	require.NotNil(t, dalBypassEmpty.User())
	assert.Equal(t, "", dalBypassEmpty.User().GetUserID(), "DAL bypass user ID is deliberately empty")
	assert.Nil(t, sneatauth.AuthTokenFromContext(dalBypassEmpty), "empty parent context still gains no token")

	// 2. A parent with an existing authenticated token preserves that exact token alongside the empty DAL userID.
	ordinaryToken := &sneatauth.Token{
		UID:    "ordinary-user-789",
		Issuer: "https://auth.provider",
	}
	parentWithToken := sneatauth.NewContextWithAuthToken(baseCtx, ordinaryToken)
	dalBypassWithToken := facade.SpaceAccessAuthorizedByCaller(parentWithToken)
	require.NotNil(t, dalBypassWithToken)
	require.NotNil(t, dalBypassWithToken.User())
	assert.Equal(t, "", dalBypassWithToken.User().GetUserID(), "DAL bypass user ID remains empty")
	survivingToken := sneatauth.AuthTokenFromContext(dalBypassWithToken)
	require.NotNil(t, survivingToken)
	assert.Same(t, ordinaryToken, survivingToken, "SpaceAccessAuthorizedByCaller preserves existing parent context authToken pointer")
	assert.Equal(t, "ordinary-user-789", survivingToken.UID)
}
