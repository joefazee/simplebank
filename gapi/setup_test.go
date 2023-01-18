package gapi

import (
	"github.com/hibiken/asynq"
	"github.com/joefazee/simplebank/worker"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/joefazee/simplebank/db/sqlc"
	"github.com/joefazee/simplebank/util"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	redisOpt := asynq.RedisClientOpt{Addr: config.RedisAddress}
	taskDistributor := worker.NewRedisTaskTaskDistributor(redisOpt)

	server, err := NewServer(config, store, taskDistributor)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func createRandomUser(t *testing.T) (user db.User, password string) {

	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	return

}
