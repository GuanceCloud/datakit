// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package trace

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hashcode"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

// NoneFilter always return current trace.
func NoneFilter(dktrace DatakitTrace) (DatakitTrace, bool) {
	return dktrace, false
}

func OmitStatusCodeFilterWrapper(statusCodeList []string) FilterFunc {
	if len(statusCodeList) == 0 {
		return NoneFilter
	} else {
		return func(dktrace DatakitTrace) (DatakitTrace, bool) {
			for i := range dktrace {
				for j := range statusCodeList {
					if dktrace[i].HTTPStatusCode == statusCodeList[j] {
						log.Debugf("omit trace with status code: %s", dktrace[i].HTTPStatusCode)

						return nil, true
					}
				}
			}

			return dktrace, false
		}
	}
}

func PenetrateErrorTracing(dktrace DatakitTrace) (DatakitTrace, bool) {
	if len(dktrace) == 0 {
		return dktrace, true
	}

	for i := range dktrace {
		switch dktrace[i].Status {
		case STATUS_ERR, STATUS_CRITICAL:
			log.Debugf("penetrate tracing %s:%s with status %s", dktrace[i].Service, dktrace[i].Resource, dktrace[i].Status)

			return dktrace, true
		}
	}

	return dktrace, false
}

type CloseResource struct {
	sync.Mutex
	IgnoreResources map[string][]*regexp.Regexp
}

func (cres *CloseResource) Close(dktrace DatakitTrace) (DatakitTrace, bool) {
	if len(dktrace) == 0 {
		return dktrace, true
	}
	if len(cres.IgnoreResources) == 0 {
		return dktrace, false
	}

	for i := range dktrace {
		if dktrace[i].SpanType == SPAN_TYPE_ENTRY {
			for service, resList := range cres.IgnoreResources {
				if service == "*" || service == dktrace[i].Service {
					for j := range resList {
						if resList[j].MatchString(dktrace[i].Resource) {
							log.Debugf("close trace from service: %s resource: %s send by source: %s", dktrace[i].Service, dktrace[i].Resource, dktrace[i].Source)

							return nil, true
						}
					}
				}
			}
		}
	}

	return dktrace, false
}

func (cres *CloseResource) UpdateIgnResList(ignResList map[string][]string) {
	cres.Lock()
	defer cres.Unlock()

	if len(ignResList) == 0 {
		cres.IgnoreResources = nil
	} else {
		ignResRegs := make(map[string][]*regexp.Regexp)
		for service, resList := range ignResList {
			ignResRegs[service] = []*regexp.Regexp{}
			for i := range resList {
				ignResRegs[service] = append(ignResRegs[service], regexp.MustCompile(resList[i]))
			}
		}
		cres.IgnoreResources = ignResRegs
	}
}

type KeepRareResource struct {
	Open       bool
	Duration   time.Duration
	presentMap sync.Map
}

func (kprres *KeepRareResource) Keep(dktrace DatakitTrace) (DatakitTrace, bool) {
	if len(dktrace) == 0 {
		return dktrace, true
	}
	if !kprres.Open {
		return dktrace, false
	}

	var skip bool
	for i := range dktrace {
		if dktrace[i].SpanType == SPAN_TYPE_ENTRY {
			sed := fmt.Sprintf("%s%s%s", dktrace[i].Service, dktrace[i].Resource, dktrace[i].Source)
			if len(sed) == 0 {
				break
			}

			checksum := hashcode.GenStringsHash(sed)
			if v, ok := kprres.presentMap.Load(checksum); !ok || time.Since(v.(time.Time)) >= kprres.Duration {
				log.Debugf("got rare trace from service: %s resource: %s send by %s", dktrace[i].Service, dktrace[i].Resource, dktrace[i].Source)
				skip = true
			}
			kprres.presentMap.Store(checksum, time.Now())
			break
		}
	}

	return dktrace, skip
}

func (kprres *KeepRareResource) UpdateStatus(open bool, span time.Duration) {
	kprres.Open = open
	kprres.Duration = span
	if !open {
		kprres.presentMap = sync.Map{}
	}
}

// tracing data storing priority.
const (
	// reject trace before send to dataway.
	PriorityReject = -1
	// auto calculate with sampling rate.
	PriorityAuto = 0
	// always send to dataway and do not consider sampling and filters.
	PriorityKeep = 1
)

type Sampler struct {
	Priority           int     `toml:"priority" json:"priority"`
	SamplingRateGlobal float64 `toml:"sampling_rate" json:"sampling_rate"`
	ratio              int
	once               sync.Once
}

func (smp *Sampler) Sample(dktrace DatakitTrace) (DatakitTrace, bool) {
	smp.once.Do(func() {
		smp.ratio = int(smp.SamplingRateGlobal * 100)
	})

	if len(dktrace) == 0 {
		return dktrace, true
	}

	switch smp.Priority {
	case PriorityAuto:
		if smp.SamplingRateGlobal >= 1 {
			return dktrace, false
		}
		if int(UnifyToInt64ID(dktrace[0].TraceID)%100) < smp.ratio {
			return dktrace, false
		} else {
			log.Debugf("drop service: %s resource: %s trace_id: %s span_id: %s according to sampling ratio: %d%%",
				dktrace[0].Service, dktrace[0].Resource, dktrace[0].TraceID, dktrace[0].SpanID, smp.ratio)

			return nil, true
		}
	case PriorityReject:
		return nil, true
	case PriorityKeep:
		return dktrace, true
	default:
		log.Debug("unrecognized trace proority")
	}

	return dktrace, false
}

func (smp *Sampler) UpdateArgs(priority int, samplingRateGlobal float64) {
	switch priority {
	case PriorityAuto, PriorityReject, PriorityKeep:
		smp.Priority = priority
	}
	if samplingRateGlobal >= 0 && samplingRateGlobal <= 1 {
		smp.SamplingRateGlobal = samplingRateGlobal
		smp.ratio = int(smp.SamplingRateGlobal * 100)
	}
}

func PiplineFilterWrapper(source string, piplines map[string]string) FilterFunc {
	if len(piplines) == 0 {
		return NoneFilter
	}

	var (
		pips = make(map[string]*pipeline.Pipeline)
		err  error
	)
	for k := range piplines {
		if pips[k], err = pipeline.NewPipeline(piplines[k]); err != nil {
			log.Debugf("create pipeline %s failed: %s", k, err)
			continue
		}
	}
	if len(pips) == 0 {
		return NoneFilter
	}

	return func(dktrace DatakitTrace) (DatakitTrace, bool) {
		if len(dktrace) == 0 {
			return dktrace, true
		}

		for s, p := range pips {
			for i := range dktrace {
				if dktrace[i].Service == s {
					if rslt, err := p.Run(dktrace[i].Content, source); err != nil {
						log.Debugf("run pipeline %s.p failed: %s", s, err.Error())
					} else {
						if len(rslt.Output.Tags) > 0 {
							if dktrace[i].Tags == nil {
								dktrace[i].Tags = make(map[string]string)
							}
							for k, v := range rslt.Output.Tags {
								dktrace[i].Tags[k] = v
							}
						}
						if len(rslt.Output.Fields) > 0 {
							if dktrace[i].Metrics == nil {
								dktrace[i].Metrics = make(map[string]interface{})
							}
							for k, v := range rslt.Output.Fields {
								dktrace[i].Metrics[k] = v
							}
						}
					}
				}
			}
		}

		return dktrace, false
	}
}
