package emails

import "testing"

func TestInit(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("panic expected")
		}
	}()
	Init(nil)
}
