package core

import (
	"os"
)

// IsInProd indicates if app is running in APP ENGINE
func IsInProd() bool {
	return os.Getenv("GAE_APPLICATION") != ""
}
