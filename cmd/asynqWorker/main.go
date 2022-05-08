package main

import (
	"log"

	"github.com/17media/api/base/ctx"
	"github.com/17media/category-lib/taskqueue"
	"github.com/17media/category-lib/taskqueue/asynq"

	"github.com/seanhsu17/workerPoc/internal/handler/math"
)

const redisAddr = "127.0.0.1:6379"
const etcdHost = "127.0.0.1:12379"

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
	if err != nil {
		log.Panic(err)
	}

	mh := math.ProvideHandler()
	if err := svc.RegisterTasks(ctx.Background(), mh.GetTaskMapping()); err != nil {
		log.Panic(err)
	}

	if err := svc.WorkerRun(); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
