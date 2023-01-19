package gapi

import (
	"context"
	"database/sql"
	"github.com/golang/mock/gomock"
	mockdb "github.com/joefazee/simplebank/db/mock"
	db "github.com/joefazee/simplebank/db/sqlc"
	"github.com/joefazee/simplebank/pb"
	mockdistributor "github.com/joefazee/simplebank/worker/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestServer_LoginUser(t *testing.T) {

	user, password := createRandomUser(t)

	testCases := []struct {
		name          string
		req           *pb.LoginUserRequest
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rr *pb.LoginUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.LoginUserRequest{
				Username: user.Username,
				Password: password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
				store.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Times(1)
			},
			checkResponse: func(rr *pb.LoginUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, rr)
				require.Equal(t, convertUser(&user), rr.User)
			},
		},
		{
			name: "Invalid password",
			req: &pb.LoginUserRequest{
				Username: user.Username,
				Password: "invalid password",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
				store.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rr *pb.LoginUserResponse, err error) {
				require.NotNil(t, err)
				require.Nil(t, rr)
			},
		},
		{
			name: "NotFound",
			req: &pb.LoginUserRequest{
				Username: "invalid_user",
				Password: password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq("invalid_user")).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)

				store.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rr *pb.LoginUserResponse, err error) {
				require.NotNil(t, err)
				require.Nil(t, rr)
			},
		},
		{
			name: "Internal Error",
			req: &pb.LoginUserRequest{
				Username: user.Username,
				Password: password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)

				store.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rr *pb.LoginUserResponse, err error) {
				require.NotNil(t, err)
				require.Nil(t, rr)
			},
		},

		{
			name: "Invalid data",
			req: &pb.LoginUserRequest{
				Username: "a",
				Password: "a",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rr *pb.LoginUserResponse, err error) {
				require.NotNil(t, err)
				require.Nil(t, rr)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			taskDistributor := mockdistributor.NewMockTaskDistributor(ctrl)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, taskDistributor)
			res, err := server.LoginUser(context.Background(), tc.req)
			tc.checkResponse(res, err)
		})
	}

}
