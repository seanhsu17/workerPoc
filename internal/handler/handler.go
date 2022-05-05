package handler

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/hibiken/asynq"

	"github.com/seanhsu17/workerPoc/internal/handler/task"
)

var TaskMapToFunction = map[string]string{
	task.AddTask: "Add",
}

func Add(nums []int) error {
	time.Sleep(5 * time.Second)
	result := 0
	for _, num := range nums {
		result += num
	}
	log.Printf("result %d\n", result)
	return nil
}

func AsyncqAdd(ctx context.Context, t *asynq.Task) error {
	payload := task.AddPayload{}
	json.Unmarshal(t.Payload(), &payload)
	return Add(payload.Nums)
}
