package io

import (
	"errors"
	"fmt"
	"time"
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
	TraceId    int64
	Service    string
	Resource   string
	VisitCount int
	ErrCount   int
	Duration   time.Duration
	StatSpan   time.Duration
}

var ErrSendSpanInfoFailed = errors.New("send span information failed")

var (
	statUnit                   = make(map[string]*TracingStatistic)
	spanInfoChan               = make(chan *SpanInfo)
	sendTimeout  time.Duration = 10 * time.Second
	retry        int           = 3
)

func StartTracingStatWorker(d time.Duration) {
	go func() {
		tick := time.NewTicker(d)
		for range tick.C {
			spanInfoChan <- &SpanInfo{reStat: true}
		}
	}()
	go func() {
		for sinfo := range spanInfoChan {
			if sinfo.reStat {
				// TODO: calc statistics, make point then send to io
				for _, unit := range statUnit {
					unit.Duration /= time.Duration(unit.VisitCount)
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
					unit.Duration += sinfo.Duration
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
						Duration: sinfo.Duration,
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
