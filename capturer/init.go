package capturer

import (
	"context"
	"fmt"
)

// ErrorLogger defines an error logger interface
type ErrorLogger interface {
	LogError(ctx context.Context, err error)
}

var loggers []ErrorLogger

// AddErrorLogger inits with a errLogger
func AddErrorLogger(errLogger ErrorLogger) {
	if errLogger == nil {
		panic("errLogger == nil")
	}
	for _, l := range loggers {
		if l == errLogger {
			panic(fmt.Sprintf("an attempt to add duplicate errLogger: %T: %v", errLogger, errLogger))
		}
	}
	loggers = append(loggers, errLogger)
}
