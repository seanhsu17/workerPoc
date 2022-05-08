package task

import (
	"encoding/json"

	"github.com/17media/category-lib/taskqueue"
)

const (
	AddTask    = "add"
	ReduceTask = "reduce"
)

type AddPayload struct {
	Nums []int
}

type AddResult struct {
	Result int
}

func NewAddPayload(nums ...int) (taskqueue.TaskPayload, error) {
	payload, err := json.Marshal(AddPayload{Nums: nums})
	return taskqueue.TaskPayload{Name: AddTask, Payload: payload}, err
}

type ReducePayload struct {
	Base int
	Nums []int
}

type ReduceResult struct {
	Result int
}

func NewReducePayload(base int, nums ...int) (taskqueue.TaskPayload, error) {
	payload, err := json.Marshal(ReducePayload{Base: base, Nums: nums})
	return taskqueue.TaskPayload{Name: ReduceTask, Payload: payload}, err
}
