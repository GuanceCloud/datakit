package io

import (
	"errors"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	tracing_stat_name = "tracingstat"
)

type SpanInfo struct {
	TraceID  int64
	Service  string
	Resource string
	Duration time.Duration
	IsErr    bool
	reStat   bool
}

type TracingStatistic struct {
	Toolkit      string
	TraceId      int64
	Service      string
	Resource     string
	VisitCount   int
	ErrCount     int
	DurationAvg  time.Duration
	StatInterval time.Duration
}

var ErrSendSpanInfoFailed = errors.New("send span information failed")

var (
	statUnit                   = make(map[string]*TracingStatistic)
	spanInfoChan               = make(chan *SpanInfo, 100)
	sendTimeout  time.Duration = 10 * time.Second
	retry        int           = 3
)

func startTracingStatWorker(d time.Duration) {
	go func() {
		tick := time.NewTicker(d)
		for range tick.C {
			spanInfoChan <- &SpanInfo{reStat: true}
		}
	}()
	go func() {
		for sinfo := range spanInfoChan {
			if sinfo.reStat {
				var (
					pts []*Point
					now = time.Now()
				)
				for _, unit := range statUnit {
					unit.DurationAvg /= time.Duration(unit.VisitCount)
					pt, err := MakePoint(tracing_stat_name, nil, map[string]interface{}{
						"toolkit":       unit.Toolkit,
						"trace_id":      unit.TraceId,
						"service":       unit.Service,
						"resource":      unit.Resource,
						"visit_count":   unit.VisitCount,
						"err_count":     unit.ErrCount,
						"duration_avg":  unit.DurationAvg,
						"stat_interval": d,
					}, now)
					if err != nil {
						log.Errorf("make point failed in Tracing Statistic worker, err: %s", err.Error())
						continue
					}
					pts = append(pts, pt)
				}
				if len(pts) != 0 {
					if err := Feed(tracing_stat_name, datakit.Metric, pts, nil); err != nil {
						log.Errorf("io feed points failed in Tracing Statistic worker, err: %s", err.Error())
					}
				}

				statUnit = make(map[string]*TracingStatistic)
			} else {
				key := fmt.Sprintf("%d:%s:%s", sinfo.TraceID, sinfo.Service, sinfo.Resource)
				unit, ok := statUnit[key]
				if ok {
					unit.VisitCount++
					if sinfo.IsErr {
						unit.ErrCount++
					}
					unit.DurationAvg += sinfo.Duration
				} else {
					statUnit[key] = &TracingStatistic{
						TraceId:    sinfo.TraceID,
						Service:    sinfo.Service,
						Resource:   sinfo.Resource,
						VisitCount: 1,
						ErrCount: func() int {
							if sinfo.IsErr {
								return 1
							} else {
								return 0
							}
						}(),
						DurationAvg: sinfo.Duration,
					}
				}
			}
		}
	}()
}

func SendSpanInfo(sinfo *SpanInfo) {
	if sinfo == nil {
		return
	}

	var (
		timeout = time.NewTimer(sendTimeout)
		retry   = retry
	)
	for ; retry > 0; retry-- {
		select {
		case <-timeout.C:
			timeout = time.NewTimer(sendTimeout)
		case spanInfoChan <- sinfo:
			return
		}
	}

	if retry <= 0 {
		log.Error(ErrSendSpanInfoFailed.Error())
	}
}
