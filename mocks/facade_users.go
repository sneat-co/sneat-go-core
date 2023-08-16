package mocks

import (
	"context"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go/src/modules/userus/facade4userus"
	"github.com/sneat-co/sneat-go/src/modules/userus/models4userus"
	"github.com/sneat-team/mockingo"
	"testing"
	"time"
)

var usersByID map[string]*models4userus.UserDto

//func MockGetUserByID(id string, user models2spotbuddies.UserDto) MockedFunc {
//	if usersByID == nil {
//		usersByID = make(map[string]*models2spotbuddies.UserDto, 1)
//	}
//	usersByID[id] = &user
//	api4meetingus.GetUserByID = func(ctx context.Context, fs *firestore.Client, id string) (user models2spotbuddies.UserDto, err error) {
//		calledGetUserByID++
//		if user, ok := usersByID[id]; !ok {
//			return *user, fmt.Errorf("user not found by id: %v", id)
//		}
//		return user, err
//	}
//	return MockedFunc{p: &calledGetUserByID}
//}

// MockTxGetUserByID mocks getting user by ContactID
func MockTxGetUserByID(t *testing.T, err error, id string, user models4userus.UserDto) *mockingo.MockedFunc {
	if usersByID == nil {
		usersByID = make(map[string]*models4userus.UserDto, 1)
	}
	usersByID[id] = &user
	mockedFunc := mockingo.NewMockedFunc(t, "txGetUserByID")
	facade4userus.TxGetUserByID = func(ctx context.Context, transaction dal.ReadwriteTransaction, user dal.Record) error {
		mockedFunc.Called(mockingo.NewArgument("id", user.Key().ID))
		user2 := user.Data().(*models4userus.UserDto)
		*user2 = *usersByID[id]
		return err
	}
	return mockedFunc
}

// MockTxUpdateUser mocks updating user
func MockTxUpdateUser(t *testing.T, err error) *mockingo.MockedFunc {
	var (
		mockedFunc = mockingo.NewMockedFunc(t, "txUpdateUser")
	)
	facade4userus.TxUpdateUser = func(ctx context.Context, _ dal.ReadwriteTransaction, timestamp time.Time, userKey *dal.Key, data []dal.Update, opts ...dal.Precondition) error {
		mockedFunc.Called(mockingo.NewArgument("userKey", userKey))
		return err
	}
	return mockedFunc
}
