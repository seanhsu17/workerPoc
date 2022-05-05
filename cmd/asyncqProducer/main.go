package main

import (
	"log"

	"github.com/hibiken/asynq"

	"github.com/seanhsu17/workerPoc/internal/handler/task"
)

const redisAddr = "127.0.0.1:6379"

func main() {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer client.Close()

	// ------------------------------------------------------
	// Example 1: Enqueue task to be processed immediately.
	//            Use (*Client).Enqueue method.
	// ------------------------------------------------------

	t, err := task.NewAddPayload([]int{1, 2, 3}...)
	info, err := client.Enqueue(t)
	if err != nil {
		log.Fatalf("could not enqueue task: %v", err)
	}
	log.Printf("enqueued task: id=%s queue=%s", info.ID, info.Queue)
}
