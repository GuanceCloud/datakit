package consumerLibrary

import (
	"github.com/aliyun/aliyun-log-go-sdk"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
	"os"
	"sync"
	"time"
)

type ConsumerWorker struct {
	consumerHeatBeat   *ConsumerHeatBeat
	client             *ConsumerClient
	workerShutDownFlag bool
	shardConsumer      map[int]*ShardConsumerWorker
	do                 func(shard int, logGroup *sls.LogGroupList) string
	waitGroup          sync.WaitGroup
	Logger             log.Logger
}

func InitConsumerWorker(option LogHubConfig, do func(int, *sls.LogGroupList) string) *ConsumerWorker {
	logger := logConfig(option)
	consumerClient := initConsumerClient(option, logger)
	consumerHeatBeat := initConsumerHeatBeat(consumerClient, logger)
	consumerWorker := &ConsumerWorker{
		consumerHeatBeat:   consumerHeatBeat,
		client:             consumerClient,
		workerShutDownFlag: false,
		shardConsumer:      make(map[int]*ShardConsumerWorker),
		do:                 do,
		Logger:             logger,
	}
	consumerClient.createConsumerGroup()
	return consumerWorker
}

func (consumerWorker *ConsumerWorker) Start() {
	consumerWorker.waitGroup.Add(1)
	go consumerWorker.run()
}

func (consumerWorker *ConsumerWorker) StopAndWait() {
	level.Info(consumerWorker.Logger).Log("msg", "*** try to exit ***")
	consumerWorker.workerShutDownFlag = true
	consumerWorker.consumerHeatBeat.shutDownHeart()
	consumerWorker.waitGroup.Wait()
	level.Info(consumerWorker.Logger).Log("msg", "consumer worker %v stopped", "consumer name", consumerWorker.client.option.ConsumerName)
}

func (consumerWorker *ConsumerWorker) run() {
	level.Info(consumerWorker.Logger).Log("msg", "consumer worker start", "worker name", consumerWorker.client.option.ConsumerName)
	defer consumerWorker.waitGroup.Done()
	go consumerWorker.consumerHeatBeat.heartBeatRun()

	for !consumerWorker.workerShutDownFlag {
		heldShards := consumerWorker.consumerHeatBeat.getHeldShards()
		lastFetchTime := time.Now().UnixNano() / 1000 / 1000

		for _, shard := range heldShards {
			if consumerWorker.workerShutDownFlag {
				break
			}
			shardConsumer := consumerWorker.getShardConsumer(shard)
			if shardConsumer.getConsumerIsCurrentDoneStatus() == true {
				shardConsumer.consume()
			} else {
				continue
			}
		}
		consumerWorker.cleanShardConsumer(heldShards)
		TimeToSleepInMillsecond(consumerWorker.client.option.DataFetchIntervalInMs, lastFetchTime, consumerWorker.workerShutDownFlag)

	}
	level.Info(consumerWorker.Logger).Log("msg", "consumer worker try to cleanup consumers", "worker name", consumerWorker.client.option.ConsumerName)
	consumerWorker.shutDownAndWait()
}

func (consumerWorker *ConsumerWorker) shutDownAndWait() {
	for {
		time.Sleep(500 * time.Millisecond)
		for shard, consumer := range consumerWorker.shardConsumer {
			if !consumer.isShutDownComplete() {
				consumer.consumerShutDown()
			} else if consumer.isShutDownComplete() {
				delete(consumerWorker.shardConsumer, shard)
			}
		}
		if len(consumerWorker.shardConsumer) == 0 {
			break
		}
	}

}

func (consumerWorker *ConsumerWorker) getShardConsumer(shardId int) *ShardConsumerWorker {
	consumer := consumerWorker.shardConsumer[shardId]
	if consumer != nil {
		return consumer
	}
	consumer = initShardConsumerWorker(shardId, consumerWorker.client, consumerWorker.do, consumerWorker.Logger)
	consumerWorker.shardConsumer[shardId] = consumer
	return consumer

}

func (consumerWorker *ConsumerWorker) cleanShardConsumer(owned_shards []int) {
	for shard, consumer := range consumerWorker.shardConsumer {

		if !Contain(shard, owned_shards) {
			level.Info(consumerWorker.Logger).Log("msg", "try to call shut down for unassigned consumer shard", "shardId", shard)
			consumer.consumerShutDown()
			level.Info(consumerWorker.Logger).Log("msg", "Complete call shut down for unassigned consumer shard", "shardId", shard)
		}

		if consumer.isShutDownComplete() {
			isDeleteShard := consumerWorker.consumerHeatBeat.removeHeartShard(shard)
			if isDeleteShard {
				level.Info(consumerWorker.Logger).Log("msg", "Remove an assigned consumer shard", "shardId", shard)
				delete(consumerWorker.shardConsumer, shard)
			} else {
				level.Info(consumerWorker.Logger).Log("msg", "Remove an assigned consumer shard failed", "shardId", shard)
			}
		}
	}
}

// This function is used to initialize the global log configuration
func logConfig(option LogHubConfig) log.Logger {
	var logger log.Logger

	if option.LogFileName == "" {
		if option.IsJsonType {
			logger = log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
		} else {
			logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
		}
	} else {
		if option.IsJsonType {
			logger = log.NewLogfmtLogger(initLogFlusher(option))
		} else {
			logger = log.NewJSONLogger(initLogFlusher(option))
		}
	}
	switch option.AllowLogLevel {
	case "debug":
		logger = level.NewFilter(logger, level.AllowDebug())
	case "info":
		logger = level.NewFilter(logger, level.AllowInfo())
	case "warn":
		logger = level.NewFilter(logger, level.AllowWarn())
	case "error":
		logger = level.NewFilter(logger, level.AllowError())
	default:
		logger = level.NewFilter(logger, level.AllowInfo())
	}
	logger = log.With(logger, "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	return logger
}

func initLogFlusher(option LogHubConfig) *lumberjack.Logger {
	if option.LogMaxSize == 0 {
		option.LogMaxSize = 10
	}
	if option.LogMaxBackups == 0 {
		option.LogMaxBackups = 10
	}
	return &lumberjack.Logger{
		Filename:   option.LogFileName,
		MaxSize:    option.LogMaxSize,
		MaxBackups: option.LogMaxBackups,
		Compress:   option.LogCompass,
	}
}
