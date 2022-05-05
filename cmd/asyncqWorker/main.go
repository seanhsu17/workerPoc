package main

import (
	"log"

	"github.com/hibiken/asynq"

	"github.com/seanhsu17/workerPoc/internal/handler"
	"github.com/seanhsu17/workerPoc/internal/handler/task"
)

const redisAddr = "127.0.0.1:6379"

func main() {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
			// Optionally specify multiple queues with different priority.
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			// See the godoc for other configuration options
		},
	)

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(task.AddTask, handler.AsyncqAdd)
	// ...register other handlers...

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
