package facade4teamus

import (
	"context"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/mocks4dalgo/mocks4dal"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateTeam(t *testing.T) { // TODO: Implement unit tests
}

func Test_getUniqueTeamID(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	readSession := mocks4dal.NewMockReadSession(mockCtrl)
	readSession.EXPECT().Get(gomock.Any(), gomock.Any()).Return(dal.ErrRecordNotFound)
	teamID, err := getUniqueTeamID(ctx, readSession, "TestCompany LTD")
	assert.NoError(t, err)
	assert.Equal(t, "testcompanyltd", teamID)
}
