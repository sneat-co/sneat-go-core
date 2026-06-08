package core

import (
	"os"
	"testing"
)

func TestIsInProd(t *testing.T) {
	const gaeAppVar = "GAE_APPLICATION" // set by App Engine
	const cloudRunVar = "K_SERVICE"     // set by Cloud Run

	for _, envVar := range []string{gaeAppVar, cloudRunVar} {
		original := os.Getenv(envVar)
		t.Cleanup(func() {
			_ = os.Setenv(envVar, original)
		})
		_ = os.Setenv(envVar, "")
	}

	if IsInProd() {
		t.Error("expected IsInProd() == false when neither GAE_APPLICATION nor K_SERVICE is set")
	}

	_ = os.Setenv(gaeAppVar, "s~my-app")
	if !IsInProd() {
		t.Error("expected IsInProd() == true when GAE_APPLICATION is set")
	}
	_ = os.Setenv(gaeAppVar, "")

	_ = os.Setenv(cloudRunVar, "sneat-app")
	if !IsInProd() {
		t.Error("expected IsInProd() == true when K_SERVICE is set")
	}
}
