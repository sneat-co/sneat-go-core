package capturer

import (
	"testing"
)

func TestAddErrorLogger_nil(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("panic expected")
		}
	}()
	AddErrorLogger(nil)
}

func TestAddErrorLogger_should_pass(t *testing.T) {
	l := NoOpErrorLogger{}
	AddErrorLogger(l)
	if len(loggers) == 0 {
		t.Fatal("len(loggers) == 0")
	}
}
