package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/17media/configv3"
	log "github.com/17media/logrus"

	"github.com/17media/api/base/ctx"
	"github.com/17media/category-lib/config"
	"github.com/17media/category-lib/taskqueue"
	"github.com/17media/category-lib/taskqueue/asynq"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	"github.com/seanhsu17/workerPoc/internal/handler/math/task"
)

const redisAddr = "127.0.0.1:6379"
const etcdHost = "127.0.0.1:12379"

const (
	cronConfigPath = "/worker/scheduler.yaml"
)

var (
	cronConfig = CronConfig{}
)

func init() {
	if err := config.Register(cronConfigPath, &cronConfig); err != nil {
		log.Panic(fmt.Sprintf("unable to register watcher %v, err %v", cronConfigPath, err))
	}
}

type CronConfig struct {
	CronConfigPayload []byte `yaml:"-"`
}

func (c *CronConfig) Check(data []byte) (interface{}, []string, error) {
	conf := CronConfig{}
	log.Println(data)
	conf.CronConfigPayload = data
	return conf, []string{}, nil
}

// Apply UserConfig
func (c *CronConfig) Apply(v interface{}) {
	*c = v.(CronConfig)
}

func main() {
	flag.Parse()
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdHost},
		DialTimeout: 10 * time.Second,
		// see https://github.com/etcd-io/etcd/issues/9877, solves New won't return error for invalid endpoints
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	})
	if err != nil {
		log.Panic(err)
	}
	ec, err := configv3.NewClientV3(etcdClient, "/configs/envs/dev/discovery", configv3.NoMatchingLogs())
	if err != nil {
		log.Panic(err)
	}

	if err := config.StartWatcher(config.Params{
		Client: ec,
	}); err != nil {
		log.Panic("Fail to start watchers")
	}

	conf := taskqueue.Config{
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
			PeriodicTaskManagerOpts: taskqueue.AsynqPeriodicTaskManagerOpts{
				SyncInterval: 5,
			},
		},
	}

	svc, err := asynq.Init(conf)

	t, err := task.NewAddPayload([]int{1, 2, 3}...)
	err = svc.RegisterPeriodicTask(ctx.Background(), "@every 10s", "test", t)
	if err != nil {
		log.Fatal(err)
	}
	err = svc.RegisterDynamicPeriodicTasks(ctx.Background(), &cronConfig.CronConfigPayload)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		if err := svc.DynamicSchedulerRun(); err != nil {
			log.Fatal(err)
		}
	}()

	if err := svc.SchedulerRun(); err != nil {
		log.Fatal(err)
	}
}
