package worker

import (
	"context"
	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeTaskSendVerifyEmail(
		ctx context.Context,
		payload *PayloadSendVerifyEmail,
		opts ...asynq.Option,
	) error
}

type RedisTaskTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {

	client := asynq.NewClient(redisOpt)

	return &RedisTaskTaskDistributor{
		client: client,
	}
}
