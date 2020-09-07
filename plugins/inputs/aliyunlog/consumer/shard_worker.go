package consumerLibrary

import (
	"github.com/aliyun/aliyun-log-go-sdk"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"sync"
	"time"
)

var shutDownLock sync.RWMutex
var consumeStatusLock sync.RWMutex
var consumerTaskLock sync.RWMutex

type ShardConsumerWorker struct {
	client                        *ConsumerClient
	consumerCheckPointTracker     *ConsumerCheckPointTracker
	consumerShutDownFlag          bool
	lastFetchLogGroupList         *sls.LogGroupList
	nextFetchCursor               string
	lastFetchGroupCount           int
	lastFetchtime                 int64
	consumerStatus                string
	process                       func(shard int, logGroup *sls.LogGroupList) string
	shardId                       int
	tempCheckPoint                string
	isCurrentDone                 bool
	logger                        log.Logger
	lastFetchTimeForForceFlushCpt int64
	isFlushCheckpointDone         bool
	rollBackCheckpoint            string
}

func (consumer *ShardConsumerWorker) setConsumerStatus(status string) {
	consumeStatusLock.Lock()
	defer consumeStatusLock.Unlock()
	consumer.consumerStatus = status
}

func (consumer *ShardConsumerWorker) getConsumerStatus() string {
	consumeStatusLock.RLock()
	defer consumeStatusLock.RUnlock()
	return consumer.consumerStatus
}

func initShardConsumerWorker(shardId int, consumerClient *ConsumerClient, do func(shard int, logGroup *sls.LogGroupList) string, logger log.Logger) *ShardConsumerWorker {
	shardConsumeWorker := &ShardConsumerWorker{
		consumerShutDownFlag:          false,
		process:                       do,
		consumerCheckPointTracker:     initConsumerCheckpointTracker(shardId, consumerClient, logger),
		client:                        consumerClient,
		consumerStatus:                INITIALIZING,
		shardId:                       shardId,
		lastFetchtime:                 0,
		isCurrentDone:                 true,
		isFlushCheckpointDone:         true,
		logger:                        logger,
		lastFetchTimeForForceFlushCpt: 0,
		rollBackCheckpoint:            "",
	}
	return shardConsumeWorker
}

