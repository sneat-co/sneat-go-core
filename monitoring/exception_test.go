package monitoring

import "testing"

func TestSetExceptionCapturer(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("SetErrorCapturer() should panic")
			}
		}()
		SetErrorCapturer(nil)
	})
}
