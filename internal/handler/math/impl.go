package math

import (
	"encoding/json"
	"time"

	"github.com/17media/api/base/ctx"
	"github.com/17media/category-lib/metrics"
	"github.com/17media/category-lib/mongo"
	"github.com/17media/category-lib/taskqueue"

	"github.com/seanhsu17/workerPoc/internal/handler/math/task"
)

var (
	met = metrics.New("handler.math")
)

type impl struct {
	mongo       mongo.Service
	taskMapping taskqueue.TaskHandlerMap
}

func ProvideHandler() Handler {
	im := &impl{}
	im.taskMapping = taskqueue.TaskHandlerMap{
		task.AddTask:    im.Add,
		task.ReduceTask: im.Reduce,
	}
	return im
}

func (im *impl) GetTaskMapping() taskqueue.TaskHandlerMap {
	return im.taskMapping
}

func (im *impl) Add(context ctx.CTX, data []byte) (result []byte, err error) {
	time.Sleep(5 * time.Second)
	payload := task.AddPayload{}
	if err = json.Unmarshal(data, &payload); err != nil {
		context.WithField("err", err).Error("json.Unmarshal failed")
		return
	}
	r := 0
	for _, num := range payload.Nums {
		r += num
	}
	context.Infof("result %d\n", r)
	if result, err = json.Marshal(task.AddResult{Result: r}); err != nil {
		context.WithField("err", err).Error("json.Marshal failed")
		return
	}
	return
}

func (im *impl) Reduce(context ctx.CTX, data []byte) (result []byte, err error) {
	time.Sleep(5 * time.Second)
	payload := task.ReducePayload{}
	if err = json.Unmarshal(data, &payload); err != nil {
		context.WithField("err", err).Error("json.Unmarshal failed")
		return
	}
	r := payload.Base
	for _, num := range payload.Nums {
		r -= num
	}
	context.Infof("result %d\n", r)
	if result, err = json.Marshal(task.ReduceResult{Result: r}); err != nil {
		context.WithField("err", err).Error("json.Marshal failed")
		return
	}
	return
}
