package facade

import "testing"

func TestAuthUser_GetUserID(t *testing.T) {
	const userID = "123"
	v := AuthUserContext{ID: userID}
	if actual := v.GetUserID(); actual != userID {
		t.Errorf("v.GetUserIDFromContext() != userID: %s != %s", actual, userID)
	}
}

func TestAuthUser_String(t *testing.T) {
	const userID = "123"
	v := AuthUserContext{ID: userID}
	if actual := v.String(); actual != "{id=123}" {
		t.Errorf("v.String() != \"{id=123}\": %v", actual)
	}
}

func TestAuthUser_Validate(t *testing.T) {
	t.Run("empty ID", func(t *testing.T) {
		v := AuthUserContext{}
		if err := v.Validate(); err == nil {
			t.Error("err == nil")
		}
	})
	t.Run("non-empty ID", func(t *testing.T) {
		v := AuthUserContext{ID: "123"}
		if err := v.Validate(); err != nil {
			t.Errorf("err != nil: %v", err)
		}
	})
}
