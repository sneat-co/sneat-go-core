package mocks

// MockTxUpdateTeam mocks transactional update
//func MockTxUpdateTeam(t *testing.T, err error) (
//	mocked *mockingo.MockedFunc,
//	mock func(ctx context.Context, tx dal.ReadwriteTransaction, timestamp time.Time, team *models4teamus.TeamDto, key *dal.Key, data []dal.Update, opts ...dal.Precondition) error,
//) {
//	mocked = mockingo.NewMockedFunc(t, "txUpdateTeam")
//	mock = func(ctx context.Context, tx dal.ReadwriteTransaction, timestamp time.Time, team *models4teamus.TeamDto, key *dal.Key, data []dal.Update, opts ...dal.Precondition) error {
//		mocked.Called(
//			mockingo.NewArgument("key", key),
//			mockingo.NewArgument("data", data),
//			mockingo.NewArgument("opts", opts),
//		)
//		return err
//	}
//	return
//}
