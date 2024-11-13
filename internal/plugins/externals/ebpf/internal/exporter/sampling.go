package exporter

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/stats"
)

type sampling struct {
	rate         map[string]float64
	ptsPerMinute map[string]float64

	sync.RWMutex
}

func newSampling(ctx context.Context, c *cfg) *sampling {
	switch {
	case c.samplingRate != "":
		r, rmp := parseRate(
			c.samplingRate, 0.3)
		log.Infof("set samping rate %.2f", r)
		return &sampling{
			rate:         rmp,
			ptsPerMinute: map[string]float64{},
		}
	case c.samplingRatePtsPerMin != "":
		r, rmp := parseRate(
			c.samplingRatePtsPerMin, 1500)
		log.Infof("set samping rate %.2f pts/min", r)

		s := &sampling{
			rate:         map[string]float64{},
			ptsPerMinute: rmp,
		}
		s.autoRate(ctx)
		return s
	default:
		return &sampling{
			rate:         map[string]float64{},
			ptsPerMinute: map[string]float64{},
		}
	}
}

func parseRate(r string, defaultVal float64) (float64, map[string]float64) {
	v, err := strconv.ParseFloat(r, 64)
	if err != nil {
		log.Error("unsupported rate %s, use default rate %f",
			r, defaultVal)
		v = defaultVal
	}
	ret := map[string]float64{}
	for _, c := range point.AllCategories() {
		ret[c.String()] = v
	}
	return v, ret
}

func (s *sampling) sampling(cat string, pts []*point.Point) []*point.Point {
	s.RLock()
	defer s.RUnlock()

	if v, ok := s.rate[cat]; ok && v > 0 {
		result := make([]*point.Point, 0, int(float64(len(pts))*v)+1)
		if n := int(1 / v); n > 1 {
			for i := 0; i < len(pts); i += n {
				result = append(result, pts[i])
			}
			return result
		}
	}
	return pts
}

func (s *sampling) autoRate(ctx context.Context) {
	fn := func() {
		var lastStats map[string]float64
		tk := time.NewTicker(time.Minute)
		for {
			select {
			case <-tk.C:
				lastStats = s.changeRate(lastStats)
			case <-ctx.Done():
				return
			}
		}
	}

	go fn()
}

func (s *sampling) changeRate(lastStats map[string]float64) map[string]float64 {
	s.Lock()
	defer s.Unlock()

	curStats, _ := getTotalPtsMetric()
	rate := calRate(lastStats, curStats, s.ptsPerMinute)
	lastStats = curStats
	for cat, val := range rate {
		s.rate[cat] = val
	}
	log.Infof("rate changed: %v", s.rate)
	return lastStats
}

func calRate(prv, cur, limit map[string]float64) map[string]float64 {
	rate := map[string]float64{}
	for c, v := range cur {
		if prv, ok := prv[c]; ok {
			count := v - prv
			if limit, ok := limit[c]; ok && count > 0 && limit > 0 {
				r := limit / count
				if r > 0.99 {
					r = 1
				}
				if r < 0.01 {
					r = 0.01
				}
				rate[c] = r
			}
		}
	}
	return rate
}

func getTotalPtsMetric() (map[string]float64, error) {
	r := map[string]float64{}
	mfs, err := stats.GetRegistry().Gather()
	if err != nil {
		return nil, err
	}
	for _, m := range mfs {
		if m.GetName() != "dkebpf_exporter_points_total" {
			continue
		}
		for _, v := range m.GetMetric() {
			for _, lb := range v.GetLabel() {
				if *lb.Name == "category" {
					r[*lb.Value] += *(v.GetCounter().Value)
				}
			}
		}
	}
	return r, nil
}
