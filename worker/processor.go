package worker

import (
	"context"
	"github.com/hibiken/asynq"
	db "github.com/joefazee/simplebank/db/sqlc"
)

const (
	QueueCritical       = "critical"
	QueueDefault        = "default"
	QueuePriorityHigher = 10
	QueuePriorityLower  = 5
)

type TaskProcessor interface {
	Start() error
	ProcessTaskVerifyEmail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store) TaskProcessor {

	return &RedisTaskProcessor{
		server: asynq.NewServer(redisOpt, asynq.Config{
			Queues: map[string]int{
				QueueCritical: QueuePriorityHigher,
				QueueDefault:  QueuePriorityLower,
			},
		}),
		store: store,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskVerifyEmail)
	return processor.server.Start(mux)
}
