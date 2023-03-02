// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"time"
)

var (
	metricCh = make(chan *APIMetric, 32)
	qch      = make(chan *qAPIStats)
	exitch   = make(chan interface{})
)

type APIMetric struct {
	API        string
	Latency    time.Duration
	StatusCode int
	Limited    bool
}

// APIMetricReporter used to collects API metrics during API handing.
type APIMetricReporter interface {
	Report(*APIMetric) // report these metrics
}

// ReporterImpl used to report API stats
// TODO: We should implemente a default API metric reporter under cliutils.
type ReporterImpl struct{}

func (r *ReporterImpl) Report(m *APIMetric) {
	select {
	case metricCh <- m:
	default: // unblocking
	}
}

type APIStat struct {
	Total int
	LatencyMax,
	LatencyAvg,
	Latency time.Duration

	Status1XX,
	Status2XX,
	Status3XX,
	Status4XX,
	Status5XX int

	Limited int
}

type qAPIStats struct {
	stats chan map[string]*APIStat
}

func GetStats() map[string]*APIStat {
	q := &qAPIStats{stats: make(chan map[string]*APIStat)}

	defer close(q.stats)

	tick := time.NewTicker(time.Second * 1)
	defer tick.Stop()
	qch <- q
	select {
	case r := <-q.stats:
		return r
	case <-tick.C:
		return nil
	}
}

func copyStats(s map[string]*APIStat) map[string]*APIStat {
	res := map[string]*APIStat{}
	for k, v := range s {
		res[k] = &APIStat{
			Total:      v.Total,
			LatencyMax: v.LatencyMax,
			LatencyAvg: v.LatencyAvg,
			Latency:    v.Latency,
			Status1XX:  v.Status1XX,
			Status2XX:  v.Status2XX,
			Status3XX:  v.Status3XX,
			Status4XX:  v.Status4XX,
			Status5XX:  v.Status5XX,
			Limited:    v.Limited,
		}
	}
	return res
}

func StopReporter() {
	select {
	case <-exitch: // closed?
		return
	default:
		close(exitch)
	}
}

func StartReporter() {
	stats := map[string]*APIStat{}
	for {
		select {
		case <-exitch:
			return

		case q := <-qch:
			select {
			case <-q.stats: // is closed?
			case q.stats <- copyStats(stats):
			default: // unblocking
			}

		case m := <-metricCh:
			v, ok := stats[m.API]
			if !ok {
				v = &APIStat{}
				stats[m.API] = v
			}

			v.Total++
			v.Latency += m.Latency
			v.LatencyAvg = v.Latency / time.Duration(v.Total)
			if m.Limited {
				v.Limited++
			}
			if m.Latency > v.LatencyMax {
				v.LatencyMax = m.Latency
			}
			switch m.StatusCode / 100 {
			case 1:
				v.Status1XX++
			case 2:
				v.Status2XX++
			case 3:
				v.Status3XX++
			case 4:
				v.Status4XX++
			case 5:
				v.Status5XX++
			}
		}
	}
}
