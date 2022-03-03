package worker

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const Pow10_17 = int64(100_000_000_000_000_000)

var plWkrStats = NewPLWkrStats()

func NewPLWkrStats() *PLWkrStats {
	stats := &PLWkrStats{}
	go func(s *PLWkrStats) {
		ticker := time.NewTicker(time.Second * 45)
		for {
			select {
			case <-ticker.C:
				n := atomic.LoadInt64(&s.taskNum)
				if n >= Pow10_17 {
					// 由 ShowPLWkrStats 处理此处可能导致的进位问题
					atomic.AddInt64(&s.taskNum, -Pow10_17)
					atomic.AddUint64(&s.tNumPow10_17, 1)
				}

			case <-datakit.Exit.Wait():
				return
			}
		}
	}(stats)
	return stats
}

type PLWkrStats struct {
	taskChLen int64

	doFeedTaskNum int64
	taskNum       int64
	tNumPow10_17  uint64
}

func (s PLWkrStats) String() string {
	totalProcessed := strconv.FormatInt(s.taskNum, 10)
	if s.tNumPow10_17 > 0 {
		totalProcessed += fmt.Sprintf("+ %d*10^17", s.tNumPow10_17)
	}
	return fmt.Sprintf(
		"taskCh: %d/%d waitingQueue: %d totalProcessed: %s",
		s.taskChLen, taskChMaxL, s.doFeedTaskNum, totalProcessed,
	)
}

var (
	lastPLWkrStats      = &PLWkrStats{}
	lastPLWkrStatsMutex = sync.Mutex{}
)

func ShowPLWkrStats() PLWkrStats {
	// lock
	lastPLWkrStatsMutex.Lock()
	defer lastPLWkrStatsMutex.Unlock()

	s := PLWkrStats{
		taskChLen: int64(len(taskCh)),
		doFeedTaskNum: atomic.LoadInt64(
			&plWkrStats.doFeedTaskNum),
		taskNum: atomic.LoadInt64(&plWkrStats.taskNum),
	}
	// 可能此时仍未进位
	s.tNumPow10_17 = atomic.LoadUint64(
		&plWkrStats.tNumPow10_17)

	// 处理进位问题
	if lastPLWkrStats.taskNum > s.taskNum &&
		lastPLWkrStats.tNumPow10_17 == s.tNumPow10_17 {
		s.tNumPow10_17++
	}

	lastPLWkrStats = &s

	return s
}

func taskNumIncrease() {
	atomic.AddInt64(&(plWkrStats.taskNum), 1)
}

func taskChFeedNumIncrease() {
	atomic.AddInt64(&(plWkrStats.doFeedTaskNum), 1)
}

func taskChFeedNumDecrease() {
	atomic.AddInt64(&(plWkrStats.doFeedTaskNum), -1)
}
