package trace

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	tracing_stat_name = "tracingstat"
)

type TracingInfo struct {
	Toolkit      string
	Project      string
	Version      string
	Service      string
	Resource     string
	RequestCount int
	ErrCount     int
	DurationAvg  int64
	key          string
	reCalc       bool
}

var ErrSendSpanInfoFailed = errors.New("send span information failed")

var (
	once            = sync.Once{}
	statUnit        map[string]*TracingInfo
	tracingInfoChan chan *TracingInfo
	calcInterval                  = 30 * time.Second
	sendTimeout     time.Duration = time.Second
	retry           int           = 3
	isWorkerReady   bool          = false
)

func StartTracingStatistic() {
	once.Do(func() {
		statUnit = make(map[string]*TracingInfo)
		tracingInfoChan = make(chan *TracingInfo, 100)
		startTracingStatWorker(calcInterval)
		isWorkerReady = true
	})
}

func startTracingStatWorker(interval time.Duration) {
	log.Info("tracing statistic worker started")

	go func() {
		tick := time.NewTicker(interval)
		for range tick.C {
			sendTracingInfo(&TracingInfo{
				key:    "recalc",
				reCalc: true,
			})
		}
	}()
	go func() {
		for tinfo := range tracingInfoChan {
			if tinfo.reCalc {
				if len(statUnit) == 0 {
					continue
				}

				pts := makeTracingInfoPoint(statUnit)
				if len(pts) == 0 {
					log.Warn("empty tracing stat unit")
				} else {
					if err := io.Feed(tracing_stat_name, datakit.Metric, pts, nil); err != nil {
						log.Error(err.Error())
					}
				}
			} else {
				if tunit, ok := statUnit[tinfo.key]; !ok {
					statUnit[tinfo.key] = tinfo
				} else {
					tunit.RequestCount += tinfo.RequestCount
					tunit.ErrCount += tinfo.ErrCount
					tunit.DurationAvg += tinfo.DurationAvg
				}
			}
		}
	}()
}

func CalcTracingInfo(dktrace DatakitTrace) {
	if !isWorkerReady || len(dktrace) == 0 {
		return
	}

	tracingStatUnit := make(map[string]*TracingInfo)
	for _, dkspan := range dktrace {
		var (
			key   = fmt.Sprintf("%s-%s", dkspan.Service, dkspan.Resource)
			tinfo *TracingInfo
			ok    bool
		)
		if tinfo, ok = tracingStatUnit[key]; !ok {
			tinfo = &TracingInfo{
				Toolkit:  dkspan.Source,
				Project:  dkspan.Project,
				Version:  dkspan.Version,
				Service:  dkspan.Service,
				Resource: dkspan.Resource,
				key:      key,
			}
			tracingStatUnit[key] = tinfo
		}
		if dkspan.SpanType == SPAN_TYPE_ENTRY {
			tinfo.RequestCount++
			if dkspan.Status == STATUS_ERR {
				tinfo.ErrCount++
			}
		}
		tinfo.DurationAvg += dkspan.Duration
	}

	for _, info := range tracingStatUnit {
		sendTracingInfo(info)
	}
}

func sendTracingInfo(tinfo *TracingInfo) {
	if tinfo == nil || tinfo.key == "" {
		return
	}

	timeout := time.NewTimer(sendTimeout)
	for retry > 0 {
		select {
		case <-timeout.C:
			retry--
			timeout.Reset(sendTimeout)
		case tracingInfoChan <- tinfo:
			return
		}
	}
}

func makeTracingInfoPoint(tinfos map[string]*TracingInfo) []*io.Point {
	var pts []*io.Point
	for _, tinfo := range tinfos {
		var (
			tags   = make(map[string]string)
			fields = make(map[string]interface{})
		)
		tags["toolkit"] = tinfo.Toolkit
		tags["project"] = tinfo.Project
		tags["version"] = tinfo.Version
		tags["service"] = tinfo.Service
		tags["resource"] = tinfo.Resource

		fields["request_count"] = tinfo.RequestCount
		fields["err_count"] = -tinfo.ErrCount
		fields["duration_avg"] = tinfo.DurationAvg / int64(tinfo.RequestCount)

		if pt, err := io.MakePoint(tracing_stat_name, tags, fields, time.Now()); err != nil {
			log.Error(err.Error())
		} else {
			pts = append(pts, pt)
		}
	}

	return pts
}
