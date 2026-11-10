package sneatauth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReservedExternalUIDPolicy_InvalidInputs(t *testing.T) {
	tests := []struct {
		name  string
		input []string
	}{
		{name: "nil slice", input: nil},
		{name: "empty slice", input: []string{}},
		{name: "single empty string", input: []string{""}},
		{name: "single space", input: []string{" "}},
		{name: "multiple spaces", input: []string{"   "}},
		{name: "whitespace tab and newline", input: []string{"\t\n"}},
		{name: "valid followed by empty", input: []string{"uid1", ""}},
		{name: "valid followed by whitespace", input: []string{"uid1", "   "}},
		{name: "empty followed by valid", input: []string{"", "uid2"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			policy, err := NewReservedExternalUIDPolicy(tc.input)
			assert.Nil(t, policy)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, ErrInvalidExternalAuthConfig), "expected ErrInvalidExternalAuthConfig, got: %v", err)
		})
	}
}

func TestNewReservedExternalUIDPolicy_Deduplication(t *testing.T) {
	policy, err := NewReservedExternalUIDPolicy([]string{"uid1", "uid2", "uid1", "uid2", "uid3"})
	require.NoError(t, err)
	require.NotNil(t, policy)

	assert.Equal(t, 3, len(policy.reserved))
	assert.True(t, errors.Is(policy.CheckUID("uid1"), ErrReservedExternalUID))
	assert.True(t, errors.Is(policy.CheckUID("uid2"), ErrReservedExternalUID))
	assert.True(t, errors.Is(policy.CheckUID("uid3"), ErrReservedExternalUID))
	assert.NoError(t, policy.CheckUID("uid4"))
}

func TestReservedExternalUIDPolicy_DefensiveCopy(t *testing.T) {
	input := []string{"reserved-1", "reserved-2"}
	policy, err := NewReservedExternalUIDPolicy(input)
	require.NoError(t, err)
	require.NotNil(t, policy)

	// Mutate the original slice
	input[0] = "mutated-1"
	input[1] = "mutated-2"

	// Original reserved values must remain reserved
	assert.True(t, errors.Is(policy.CheckUID("reserved-1"), ErrReservedExternalUID))
	assert.True(t, errors.Is(policy.CheckUID("reserved-2"), ErrReservedExternalUID))

	// Mutated values must NOT become reserved
	assert.NoError(t, policy.CheckUID("mutated-1"))
	assert.NoError(t, policy.CheckUID("mutated-2"))
}

func TestReservedExternalUIDPolicy_NilAndZero(t *testing.T) {
	t.Run("nil pointer policy", func(t *testing.T) {
		var nilPolicy *ReservedExternalUIDPolicy
		err := nilPolicy.CheckUID("some-uid")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidExternalAuthConfig))

		errEmpty := nilPolicy.CheckUID("")
		assert.Error(t, errEmpty)
		assert.True(t, errors.Is(errEmpty, ErrInvalidExternalAuthConfig))
	})

	t.Run("zero value struct policy", func(t *testing.T) {
		zeroPolicy := ReservedExternalUIDPolicy{}
		err := zeroPolicy.CheckUID("some-uid")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidExternalAuthConfig))

		errEmpty := zeroPolicy.CheckUID("")
		assert.Error(t, errEmpty)
		assert.True(t, errors.Is(errEmpty, ErrInvalidExternalAuthConfig))
	})

	t.Run("zero value pointer policy", func(t *testing.T) {
		zeroPtrPolicy := &ReservedExternalUIDPolicy{}
		err := zeroPtrPolicy.CheckUID("some-uid")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidExternalAuthConfig))
	})
}

