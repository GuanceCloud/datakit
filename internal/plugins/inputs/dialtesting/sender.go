// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package dialtesting

import (
	"context"
	"fmt"
	"sync"
	"time"

	pt "github.com/GuanceCloud/cliutils/point"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
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
	l.Warnf("Sender is not set correctly. This is empty sender. dataway url: %s", getMaskURL(url))
	return nil
}

// dwSender uses dataway as sender.
type dwSender struct {
	dw *dataway.DialtestingSender
}

func (s *dwSender) send(url string, point *pt.Point) error {
	var dkPts []*dkpt.Point

	if s.dw == nil {
		return fmt.Errorf("sender dw is nil")
	}

	dkPoint, err := influxdb.NewPoint(string(point.Name()), point.InfluxTags(), point.InfluxFields(), point.Time())
	if err != nil {
		return fmt.Errorf("transform v2 point to dk piont error: %w", err)
	}

	dkPts = []*dkpt.Point{{Point: dkPoint}}

	return s.dw.WriteData(url, dkPts)
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
	DefaultWorkerCacheMaxPoints = 100000
)

type jobData struct {
	regionName string
	class      string
	url        string
	pt         *pt.Point
}

// woker collect all points and send points using sender.
type worker struct {
	sender           sender
	maxJobNumber     int              // max job in parallel
	maxJobChanNumber int              // max job chans
	jobChans         chan *jobData    // point to be dealt
	pointCache       []*jobData       // cache point when jobChans is full
	flushInterval    time.Duration    // flush interval to flush cached points
	flushChan        chan interface{} // flush points mannualy
	mu               sync.RWMutex

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
					} else {
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
	select {
	case w.jobChans <- data:
	default:
		w.mu.Lock()

		// drop half of the cached points when the number of the cached points is over DefaultWorkerCacheMaxPoints.
		if len(w.pointCache) > DefaultWorkerCacheMaxPoints {
			w.pointCache = w.pointCache[DefaultWorkerCacheMaxPoints>>1:]
		}

		w.pointCache = append(w.pointCache, data)
		w.mu.Unlock()
		workerCachePointsGauge.WithLabelValues(data.regionName, data.class).Add(1)
	}
	workerJobChanGauge.WithLabelValues("used").Set(float64(len(w.jobChans)))
}

// flush put the cached points into the jobChans. when the jobChans is full, put back into the cache.
func (w *worker) flush() {
	flushedPoints := [][2]string{}
	w.mu.Lock()
	index := 0
out:
	for index < len(w.pointCache) {
		point := w.pointCache[index]
		select {
		case w.jobChans <- point:
			index++
			flushedPoints = append(flushedPoints, [2]string{point.regionName, point.class})
		default:
			break out
		}
	}

	w.pointCache = w.pointCache[index:]
	w.mu.Unlock()
	workerJobChanGauge.WithLabelValues("used").Set(float64(len(w.jobChans)))
	for _, v := range flushedPoints {
		workerCachePointsGauge.WithLabelValues(v[0], v[1]).Add(-1)
	}

	// flush instanly if cache is not empty
	if len(w.pointCache) > 0 {
		select {
		case w.flushChan <- struct{}{}:
		default:
		}
	}
}
