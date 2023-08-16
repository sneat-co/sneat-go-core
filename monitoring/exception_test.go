package monitoring

import "testing"

func TestSetExceptionCapturer(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("SetExceptionCapturer() should panic")
			}
		}()
		SetExceptionCapturer(nil)
	})
}
