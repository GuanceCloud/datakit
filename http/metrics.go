package http

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

type apiMetric struct {
	api        string
	latency    time.Duration
	statusCode int
	limited    bool
}

type apiStat struct {
	TotalCount     int     `json:"total_count"`
	Limited        int     `json:"limited"`
	LimitedPercent float64 `json:"limited_percent"`

	Status2xx int `json:"2xx"`
	Status3xx int `json:"3xx"`
	Status4xx int `json:"4xx"`
	Status5xx int `json:"5xx"`

	MaxLatency   time.Duration `json:"max_letency"`
	AvgLatency   time.Duration `json:"avg_latency"`
	totalLatency time.Duration
}

type metricQ struct {
	result chan map[string]*apiStat
}

var (
	statsCh = make(chan *apiMetric, 128)
	qch     = make(chan *metricQ, 8)
)

func getMetrics() map[string]*apiStat {
	q := &metricQ{
		result: make(chan map[string]*apiStat),
	}

	defer close(q.result)
	qch <- q

	// Why timeout 1s?
	// because monitor minimal refresh is 1s, we should not block monitor.
	tick := time.NewTicker(time.Second * 1)
	defer tick.Stop()
	select {
	case r := <-q.result:
		return r
	case <-tick.C: // timeout
		return nil
	}
}

func feedMetric(m *apiMetric) {
	select {
	case statsCh <- m:
	default: // unblocking
	}
}

func metrics() {
	apiStats := map[string]*apiStat{}

	for {
		select {
		case s := <-statsCh:
			if s != nil {
				x, ok := apiStats[s.api]
				if !ok {
					x = &apiStat{}
					apiStats[s.api] = x
				}

				if s.limited {
					x.Limited++
				}
				x.TotalCount++
				if s.latency > x.MaxLatency {
					x.MaxLatency = s.latency
				}

				x.totalLatency += s.latency
				x.AvgLatency = x.totalLatency / time.Duration(x.TotalCount)
				x.LimitedPercent = float64(x.Limited) * 100 / float64(x.TotalCount)

				switch s.statusCode / 100 {
				case 2:
					x.Status2xx++
				case 3:
					x.Status3xx++
				case 4:
					x.Status4xx++
				case 5:
					x.Status5xx++
				}
			}

		case <-datakit.Exit.Wait():
			l.Infof("http metrics exit")
			return

		case q := <-qch:
			select {
			case <-q.result: // check if closed
			case q.result <- copyStats(apiStats):
			default: // unblocking
			}
		}
	}
}

func copyStats(from map[string]*apiStat) map[string]*apiStat {
	res := map[string]*apiStat{}
	for k, v := range from {
		res[k] = &apiStat{
			Limited:        v.Limited,
			TotalCount:     v.TotalCount,
			MaxLatency:     v.MaxLatency,
			AvgLatency:     v.AvgLatency,
			LimitedPercent: v.LimitedPercent,
			totalLatency:   v.totalLatency,
			Status2xx:      v.Status2xx,
			Status3xx:      v.Status3xx,
			Status4xx:      v.Status4xx,
			Status5xx:      v.Status5xx,
		}
	}

	return res
}
