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
	Toolkit  string
	Project  string
	Version  string
	Service  string
	Resource string
	Duration time.Duration
	IsEntry  bool
	IsErr    bool
	reStat   bool
}

type TracingStatistic struct {
	Toolkit      string
	Project      string
	Version      string
	Service      string
	Resource     string
	RequestCount int
	ErrCount     int
	DurationAvg  int64
}

var ErrSendSpanInfoFailed = errors.New("send span information failed")

var (
	statUnit                   = make(map[string]*TracingStatistic)
	spanInfoChan               = make(chan *SpanInfo, 100)
	sendTimeout  time.Duration = time.Second
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
					unit.DurationAvg /= int64(unit.RequestCount)
					pt, err := MakePoint(tracing_stat_name,
						map[string]string{
							"toolkit":  unit.Toolkit,
							"project":  unit.Project,
							"version":  unit.Version,
							"service":  unit.Service,
							"resource": unit.Resource,
						},
						map[string]interface{}{
							"request_count": unit.RequestCount,
							"err_count":     unit.ErrCount,
							"duration_avg":  unit.DurationAvg,
							"stat_interval": int64(d),
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
				key := fmt.Sprintf("%s-%s", sinfo.Service, sinfo.Resource)
				unit, ok := statUnit[key]
				if ok {
					if sinfo.IsEntry {
						unit.RequestCount++
					}
					if sinfo.IsErr {
						unit.ErrCount++
					}
					unit.DurationAvg += int64(sinfo.Duration)
				} else {
					tstat := &TracingStatistic{
						Toolkit:     sinfo.Toolkit,
						Project:     sinfo.Project,
						Version:     sinfo.Version,
						Service:     sinfo.Service,
						Resource:    sinfo.Resource,
						DurationAvg: int64(sinfo.Duration),
					}
					if sinfo.IsEntry {
						tstat.RequestCount = 1
					}
					if sinfo.IsErr {
						tstat.ErrCount = 1
					}
					statUnit[key] = tstat
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

	log.Error(ErrSendSpanInfoFailed.Error())
}
