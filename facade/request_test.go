package facade

import "testing"

func TestIdRequest_Validate(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		request := IDRequest{}
		if err := request.Validate(); err == nil {
			t.Fatal("expected to get error for empty request")
		}
	})
	t.Run("valid", func(t *testing.T) {
		request := IDRequest{ID: "123"}
		if err := request.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
