package facade

import (
	"context"
	"testing"
)

func TestNewContextWithUser(t *testing.T) {
	ctx := NewContextWithUserID(context.Background(), "123")
	userCtx := ctx.User()
	if userCtx.GetUserID() != "123" {
		t.Errorf("userCtx.GetUserIDFromContext() != \"123\": %v", userCtx.GetUserID())
	}
}

func TestGetUserContext(t *testing.T) {
	var ctx context.Context = NewContextWithUserID(context.Background(), "123")
	var key = "abc"
	ctx = context.WithValue(ctx, &key, "def")
	userCtx := GetUserContext(ctx)
	if userCtx == nil {
		t.Fatal("userCtx == nil")
	}
	if userCtx.GetUserID() != "123" {
		t.Errorf("userCtx.GetUserIDFromContext() != \"123\": %v", userCtx.GetUserID())
	}
}

func Test_contextWithUser_User(t *testing.T) {
	userCtx := NewUserContext("123")
	ctx := contextWithUser{user: userCtx}
	if ctx.User() != userCtx {
		t.Error("ctx.User() != userCtx")
	}
}

func TestSpaceAccessAuthorizedByCaller(t *testing.T) {
	type ctxKey struct{}
	base := context.WithValue(context.Background(), ctxKey{}, "preserved")

	got := SpaceAccessAuthorizedByCaller(base)

	if got == nil {
		t.Fatal("SpaceAccessAuthorizedByCaller returned nil")
	}
	if userID := got.User().GetUserID(); userID != "" {
		t.Errorf("got.User().GetUserID() = %q, want empty string", userID)
	}
	if v := got.Value(ctxKey{}); v != "preserved" {
		t.Errorf("got.Value(ctxKey{}) = %v, want %q -- the base context must be preserved", v, "preserved")
	}
}

// TestSpaceAccessAuthorizedByCaller_SpaceWorkerSkipsMembershipGate documents,
// at the facade level, the contract that dal4spaceus relies on: a
// facade.UserContext with GetUserID() == "" is what a space worker's
// membership gate treats as "do not check membership" -- exactly what
// SpaceAccessAuthorizedByCaller returns. The membership gate itself lives in
// sneat-core-modules (spaceus/dal4spaceus), which imports this package, so
// the end-to-end behaviour is covered there
// (see dal4spaceus.TestSpaceWorkerParams_GetRecords); this test pins the
// half of the contract that belongs to sneat-go-core: the returned user
// context's ID is empty and stays empty regardless of the wrapped context.
func TestSpaceAccessAuthorizedByCaller_SpaceWorkerSkipsMembershipGate(t *testing.T) {
	// A stand-in "space worker membership gate" shaped exactly like
	// dal4spaceus's real one: it only enforces membership when GetUserID()
	// is non-empty.
	membershipGateWouldRun := func(userCtx UserContext) bool {
		return userCtx.GetUserID() != ""
	}

	authenticated := NewContextWithUserID(context.Background(), "user-1")
	if !membershipGateWouldRun(authenticated.User()) {
		t.Fatal("sanity check failed: an authenticated context must still trip the membership gate")
	}

	neutral := SpaceAccessAuthorizedByCaller(context.Background())
	if membershipGateWouldRun(neutral.User()) {
		t.Error("SpaceAccessAuthorizedByCaller's context must not trip the space worker's membership gate")
	}
}
