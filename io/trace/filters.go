package trace

import (
	"regexp"
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
}

func (smp *Sampler) Sample(dktrace DatakitTrace) (DatakitTrace, bool) {
	for i := range dktrace {
		if IsRootSpan(dktrace[i]) {
			switch dktrace[i].Priority {
			case PriorityAuto:
				if smp.SamplingRateGlobal >= 1 {
					return dktrace, false
				}
				tid := UnifyToInt64ID(dktrace[i].TraceID)
				if tid%100 < int64(smp.SamplingRateGlobal*100) {
					return dktrace, false
				} else {
					log.Debugf("drop service: %s resource %s trace_id: %s span_id: %s by default sampler",
						dktrace[i].Service, dktrace[i].Resource, dktrace[i].TraceID, dktrace[i].SpanID)

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
	}
}

type CloseResource struct {
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
							log.Debugf("close service: %s resource: %s from %s", dktrace[i].Service, dktrace[i].Resource, dktrace[i].Source)

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
	Span       time.Duration
	presentMap map[string]time.Time
}

func (kprres *KeepRareResource) Keep(dktrace DatakitTrace) (DatakitTrace, bool) {
	if !kprres.Open {
		return dktrace, false
	}
	if kprres.presentMap == nil {
		kprres.presentMap = make(map[string]time.Time)
	}

	var skip bool
	for i := range dktrace {
		if IsRootSpan(dktrace[i]) {
			var (
				checkSum = hashcode.GenMapHash(map[string]string{
					"service":  dktrace[i].Service,
					"resource": dktrace[i].Resource,
					"env":      dktrace[i].Env,
				})
				lastCheck time.Time
				ok        bool
			)
			if lastCheck, ok = kprres.presentMap[checkSum]; !ok || time.Since(lastCheck) >= kprres.Span {
				log.Debugf("got rare service: %s resource: %s from %s", dktrace[i].Service, dktrace[i].Resource, dktrace[i].Source)
				skip = true
			}
			kprres.presentMap[checkSum] = time.Now()
		}
	}

	return dktrace, skip
}

func (kprres *KeepRareResource) UpdateStatus(open bool, span time.Duration) {
	kprres.Open = open
	kprres.Span = span
}
