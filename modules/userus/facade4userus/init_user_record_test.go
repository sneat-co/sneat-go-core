package facade4userus

import (
	"context"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/mocks4dalgo/mocks4dal"
	"github.com/golang/mock/gomock"
	"github.com/sneat-co/sneat-go-core/facade"
	dbmodels2 "github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/modules/userus/dto4userus"
	"github.com/sneat-co/sneat-go-core/modules/userus/models4userus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_InitUserRecord(t *testing.T) {
	ctx := context.Background()
	type args struct {
		user    facade.User
		request dto4userus.InitUserRecordRequest
	}
	tests := []struct {
		name     string
		args     args
		wantUser models4userus.UserContext
		wantErr  bool
	}{
		{
			name: "should_create_user_record",
			args: args{
				user: models4userus.NewUserContext("test_user_1"),
				request: dto4userus.InitUserRecordRequest{
					AuthProvider: "password",
					Name: &dbmodels2.Name{
						First: "First",
						Last:  "User",
					},
					IanaTimezone: "Europe/Paris",
					Email:        "u1@example.com",
					RemoteClient: dbmodels2.RemoteClientInfo{
						HostOrApp:  "unit-test",
						RemoteAddr: "127.0.0.1",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// SETUP MOCKS BEGINS
			runReadwriteTransaction = func(ctx context.Context, f dal.RWTxWorker, options ...dal.TransactionOption) error {
				mockCtrl := gomock.NewController(t)
				tx := mocks4dal.NewMockReadwriteTransaction(mockCtrl)
				tx.EXPECT().Get(ctx, gomock.Any()).Return(dal.ErrRecordNotFound)
				tx.EXPECT().Insert(ctx, gomock.Any()).Return(nil)
				return f(ctx, tx)
			}
			// SETUP MOCKS ENDS

			// TEST CALL BEGINS
			gotUser, err := InitUserRecord(ctx, tt.args.user, tt.args.request)
			// TEST CALL ENDS

			if (err != nil) != tt.wantErr {
				t.Errorf("initUserRecordTxWorker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.args.request.Email, gotUser.Dto.Email)
		})
	}
}