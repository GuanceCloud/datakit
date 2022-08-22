// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package trace

import (
	"fmt"
	"math"
	"regexp"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hashcode"
)

// NoneFilter always return current trace.
func NoneFilter(dktrace DatakitTrace) (DatakitTrace, bool) {
	return dktrace, false
}

type CloseResource struct {
	sync.Mutex
	IgnoreResources map[string][]*regexp.Regexp
}

func (cres *CloseResource) Close(dktrace DatakitTrace) (DatakitTrace, bool) {
	if len(cres.IgnoreResources) == 0 {
		return dktrace, false
	}
	if len(dktrace) == 0 {
		return nil, true
	}

	for i := range dktrace {
		for service, resList := range cres.IgnoreResources {
			if service == "*" || service == dktrace[i].Service {
				for j := range resList {
					if resList[j].MatchString(dktrace[i].Resource) {
						log.Debugf("close trace tid: %s from service: %s resource: %s send by source: %s",
							dktrace[i].TraceID, dktrace[i].Service, dktrace[i].Resource, dktrace[i].Source)

						return nil, true
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

func RespectUserRule(dktrace DatakitTrace) (DatakitTrace, bool) {
	if len(dktrace) == 0 {
		return nil, true
	}

	for i := range dktrace {
		if p, ok := dktrace[i].Metrics[FIELD_PRIORITY]; ok {
			var priority int
			if priority, ok = p.(int); !ok {
				log.Debugf("wrong type for priority %v", p)
				continue
			}
			switch priority {
			case PRIORITY_USER_REJECT, PRIORITY_RULE_SAMPLER_REJECT:
				log.Debugf("drop tid: %s service: %s resource: %s according to %s.",
					dktrace[i].TraceID, dktrace[i].Service, dktrace[i].Resource, priorityRules[priority])

				return nil, true
			case PRIORITY_USER_KEEP, PRIORITY_RULE_SAMPLER_KEEP:
				log.Debugf("send tid: %s service: %s resource: %s according to %s.",
					dktrace[i].TraceID, dktrace[i].Service, dktrace[i].Resource, priorityRules[priority])

				return dktrace, true
			case PRIORITY_AUTO_REJECT, PRIORITY_AUTO_KEEP:
				log.Debugf("keep tid: %s service: %s resource: %s according to %s.",
					dktrace[i].TraceID, dktrace[i].Service, dktrace[i].Resource, priorityRules[priority])

				return dktrace, false
			default:
				log.Infof("[note:] no proper priority number(%s) found, this may be a potential bug, tid: %s service: %s resource: %s",
					priority, dktrace[i].TraceID, dktrace[i].Service, dktrace[i].Resource)

				return dktrace, false
			}
		}
	}

	log.Infof("[note] no priority span found in trace, this may be a potential bug, tid: %s service: %s resource: %s",
		dktrace[0].TraceID, dktrace[0].Service, dktrace[0].Resource)

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
		return nil, true
	}

	for i := range dktrace {
		switch dktrace[i].Status {
		case STATUS_ERR, STATUS_CRITICAL:
			log.Debugf("penetrate error trace tid: %s service: %s resource: %s", dktrace[i].TraceID, dktrace[i].Service, dktrace[i].Resource)

			return dktrace, true
		}
	}

	return dktrace, false
}

type KeepRareResource struct {
	Open       bool
	Duration   time.Duration
	presentMap sync.Map
}

func (kprres *KeepRareResource) Keep(dktrace DatakitTrace) (DatakitTrace, bool) {
	if len(dktrace) == 0 {
		return nil, true
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

// constants used for the Knuth hashing, same as dd-agent.
const knuthFactor = uint64(1111111111111111111)

func multiplicativeHashFunc(n uint64, rate float64) bool {
	if rate < 1 {
		return n*knuthFactor < uint64(math.MaxUint64*rate)
	}

	return true
}

type Sampler struct {
	Priority           int     `toml:"priority" json:"priority"` // deprecated
	SamplingRateGlobal float64 `toml:"sampling_rate" json:"sampling_rate"`
}

func (smp *Sampler) Sample(dktrace DatakitTrace) (DatakitTrace, bool) {
	if len(dktrace) == 0 {
		return nil, true
	}

	for i := range dktrace {
		if p, ok := dktrace[i].Metrics[FIELD_PRIORITY]; ok {
			var priority int
			if priority, ok = p.(int); !ok {
				log.Debugf("wrong type for priority %v", p)
				continue
			}
			switch priority {
			case PRIORITY_AUTO_KEEP:
				if multiplicativeHashFunc(UnifyToUint64ID(dktrace[i].TraceID), smp.SamplingRateGlobal) {
					log.Debugf("keep tid: %s service: %s resource: %s according to PRIORITY_AUTO_KEEP and sampling ratio: %d%%",
						dktrace[i].TraceID, dktrace[i].Service, dktrace[i].Resource, int(smp.SamplingRateGlobal*100))

					return dktrace, false
				} else {
					log.Debugf("drop tid: %s service: %s resource: %s according to PRIORITY_AUTO_KEEP and sampling ratio: %d%%",
						dktrace[i].TraceID, dktrace[i].Service, dktrace[i].Resource, int(smp.SamplingRateGlobal*100))

					return nil, true
				}
			case PRIORITY_AUTO_REJECT:
				log.Debugf("drop tid: %s service: %s resource: %s according to PRIORITY_AUTO_REJECT.",
					dktrace[i].TraceID, dktrace[i].Service, dktrace[i].Resource)

				return nil, true
			default:
				log.Infof("[note] no proper priority(%s) rules selected, this may be a potential bug, tid: %s service: %s resource: %s",
					priority, dktrace[i].TraceID, dktrace[i].Service, dktrace[i].Resource)

				return dktrace, false
			}
		}
	}

	log.Infof("[note] no priority span found in trace, this may be a potential bug, tid: %s service: %s resource: %s",
		dktrace[0].TraceID, dktrace[0].Service, dktrace[0].Resource)

	return dktrace, false
}
