package main

import (
	"github.com/RichardKnop/machinery/v2"
	mongoBackend "github.com/RichardKnop/machinery/v2/backends/mongo"
	redisBroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	machineryConfig "github.com/RichardKnop/machinery/v2/config"
	redisLock "github.com/RichardKnop/machinery/v2/locks/redis"
	"github.com/RichardKnop/machinery/v2/log"
	"github.com/RichardKnop/machinery/v2/tasks"

	"github.com/seanhsu17/workerPoc/internal/handler"
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
	broker := redisBroker.New(&cnf, redisAddr, "", "", 0)
	backend, _ := mongoBackend.New(&cnf)
	lock := redisLock.New(&cnf, []string{redisAddr}, 0, 3)
	server := machinery.NewServer(&cnf, broker, backend, lock)

	tasksMap := map[string]interface{}{
		task.AddTask: handler.Add,
	}
	server.RegisterTasks(tasksMap)

	log.INFO.Println()
	consumerTag := "machinery_worker"

	//cleanup, err := tracers.SetupTracer(consumerTag)
	worker := server.NewWorker(consumerTag, 0)

	// Here we inject some custom code for error handling,
	// start and end of task hooks, useful for metrics for example.
	errorHandler := func(err error) {
		log.ERROR.Println("I am an error handler:", err)
	}

	preTaskHandler := func(signature *tasks.Signature) {
		log.INFO.Println("I am a start of task handler for:", signature.Name)
	}

	postTaskHandler := func(signature *tasks.Signature) {
		log.INFO.Println("I am an end of task handler for:", signature.Name)
	}

	worker.SetPostTaskHandler(postTaskHandler)
	worker.SetErrorHandler(errorHandler)
	worker.SetPreTaskHandler(preTaskHandler)

	worker.Launch()
}
