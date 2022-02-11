package trace

import (
	"regexp"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hashcode"
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

func (ds *Sampler) UpdateArgs(priority int, samplingRateGlobal float64) {
	switch priority {
	case PriorityAuto, PriorityReject, PriorityKeep:
		ds.Priority = priority
	}
	if samplingRateGlobal >= 0 && samplingRateGlobal <= 1 {
		ds.SamplingRateGlobal = samplingRateGlobal
	}
}

type CloseResource struct {
	IgnoreResources map[string][]*regexp.Regexp
}

func (close *CloseResource) Close(dktrace DatakitTrace) (DatakitTrace, bool) {
	if len(close.IgnoreResources) == 0 {
		return dktrace, false
	}

	for i := range dktrace {
		if IsRootSpan(dktrace[i]) {
			for service, resList := range close.IgnoreResources {
				if dktrace[i].Service == service {
					for j := range resList {
						if resList[j].MatchString(dktrace[i].Resource) {
							return nil, true
						}
					}
				}
			}
		}
	}

	return dktrace, false
}

func (close *CloseResource) UpdateIgnResList(ignResList map[string][]string) {
	if len(ignResList) == 0 {
		close.IgnoreResources = nil
	} else {
		ignResRegs := make(map[string][]*regexp.Regexp)
		for service, resList := range ignResList {
			ignResRegs[service] = []*regexp.Regexp{}
			for i := range resList {
				ignResRegs[service] = append(ignResRegs[service], regexp.MustCompile(resList[i]))
			}
		}
		close.IgnoreResources = ignResRegs
	}
}

type KeepRareResource struct {
	Open       bool
	Span       time.Duration
	presentMap map[string]time.Time
}

func (keep *KeepRareResource) Keep(dktrace DatakitTrace) (DatakitTrace, bool) {
	if !keep.Open {
		return dktrace, false
	}
	if keep.presentMap == nil {
		keep.presentMap = make(map[string]time.Time)
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
			if lastCheck, ok = keep.presentMap[checkSum]; !ok || time.Since(lastCheck) >= keep.Span {
				skip = true
			}
			keep.presentMap[checkSum] = time.Now()
		}
	}

	return dktrace, skip
}

func (keep *KeepRareResource) UpdateStatus(open bool, span time.Duration) {
	keep.Open = open
	keep.Span = span
}
