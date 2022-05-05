package main

import (
	"context"
	"fmt"

	"github.com/RichardKnop/machinery/v2"
	mongoBackend "github.com/RichardKnop/machinery/v2/backends/mongo"
	redisBroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	machineryConfig "github.com/RichardKnop/machinery/v2/config"
	redisLock "github.com/RichardKnop/machinery/v2/locks/redis"
	"github.com/RichardKnop/machinery/v2/log"
	"github.com/RichardKnop/machinery/v2/tasks"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	opentracinglog "github.com/opentracing/opentracing-go/log"

	"github.com/seanhsu17/workerPoc/internal/handler/task"
)

const redisAddr = "127.0.0.1:6379"
const mongoAddr = "mongodb://root:example@localhost:27017/taskresults?authSource=admin"

func main() {
	cnf := machineryConfig.Config{
		DefaultQueue:    "machinery_tasks",
		ResultsExpireIn: 86400,
		ResultBackend:   mongoAddr,
		Redis: &machineryConfig.RedisConfig{
			MaxIdle:                3,
			IdleTimeout:            240,
			ReadTimeout:            15,
			WriteTimeout:           15,
			ConnectTimeout:         15,
			NormalTasksPollPeriod:  1000,
			DelayedTasksPollPeriod: 500,
		},
	}
	broker := redisBroker.NewGR(&cnf, []string{redisAddr}, 0)
	backend, err := mongoBackend.New(&cnf)
	lock := redisLock.New(&cnf, []string{redisAddr}, 0, 3)
	server := machinery.NewServer(&cnf, broker, backend, lock)

	span, ctx := opentracing.StartSpanFromContext(context.Background(), "send")
	defer span.Finish()

	batchID := uuid.New().String()
	span.SetBaggageItem("batch.id", batchID)
	span.LogFields(opentracinglog.String("batch.id", batchID))

	log.INFO.Println("Starting batch:", batchID)
	/*
	 * First, let's try sending a single task
	 */
	addTask0 := tasks.Signature{
		Name: task.AddTask,
		Args: []tasks.Arg{
			{
				Type:  "int",
				Value: 1,
			},
			{
				Type:  "int",
				Value: 1,
			},
		},
	}
	log.INFO.Println("Single task:")

	asyncResult, err := server.SendTaskWithContext(ctx, &addTask0)
	if err != nil {
		log.ERROR.Println(fmt.Errorf("could not send task: %s", err.Error()))
	}
	fmt.Println(asyncResult.GetState())
}