func TestReservedExternalUIDPolicy_ExactMatching(t *testing.T) {
	policy, err := NewReservedExternalUIDPolicy([]string{"system-user", "  spaced-id  ", "CaseSensitive"})
	require.NoError(t, err)
	require.NotNil(t, policy)

	// Exact matches are rejected as reserved
	assert.True(t, errors.Is(policy.CheckUID("system-user"), ErrReservedExternalUID))
	assert.True(t, errors.Is(policy.CheckUID("  spaced-id  "), ErrReservedExternalUID))
	assert.True(t, errors.Is(policy.CheckUID("CaseSensitive"), ErrReservedExternalUID))

	// Case variants are distinct and not matched
	assert.NoError(t, policy.CheckUID("SYSTEM-USER"))
	assert.NoError(t, policy.CheckUID("casesensitive"))
	assert.NoError(t, policy.CheckUID("CASESENSITIVE"))

	// Prefixes and suffixes are distinct
	assert.NoError(t, policy.CheckUID("system-user-extended"))
	assert.NoError(t, policy.CheckUID("sys:system-user"))

	// Whitespace-trimmed variant is distinct from untrimmed whitespace-wrapped entry
	assert.NoError(t, policy.CheckUID("spaced-id"))

	// Empty and whitespace-only UID inputs fail with ErrInvalidExternalUID
	assert.True(t, errors.Is(policy.CheckUID(""), ErrInvalidExternalUID))
	assert.True(t, errors.Is(policy.CheckUID("   "), ErrInvalidExternalUID))
	assert.True(t, errors.Is(policy.CheckUID("\t\n"), ErrInvalidExternalUID))

	// Ordinary non-reserved UID succeeds
	assert.NoError(t, policy.CheckUID("ordinary-user-42"))
}

func TestWrapExternalTokenVerifier_ConstructorValidation(t *testing.T) {
	validPolicy, err := NewReservedExternalUIDPolicy([]string{"sys-1"})
	require.NoError(t, err)
	dummyVerifier := func(r *http.Request, authRequired bool) (*Token, error) {
		return &Token{UID: "user-1"}, nil
	}

	t.Run("nil verifier", func(t *testing.T) {
		wrapper, err := WrapExternalTokenVerifier(nil, validPolicy)
		assert.Nil(t, wrapper)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidExternalAuthConfig))
	})

	t.Run("nil policy", func(t *testing.T) {
		wrapper, err := WrapExternalTokenVerifier(dummyVerifier, nil)
		assert.Nil(t, wrapper)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidExternalAuthConfig))
	})

	t.Run("zero-value policy pointer", func(t *testing.T) {
		wrapper, err := WrapExternalTokenVerifier(dummyVerifier, &ReservedExternalUIDPolicy{})
		assert.Nil(t, wrapper)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidExternalAuthConfig))
	})

	t.Run("WrapVerifier convenience method valid", func(t *testing.T) {
		wrapper, err := validPolicy.WrapVerifier(dummyVerifier)
		assert.NoError(t, err)
		assert.NotNil(t, wrapper)
	})

	t.Run("WrapVerifier convenience method on nil policy", func(t *testing.T) {
		var nilP *ReservedExternalUIDPolicy
		wrapper, err := nilP.WrapVerifier(dummyVerifier)
		assert.Nil(t, wrapper)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidExternalAuthConfig))
	})

	t.Run("WrapVerifier convenience method on zero-value policy", func(t *testing.T) {
		zeroP := &ReservedExternalUIDPolicy{}
		wrapper, err := zeroP.WrapVerifier(dummyVerifier)
		assert.Nil(t, wrapper)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidExternalAuthConfig))
	})
}

func TestWrapExternalTokenVerifier_ErrorPrecedenceOverToken(t *testing.T) {
	policy, err := NewReservedExternalUIDPolicy([]string{"sys-1"})
	require.NoError(t, err)

	providerErr := errors.New("upstream provider error")
	tokenWithErrorVerifier := func(r *http.Request, authRequired bool) (*Token, error) {
		// Even if a verifier returns both a token and an error, error must take precedence
		return &Token{UID: "user-1"}, providerErr
	}

	wrapper, err := WrapExternalTokenVerifier(tokenWithErrorVerifier, policy)
	require.NoError(t, err)
	require.NotNil(t, wrapper)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	tok, callErr := wrapper(req, true)
	assert.Nil(t, tok)
	assert.Same(t, providerErr, callErr, "must return exact SAME error object")
}

func TestWrapExternalTokenVerifier_OptionalNilToken(t *testing.T) {
	policy, err := NewReservedExternalUIDPolicy([]string{"sys-1"})
	require.NoError(t, err)

	nilVerifier := func(r *http.Request, authRequired bool) (*Token, error) {
		return nil, nil
	}

	wrapper, err := WrapExternalTokenVerifier(nilVerifier, policy)
	require.NoError(t, err)
	require.NotNil(t, wrapper)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	tok, callErr := wrapper(req, false)
	assert.Nil(t, tok)
	assert.NoError(t, callErr)
}

