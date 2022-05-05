package task

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	AddTask = "add"
)

type AddPayload struct {
	Nums []int
}

func NewAddPayload(nums ...int) (*asynq.Task, error) {
	payload, err := json.Marshal(AddPayload{Nums: nums})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(AddTask, payload), nil
}
