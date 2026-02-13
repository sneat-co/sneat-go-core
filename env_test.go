package core

import (
	"os"
	"testing"
)

func TestIsInProd(t *testing.T) {
	const envVar = "GAE_APPLICATION"

	original := os.Getenv(envVar)
	t.Cleanup(func() {
		_ = os.Setenv(envVar, original)
	})

	_ = os.Setenv(envVar, "")
	if IsInProd() {
		t.Error("expected IsInProd() == false when GAE_APPLICATION is empty")
	}

	_ = os.Setenv(envVar, "s~my-app")
	if !IsInProd() {
		t.Error("expected IsInProd() == true when GAE_APPLICATION is set")
	}
}
