package mock_extension

import (
	"testing"

	"github.com/sneat-co/sneat-go-core/coretypes"
	"go.uber.org/mock/gomock"
)

func TestMockConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockConfig(ctrl)

	mock.EXPECT().ID().Return(coretypes.ExtID("ext1"))
	mock.EXPECT().Register(gomock.Any())

	if mock.ID() != coretypes.ExtID("ext1") {
		t.Fatal("expected ext1")
	}
	mock.Register(nil)
}
