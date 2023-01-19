package gapi

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	mockdb "github.com/joefazee/simplebank/db/mock"
	db "github.com/joefazee/simplebank/db/sqlc"
	"github.com/joefazee/simplebank/pb"
	"github.com/joefazee/simplebank/util"
	mockdistributor "github.com/joefazee/simplebank/worker/mock"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserTxParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserTxParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = arg.HashedPassword
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserParams(arg db.CreateUserTxParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}

func TestServer_CreateUser(t *testing.T) {

	user, password := createRandomUser(t)

	testCases := []struct {
		name          string
		req           *pb.CreateUserRequest
		buildStubs    func(store *mockdb.MockStore, taskDistributor *mockdistributor.MockTaskDistributor)
		checkResponse func(rr *pb.CreateUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockdistributor.MockTaskDistributor) {

				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateUserTxResult{User: user}, nil)

				taskDistributor.EXPECT().DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any()).Times(1)
			},
			checkResponse: func(rr *pb.CreateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, rr)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			distributor := mockdistributor.NewMockTaskDistributor(ctrl)
			tc.buildStubs(store, distributor)

			server := newTestServer(t, store, distributor)
			rr, err := server.CreateUser(context.Background(), tc.req)
			tc.checkResponse(rr, err)

		})
	}

}
