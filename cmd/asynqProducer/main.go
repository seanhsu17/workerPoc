package main

import (
	"log"

	"github.com/17media/api/base/ctx"
	"github.com/17media/category-lib/taskqueue"
	"github.com/17media/category-lib/taskqueue/asynq"

	"github.com/seanhsu17/workerPoc/internal/handler/math/task"
)

const redisAddr = "127.0.0.1:6379"

func main() {
	config := taskqueue.Config{
		Redis: taskqueue.RedisConfig{
			Addr: redisAddr,
		},
		Asynq: taskqueue.AsynqConfig{
			Worker: taskqueue.AsynqWorkerConfig{
				Concurrency: 10,
				Queues: map[string]int{
					"critical": 6,
					"default":  3,
					"low":      1,
				},
			},
		},
		Mongo: taskqueue.MongoConfig{
			URI: "mongodb://root:example@localhost:27017/?authSource=admin",
		},
	}

	svc, err := asynq.Init(config)

	// ------------------------------------------------------
	// Example 1: Enqueue task to be processed immediately.
	//            Use (*Client).Enqueue method.
	// ------------------------------------------------------

	t, err := task.NewAddPayload([]int{1, 2, 3}...)
	err = svc.Enqueue(ctx.Background(), t)
	if err != nil {
		log.Fatalf("could not enqueue task: %v", err)
	}

	t, err = task.NewReducePayload(10, []int{1, 2, 3}...)
	err = svc.Enqueue(ctx.Background(), t)
	if err != nil {
		log.Fatalf("could not enqueue task: %v", err)
	}
}