func (consumer *ShardConsumerWorker) consume() {
	if consumer.consumerShutDownFlag {
		consumer.setIsFlushCheckpointDoneToFalse()
		go func() {
			// If the data is not consumed, save the tempCheckPoint to the server
			if consumer.getConsumerStatus() == PULL_PROCESSING_DONE {
				consumer.consumerCheckPointTracker.tempCheckPoint = consumer.tempCheckPoint
			} else if consumer.getConsumerStatus() == CONSUME_PROCESSING {
				level.Info(consumer.logger).Log("msg", "Consumption is in progress, waiting for consumption to be completed")
				consumer.setIsFlushCheckpointDoneToTrue()
				return
			}
			err := consumer.consumerCheckPointTracker.flushCheckPoint()
			if err != nil {
				level.Warn(consumer.logger).Log("msg", "Flush checkpoint error，prepare for retry", "error message:", err)
			} else {
				consumer.setConsumerStatus(SHUTDOWN_COMPLETE)
				level.Info(consumer.logger).Log("msg", "shardworker are shut down complete", "shardWorkerId", consumer.shardId)
			}
			consumer.setIsFlushCheckpointDoneToTrue()
		}()
	} else if consumer.getConsumerStatus() == INITIALIZING {
		consumer.setConsumerIsCurrentDoneToFalse()
		go func() {
			cursor, err := consumer.consumerInitializeTask()
			if err != nil {
				consumer.setConsumerStatus(INITIALIZING)
			} else {
				consumer.nextFetchCursor = cursor
				consumer.setConsumerStatus(INITIALIZING_DONE)
			}
			consumer.setConsumerIsCurrentDoneToTrue()
		}()
	} else if consumer.getConsumerStatus() == INITIALIZING_DONE || consumer.getConsumerStatus() == CONSUME_PROCESSING_DONE {
		consumer.setConsumerIsCurrentDoneToFalse()
		consumer.setConsumerStatus(PULL_PROCESSING)
		go func() {
			var isGenerateFetchTask = true
			// throttling control, similar as Java's SDK
			if consumer.lastFetchGroupCount < 100 {
				// The time used here is in milliseconds.
				isGenerateFetchTask = (time.Now().UnixNano()/1e6 - consumer.lastFetchtime) > 500
			} else if consumer.lastFetchGroupCount < 500 {
				isGenerateFetchTask = (time.Now().UnixNano()/1e6 - consumer.lastFetchtime) > 200
			} else if consumer.lastFetchGroupCount < 1000 {
				isGenerateFetchTask = (time.Now().UnixNano()/1e6 - consumer.lastFetchtime) > 50
			}
			if isGenerateFetchTask {
				consumer.lastFetchtime = time.Now().UnixNano() / 1e6
				// Set the logback cursor. If the logs are not consumed, save the logback cursor to the server.
				consumer.tempCheckPoint = consumer.nextFetchCursor

				logGroupList, nextCursor, err := consumer.consumerFetchTask()
				if err != nil {
					consumer.setConsumerStatus(INITIALIZING_DONE)
				} else {
					consumer.lastFetchLogGroupList = logGroupList
					consumer.nextFetchCursor = nextCursor
					consumer.consumerCheckPointTracker.setMemoryCheckPoint(consumer.nextFetchCursor)
					consumer.lastFetchGroupCount = GetLogGroupCount(consumer.lastFetchLogGroupList)
					level.Debug(consumer.logger).Log("shardId", consumer.shardId, "fetch log count", GetLogCount(consumer.lastFetchLogGroupList))
					if consumer.lastFetchGroupCount == 0 {
						consumer.lastFetchLogGroupList = nil
					} else {
						consumer.lastFetchTimeForForceFlushCpt = time.Now().Unix()
					}
					if consumer.lastFetchTimeForForceFlushCpt != 0 && time.Now().Unix()-consumer.lastFetchTimeForForceFlushCpt > 30 {
						err := consumer.consumerCheckPointTracker.flushCheckPoint()
						if err != nil {
							level.Warn(consumer.logger).Log("msg", "Failed to save the final checkpoint", "error:", err)
						} else {
							consumer.lastFetchTimeForForceFlushCpt = 0
						}

					}
					consumer.setConsumerStatus(PULL_PROCESSING_DONE)
				}
			} else {
				level.Debug(consumer.logger).Log("msg", "Pull Log Current Limitation and Re-Pull Log")
				consumer.setConsumerStatus(INITIALIZING_DONE)
			}
			consumer.setConsumerIsCurrentDoneToTrue()
		}()
	} else if consumer.getConsumerStatus() == PULL_PROCESSING_DONE {
		consumer.setConsumerIsCurrentDoneToFalse()
		consumer.setConsumerStatus(CONSUME_PROCESSING)
		go func() {
			rollBackCheckpoint := consumer.consumerProcessTask()
			if rollBackCheckpoint != "" {
				consumer.nextFetchCursor = rollBackCheckpoint
				level.Info(consumer.logger).Log("msg", "Checkpoints set for users have been reset", "shardWorkerId", consumer.shardId, "rollBackCheckpoint", rollBackCheckpoint)
			}
			consumer.lastFetchLogGroupList = nil
			consumer.setConsumerStatus(CONSUME_PROCESSING_DONE)
			consumer.setConsumerIsCurrentDoneToTrue()
		}()
	}

}

func (consumer *ShardConsumerWorker) consumerShutDown() {
	consumer.consumerShutDownFlag = true
	if !consumer.isShutDownComplete() {
		if consumer.getIsFlushCheckpointDoneStatus() == true {
			consumer.consume()
		} else {
			return
		}
	}
}

func (consumer *ShardConsumerWorker) isShutDownComplete() bool {
	return consumer.getConsumerStatus() == SHUTDOWN_COMPLETE
}

func (consumer *ShardConsumerWorker) setConsumerIsCurrentDoneToFalse() {
	consumerTaskLock.Lock()
	defer consumerTaskLock.Unlock()
	consumer.isCurrentDone = false

}

func (consumer *ShardConsumerWorker) setConsumerIsCurrentDoneToTrue() {
	consumerTaskLock.Lock()
	defer consumerTaskLock.Unlock()
	consumer.isCurrentDone = true
}

func (consumer *ShardConsumerWorker) getConsumerIsCurrentDoneStatus() bool {
	consumerTaskLock.RLock()
	defer consumerTaskLock.RUnlock()
	return consumer.isCurrentDone
}

func (consumer *ShardConsumerWorker) setIsFlushCheckpointDoneToFalse() {
	shutDownLock.Lock()
	defer shutDownLock.Unlock()
	consumer.isFlushCheckpointDone = false
}

func (consumer *ShardConsumerWorker) setIsFlushCheckpointDoneToTrue() {
	shutDownLock.Lock()
	defer shutDownLock.Unlock()
	consumer.isFlushCheckpointDone = true
}

func (consumer *ShardConsumerWorker) getIsFlushCheckpointDoneStatus() bool {
	shutDownLock.RLock()
	defer shutDownLock.RUnlock()
	return consumer.isFlushCheckpointDone
}