func TestWrapExternalTokenVerifier_RequiredNilToken(t *testing.T) {
	policy, err := NewReservedExternalUIDPolicy([]string{"sys-1"})
	require.NoError(t, err)

	nilVerifier := func(r *http.Request, authRequired bool) (*Token, error) {
		return nil, nil
	}

	wrapper, err := WrapExternalTokenVerifier(nilVerifier, policy)
	require.NoError(t, err)
	require.NotNil(t, wrapper)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	tok, callErr := wrapper(req, true)
	assert.Nil(t, tok)
	assert.Error(t, callErr)
	assert.True(t, errors.Is(callErr, ErrInvalidExternalUID))
}

func TestWrapExternalTokenVerifier_BlankOrWhitespaceUID(t *testing.T) {
	policy, err := NewReservedExternalUIDPolicy([]string{"sys-1"})
	require.NoError(t, err)

	blankUIDs := []string{"", " ", "   ", "\t\n"}

	for _, blank := range blankUIDs {
		verifier := func(r *http.Request, authRequired bool) (*Token, error) {
			return &Token{UID: blank}, nil
		}

		wrapper, err := WrapExternalTokenVerifier(verifier, policy)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		// Both required and optional auth must reject blank/whitespace UID with ErrInvalidExternalUID
		t.Run("authRequired=true, blank="+blank, func(t *testing.T) {
			tok, callErr := wrapper(req, true)
			assert.Nil(t, tok)
			assert.True(t, errors.Is(callErr, ErrInvalidExternalUID))
		})

		t.Run("authRequired=false, blank="+blank, func(t *testing.T) {
			tok, callErr := wrapper(req, false)
			assert.Nil(t, tok)
			assert.True(t, errors.Is(callErr, ErrInvalidExternalUID))
		})
	}
}

func TestWrapExternalTokenVerifier_ReservedUIDRegardlessClaims(t *testing.T) {
	policy, err := NewReservedExternalUIDPolicy([]string{"sys-admin-uid"})
	require.NoError(t, err)

	reservedToken := &Token{
		UID:      "sys-admin-uid",
		Issuer:   "https://securetoken.google.com/test-project",
		IsAdmin:  true,
		Original: map[string]any{"admin": true, "role": "system_owner"},
	}

	verifier := func(r *http.Request, authRequired bool) (*Token, error) {
		return reservedToken, nil
	}

	wrapper, err := WrapExternalTokenVerifier(verifier, policy)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	t.Run("required auth rejects reserved UID", func(t *testing.T) {
		tok, callErr := wrapper(req, true)
		assert.Nil(t, tok)
		assert.True(t, errors.Is(callErr, ErrReservedExternalUID))
	})

	t.Run("optional auth rejects reserved UID", func(t *testing.T) {
		tok, callErr := wrapper(req, false)
		assert.Nil(t, tok)
		assert.True(t, errors.Is(callErr, ErrReservedExternalUID))
	})
}

func TestWrapExternalTokenVerifier_OrdinarySuccess_PreservesSamePointerAndFields(t *testing.T) {
	policy, err := NewReservedExternalUIDPolicy([]string{"sys-admin-uid"})
	require.NoError(t, err)

	expectedToken := &Token{
		UID:      "ordinary-user-777",
		Issuer:   "https://securetoken.google.com/test-project",
		IsAdmin:  true,
		Original: map[string]any{"custom_claim": "test_value"},
	}

	verifier := func(r *http.Request, authRequired bool) (*Token, error) {
		return expectedToken, nil
	}

	wrapper, err := WrapExternalTokenVerifier(verifier, policy)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	tok, callErr := wrapper(req, true)
	require.NoError(t, callErr)
	assert.Same(t, expectedToken, tok, "must return exact SAME token pointer")
	assert.Equal(t, "ordinary-user-777", tok.UID)
	assert.Equal(t, "https://securetoken.google.com/test-project", tok.Issuer)
	assert.True(t, tok.IsAdmin)
	assert.Equal(t, map[string]any{"custom_claim": "test_value"}, tok.Original)
}

