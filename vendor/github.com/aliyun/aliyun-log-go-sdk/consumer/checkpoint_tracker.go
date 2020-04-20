package consumerLibrary

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"time"
)

type ConsumerCheckPointTracker struct {
	client                            *ConsumerClient
	defaultFlushCheckPointIntervalSec int64
	tempCheckPoint                    string
	lastPersistentCheckPoint          string
	trackerShardId                    int
	lastCheckTime                     int64
	logger                            log.Logger
}

func initConsumerCheckpointTracker(shardId int, consumerClient *ConsumerClient, logger log.Logger) *ConsumerCheckPointTracker {
	checkpointTracker := &ConsumerCheckPointTracker{
		defaultFlushCheckPointIntervalSec: 60,
		client:                            consumerClient,
		trackerShardId:                    shardId,
		logger:                            logger,
	}
	return checkpointTracker
}

func (checkPointTracker *ConsumerCheckPointTracker) setMemoryCheckPoint(cursor string) {
	checkPointTracker.tempCheckPoint = cursor
}

func (checkPointTracker *ConsumerCheckPointTracker) setPersistentCheckPoint(cursor string) {
	checkPointTracker.lastPersistentCheckPoint = cursor
}

func (checkPointTracker *ConsumerCheckPointTracker) flushCheckPoint() error {
	if checkPointTracker.tempCheckPoint != "" && checkPointTracker.tempCheckPoint != checkPointTracker.lastPersistentCheckPoint {
		if err := checkPointTracker.client.updateCheckPoint(checkPointTracker.trackerShardId, checkPointTracker.tempCheckPoint, true); err != nil {
			return err
		}

		checkPointTracker.lastPersistentCheckPoint = checkPointTracker.tempCheckPoint
	}
	return nil
}

func (checkPointTracker *ConsumerCheckPointTracker) flushCheck() {
	currentTime := time.Now().Unix()
	if currentTime > checkPointTracker.lastCheckTime+checkPointTracker.defaultFlushCheckPointIntervalSec {
		if err := checkPointTracker.flushCheckPoint(); err != nil {
			level.Warn(checkPointTracker.logger).Log("msg", "update checkpoint get error", "error", err)
		} else {
			checkPointTracker.lastCheckTime = currentTime
		}
	}
}

func (checkPointTracker *ConsumerCheckPointTracker) getCheckPoint() string {
	return checkPointTracker.tempCheckPoint
}
