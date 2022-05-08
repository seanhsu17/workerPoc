package main

import (
	"context"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/RichardKnop/machinery/v2"
	mongoBackend "github.com/RichardKnop/machinery/v2/backends/mongo"
	pubsubBroker "github.com/RichardKnop/machinery/v2/brokers/gcppubsub"
	brokerIface "github.com/RichardKnop/machinery/v2/brokers/iface"
	redisBroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	machineryConfig "github.com/RichardKnop/machinery/v2/config"
	eagerLock "github.com/RichardKnop/machinery/v2/locks/eager"
	ifaceLock "github.com/RichardKnop/machinery/v2/locks/iface"
	redisLock "github.com/RichardKnop/machinery/v2/locks/redis"
	"github.com/RichardKnop/machinery/v2/log"
	"github.com/RichardKnop/machinery/v2/tasks"

	"github.com/seanhsu17/workerPoc/internal/handler/math"
	"github.com/seanhsu17/workerPoc/internal/handler/math/task"
)

const redisAddr = "127.0.0.1:6379"
const mongoAddr = "mongodb://root:example@localhost:27017/taskresults?authSource=admin"

func initPubsub(cnf *machineryConfig.Config) brokerIface.Broker {
	project := "test"
	pubsubClient, err := pubsub.NewClient(
		context.Background(),
		project,
	)
	defer pubsubClient.Close()
	if err != nil {
		log.ERROR.Println(err)
	}
	topic := "machinery_tasks"
	subscription := "testSub"
	_, err = pubsubClient.CreateTopic(context.Background(), topic)
	if err != nil {
		log.ERROR.Println(err)
	}
	_, err = pubsubClient.CreateSubscription(context.Background(), subscription,
		pubsub.SubscriptionConfig{
			Topic:            pubsubClient.Topic(topic),
			AckDeadline:      10 * time.Second,
			ExpirationPolicy: 25 * time.Hour,
		},
	)
	if err != nil {
		log.ERROR.Println(err)
	}

	broker, _ := pubsubBroker.New(cnf, project, subscription)
	return broker
}

func initRedis(cnf *machineryConfig.Config) brokerIface.Broker {
	return redisBroker.NewGR(cnf, []string{redisAddr}, 0)
}

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
	backend, _ := mongoBackend.New(&cnf)
	var lock ifaceLock.Lock
	lock = redisLock.New(&cnf, []string{redisAddr}, 0, 3)
	lock = eagerLock.New()

	server := machinery.NewServer(&cnf, initRedis(&cnf), backend, lock)

	tasksMap := map[string]interface{}{
		task.AddTask: math.Add,
	}
	err := server.RegisterTasks(tasksMap)
	if err != nil {
		log.ERROR.Println(err)
	}
	err = server.RegisterPeriodicTask("@every 10s", "every-moment", &tasks.Signature{
		Name: task.AddTask,
		Args: []tasks.Arg{
			{
				Type:  "[]int",
				Value: []int{1, 2},
			},
		},
	})
	if err != nil {
		log.ERROR.Println(err)
	}

	log.INFO.Println(err)
	consumerTag := "machinery_worker"

	//cleanup, err := tracers.SetupTracer(consumerTag)
	worker := server.NewWorker(consumerTag, 10)

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
