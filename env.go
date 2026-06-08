package core

import (
	"os"
)

// IsInProd indicates if app is running in a managed cloud runtime (App Engine or Cloud Run)
func IsInProd() bool {
	return os.Getenv("GAE_APPLICATION") != "" || // App Engine
		os.Getenv("K_SERVICE") != "" // Cloud Run
}
