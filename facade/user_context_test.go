package facade

import (
	"testing"
)

func TestNewUserContext(t *testing.T) {
	userCtx := NewUserContext("123")
	if userCtx == nil {
		t.Fatal("userCtx == nil")
	}
	if userCtx.GetUserID() != "123" {
		t.Errorf("userCtx.GetUserIDFromContext() != \"123\": %v", userCtx.GetUserID())
	}
}

func TestUserContext_GetUserID(t *testing.T) {
	const userID = "123"
	v := UserContext{userID: userID}
	if actual := v.GetUserID(); actual != userID {
		t.Errorf("v.GetUserIDFromContext() != userID: %s != %s", actual, userID)
	}
}

func TestUserContext_String(t *testing.T) {
	const userID = "123"
	v := UserContext{userID: userID}
	if actual := v.String(); actual != "UserContext{id=123}" {
		t.Errorf("v.String() != \"UserContext{id=123}\": %v", actual)
	}
}

func TestUserContext_Validate(t *testing.T) {
	t.Run("empty userID", func(t *testing.T) {
		v := UserContext{}
		if err := v.Validate(); err == nil {
			t.Error("err == nil")
		}
	})
	t.Run("non-empty userID", func(t *testing.T) {
		v := UserContext{userID: "123"}
		if err := v.Validate(); err != nil {
			t.Errorf("err != nil: %v", err)
		}
	})
}