func TestWrapExternalTokenVerifier_RequestAndFlagForwarding(t *testing.T) {
	policy, err := NewReservedExternalUIDPolicy([]string{"sys-admin-uid"})
	require.NoError(t, err)

	for _, expectedFlag := range []bool{true, false} {
		testName := "authRequired=false"
		if expectedFlag {
			testName = "authRequired=true"
		}
		t.Run(testName, func(t *testing.T) {
			var capturedReq *http.Request
			var capturedAuthRequired bool
			callCount := 0

			verifier := func(r *http.Request, authRequired bool) (*Token, error) {
				callCount++
				capturedReq = r
				capturedAuthRequired = authRequired
				return &Token{UID: "user-1"}, nil
			}

			wrapper, err := WrapExternalTokenVerifier(verifier, policy)
			require.NoError(t, err)

			targetReq := httptest.NewRequest(http.MethodPost, "/api/endpoint", nil)
			tok, err := wrapper(targetReq, expectedFlag)
			require.NoError(t, err)
			assert.NotNil(t, tok)

			assert.Equal(t, 1, callCount, "captured verifier must be called exactly once")
			assert.Same(t, targetReq, capturedReq, "captured verifier must receive identical request pointer")
			assert.Equal(t, expectedFlag, capturedAuthRequired, "captured verifier must receive identical authRequired flag")
		})
	}
}

func TestWrapExternalTokenVerifier_RepeatedCallsEvaluatedPerInvocation(t *testing.T) {
	policy, err := NewReservedExternalUIDPolicy([]string{"sys-admin-uid"})
	require.NoError(t, err)

	tokensToReturn := []*Token{
		{UID: "sys-admin-uid"}, // Call 1: reserved
		{UID: "normal-user-1"}, // Call 2: ordinary
		{UID: "sys-admin-uid"}, // Call 3: reserved again
		{UID: "normal-user-2"}, // Call 4: ordinary
	}
	index := 0

	verifier := func(r *http.Request, authRequired bool) (*Token, error) {
		tok := tokensToReturn[index]
		index++
		return tok, nil
	}

	wrapper, err := WrapExternalTokenVerifier(verifier, policy)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Call 1
	tok1, err1 := wrapper(req, true)
	assert.Nil(t, tok1)
	assert.True(t, errors.Is(err1, ErrReservedExternalUID))

	// Call 2
	tok2, err2 := wrapper(req, true)
	assert.NoError(t, err2)
	assert.NotNil(t, tok2)
	assert.Equal(t, "normal-user-1", tok2.UID)

	// Call 3
	tok3, err3 := wrapper(req, true)
	assert.Nil(t, tok3)
	assert.True(t, errors.Is(err3, ErrReservedExternalUID))

	// Call 4
	tok4, err4 := wrapper(req, true)
	assert.NoError(t, err4)
	assert.NotNil(t, tok4)
	assert.Equal(t, "normal-user-2", tok4.UID)
}

func TestReservedExternalUIDPolicy_ConcurrentReads(t *testing.T) {
	policy, err := NewReservedExternalUIDPolicy([]string{"sys-admin-uid", "sys-root-uid"})
	require.NoError(t, err)

	verifier := func(r *http.Request, authRequired bool) (*Token, error) {
		// Verifier returning a regular token
		return &Token{UID: "user-regular"}, nil
	}

	wrapper, err := WrapExternalTokenVerifier(verifier, policy)
	require.NoError(t, err)

	const goroutines = 20
	const iterations = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(workerID int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			for j := 0; j < iterations; j++ {
				// Concurrent policy checks
				assert.True(t, errors.Is(policy.CheckUID("sys-admin-uid"), ErrReservedExternalUID))
				assert.True(t, errors.Is(policy.CheckUID("sys-root-uid"), ErrReservedExternalUID))
				assert.NoError(t, policy.CheckUID("user-regular"))
				assert.True(t, errors.Is(policy.CheckUID(""), ErrInvalidExternalUID))

				// Concurrent wrapper checks
				tok, callErr := wrapper(req, true)
				assert.NoError(t, callErr)
				assert.NotNil(t, tok)
				assert.Equal(t, "user-regular", tok.UID)
			}
		}(i)
	}

	wg.Wait()
}
