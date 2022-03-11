package trace

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hashcode"
)

// tracing data keep priority.
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

	for i := range dktrace {
		if IsRootSpan(dktrace[i]) {
			switch smp.Priority {
			case PriorityAuto:
				if smp.SamplingRateGlobal >= 1 {
					return dktrace, false
				}
				if int(UnifyToInt64ID(dktrace[i].TraceID)%100) < smp.ratio {
					return dktrace, false
				} else {
					log.Debugf("drop service: %s resource: %s trace_id: %s span_id: %s according to sampling ratio: %d%%",
						dktrace[i].Service, dktrace[i].Resource, dktrace[i].TraceID, dktrace[i].SpanID, smp.ratio)

					return nil, true
				}
			case PriorityReject:
				return nil, true
			case PriorityKeep:
				return dktrace, true
			default:
				log.Debug("unrecognized trace proority")
			}
		}
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

type CloseResource struct {
	sync.Mutex
	IgnoreResources map[string][]*regexp.Regexp
}

func (cres *CloseResource) Close(dktrace DatakitTrace) (DatakitTrace, bool) {
	if len(cres.IgnoreResources) == 0 {
		return dktrace, false
	}

	for i := range dktrace {
		if IsRootSpan(dktrace[i]) {
			for service, resList := range cres.IgnoreResources {
				if dktrace[i].Service == service {
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
	once       sync.Once
	presentMap sync.Map
}

func (kprres *KeepRareResource) Keep(dktrace DatakitTrace) (DatakitTrace, bool) {
	if !kprres.Open {
		return dktrace, false
	}

	var skip bool
	for i := range dktrace {
		if IsRootSpan(dktrace[i]) {
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
	if !kprres.Open {
		kprres.presentMap = sync.Map{}
	}
}
