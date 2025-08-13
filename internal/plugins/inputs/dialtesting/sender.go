// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"context"
	"fmt"
	"sync"
	"time"

	pt "github.com/GuanceCloud/cliutils/point"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

// Sender is used to save points.
type sender interface {
	send(url string, pt *pt.Point) error
	checkToken(token, scheme, host string) (bool, error)
}

// emptySender is used for debug or as default sender.
type emptySender struct{}

func (s *emptySender) checkToken(token, scheme, host string) (bool, error) {
	return true, nil
}

func (s *emptySender) send(url string, pt *pt.Point) error {
	cp.Printf("%s\n", pt.LineProto())
	cp.Infof("# Got 1 point for dataway(%s) | Ctrl+c to exit.\n", url)
	return nil
}

// dwSender uses dataway as sender.
type dwSender struct {
	dw *dataway.DialtestingSender
}

func (s *dwSender) send(url string, point *pt.Point) error {
	if s.dw == nil {
		return fmt.Errorf("sender dw is nil")
	}

	return s.dw.WriteData(url, []*pt.Point{point})
}

func (s *dwSender) checkToken(token, scheme, host string) (bool, error) {
	if s.dw == nil {
		return false, fmt.Errorf("sender dw is nil")
	}

	return s.dw.CheckToken(token, scheme, host)
}

const (
	DefaultWorkerMaxJobNumber   = 10
	DefaultWorkerChannelNumber  = 1000
	DefaultWorkerCacheMaxPoints = 10000
)

type jobData struct {
	regionName string
	class      string
	url        string
	pt         *pt.Point
}

// woker collect all points and send points using sender.
type worker struct {
	sender               sender
	maxJobNumber         int                   // max job in parallel
	maxJobChanNumber     int                   // max job chans
	maxCachePointsNumber int                   // max points number in cache
	jobChans             chan *jobData         // point to be dealt
	pointCache           map[string]*DataCache // cache point when jobChans is full
	flushInterval        time.Duration         // flush interval to flush cached points
	flushChan            chan interface{}      // flush points mannualy
	mu                   sync.RWMutex

	failInfo map[string]int
}

func (w *worker) updateFailInfo(url string, isError bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	prevCount := 0
	if count, ok := w.failInfo[url]; !ok {
		w.failInfo[url] = 0
	} else {
		prevCount = count
	}
	if isError {
		w.failInfo[url] = prevCount + 1
	} else {
		w.failInfo[url] = 0
	}
}

func (w *worker) getFailCount(url string) int {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if count, ok := w.failInfo[url]; ok {
		return count
	} else {
		return 0
	}
}

func (w *worker) init() {
	if w.maxJobNumber <= 0 {
		w.maxJobNumber = DefaultWorkerMaxJobNumber
	}

	w.pointCache = map[string]*DataCache{}

	w.failInfo = map[string]int{}

	if w.sender == nil {
		w.sender = &emptySender{}
	}

	if w.flushInterval == 0 {
		w.flushInterval = 10 * time.Second
	}

	if w.maxJobChanNumber <= 0 {
		w.maxJobChanNumber = DefaultWorkerChannelNumber
	}

	if w.maxCachePointsNumber <= 0 {
		w.maxCachePointsNumber = DefaultWorkerCacheMaxPoints
	}

	w.jobChans = make(chan *jobData, w.maxJobChanNumber)
	w.flushChan = make(chan interface{}, 1)

	workerJobGauge.Set(float64(w.maxJobNumber))
	workerJobChanGauge.WithLabelValues("total").Set(float64(cap(w.jobChans)))
	w.runConsumer()
}

func (w *worker) runConsumer() {
	g := datakit.G("dialtesting_worker")
	for i := 0; i < w.maxJobNumber; i++ {
		g.Go(func(ctx context.Context) error {
			for {
				select {
				case <-datakit.Exit.Wait():
					return nil
				case job := <-w.jobChans:
					workerSendPointsGauge.WithLabelValues(job.regionName, job.class, "sending").Add(1)
					startTime := time.Now()
					if err := w.sender.send(job.url, job.pt); err != nil {
						w.updateFailInfo(job.url, true)
						l.Warnf("send data failed: %s", err.Error())
						workerSendPointsGauge.WithLabelValues(job.regionName, job.class, "failed").Add(1)
						metrics.ErrCountVec.WithLabelValues(inputName, pt.DynamicDWCategory.String()).Inc()
					} else {
						dkio.InputsFeedVec().WithLabelValues(inputName, pt.DynamicDWCategory.String()).Inc()
						dkio.InputsFeedPtsVec().WithLabelValues(inputName, pt.DynamicDWCategory.String()).Observe(float64(1))
						dkio.InputsLastFeedVec().WithLabelValues(inputName, pt.DynamicDWCategory.String()).Set(float64(time.Now().Unix()))
						w.updateFailInfo(job.url, false)
						workerSendPointsGauge.WithLabelValues(job.regionName, job.class, "ok").Add(1)
					}
					workerSendCost.WithLabelValues(job.regionName, job.class).Observe(float64(time.Since(startTime)) / float64(time.Second))
					workerSendPointsGauge.WithLabelValues(job.regionName, job.class, "sending").Add(-1)
					workerJobChanGauge.WithLabelValues("used").Set(float64(len(w.jobChans)))
				}
			}
		})
	}

	g.Go(func(ctx context.Context) error {
		flushTicker := time.NewTicker(w.flushInterval)
		defer flushTicker.Stop()
		for {
			select {
			case <-datakit.Exit.Wait():
				return nil
			case <-w.flushChan:
				w.flush()
			case <-flushTicker.C:
				w.flush()
			}
		}
	})
}

// addPoints add point into the jobChans or pointCache when the jobChans is full.
func (w *worker) addPoints(data *jobData) {
	if data == nil {
		l.Warn("add point nil, ignored")
		return
	}
	cache := w.getCache(data.class)
	select {
	case w.jobChans <- data:
	default:
		cache.Push(data)
	}
	workerJobChanGauge.WithLabelValues("used").Set(float64(len(w.jobChans)))
	workerCachePointsGauge.WithLabelValues(data.regionName, data.class).Set(float64(cache.Len()))
	workerCacheDropPointsGauge.WithLabelValues(data.regionName, data.class).Set(float64(cache.DropCnt()))
}

func (w *worker) getCache(class string) *DataCache {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.pointCache == nil {
		w.pointCache = map[string]*DataCache{}
	}

	if _, ok := w.pointCache[class]; !ok {
		w.pointCache[class] = NewDataCache(w.maxCachePointsNumber)
	}

	return w.pointCache[class]
}

// flush put the cached points into the jobChans. when the jobChans is full, put back into the cache.
func (w *worker) flush() {
	hasCache := false
	defer func() {
		// flush instanly if cache is not empty
		if hasCache {
			select {
			case w.flushChan <- struct{}{}:
			default:
			}
		}
	}()

	for _, cache := range w.pointCache {
		if cache.Len() == 0 {
			continue
		}

		for {
			data, ok := cache.Pop()
			if !ok {
				break
			}
			// exit if job chan is full
			select {
			case w.jobChans <- data:
				workerCachePointsGauge.WithLabelValues(data.regionName, data.class).Set(float64(cache.Len()))
				workerCacheDropPointsGauge.WithLabelValues(data.regionName, data.class).Set(float64(cache.DropCnt()))
			default:
				cache.Push(data)
				hasCache = true
				return
			}
		}
	}
}

type DataCache struct {
	mu      sync.RWMutex
	current int
	maxSize int    // max cache size
	dropCnt uint64 // drop count
	data    []*jobData
}

func NewDataCache(maxSize int) *DataCache {
	if maxSize <= 0 {
		maxSize = 1
	}

	return &DataCache{
		current: -1,
		maxSize: maxSize,
		data:    make([]*jobData, maxSize),
	}
}

func (c *DataCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.current + 1
}

func (c *DataCache) DropCnt() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.dropCnt
}

func (c *DataCache) Push(data *jobData) {
	if data == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.current == c.maxSize-1 {
		c.dropCnt++
	} else {
		c.current++
	}

	c.data[c.current] = data
}

func (c *DataCache) Pop() (*jobData, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.current < 0 {
		return nil, false
	}

	data := c.data[c.current]
	c.current--

	return data, true
}
