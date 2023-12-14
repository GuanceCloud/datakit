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

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hashcode"
)

// FilterFunc is func type for data filter.
// Return the DatakitTraces that need to propagate to next action and
// return ture if one want to skip all FilterFunc afterwards, false otherwise.
type FilterFunc func(log *logger.Logger, dktrace DatakitTrace) (DatakitTrace, bool)

// NoneFilter always return current trace.
func NoneFilter(log *logger.Logger, dktrace DatakitTrace) (DatakitTrace, bool) {
	return dktrace, false
}

type CloseResource struct {
	sync.Mutex
	IgnoreResources map[string][]*regexp.Regexp
}

func (cres *CloseResource) Close(log *logger.Logger, dktrace DatakitTrace) (DatakitTrace, bool) {
	if len(cres.IgnoreResources) == 0 {
		return dktrace, false
	}
	if len(dktrace) == 0 {
		return nil, true
	}

	for i := range dktrace {
		for service, resList := range cres.IgnoreResources {
			if service == "*" || service == dktrace[i].GetTag(TagService) {
				for j := range resList {
					if resList[j].MatchString(dktrace[i].GetFiledToString(FieldResource)) {
						log.Debugf("close trace tid: %s from  resource: %s ",
							dktrace[i].GetFiledToString(FieldTraceID), dktrace[i].GetFiledToString(FieldResource))

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
				if rexp, err := regexp.Compile(resList[i]); err == nil {
					ignResRegs[service] = append(ignResRegs[service], rexp)
				}
			}
		}
		cres.IgnoreResources = ignResRegs
	}
}

func OmitHTTPStatusCodeFilterWrapper(statusCodeList []string) FilterFunc {
	if len(statusCodeList) == 0 {
		return NoneFilter
	} else {
		return func(log *logger.Logger, dktrace DatakitTrace) (DatakitTrace, bool) {
			for i := range dktrace {
				if dktrace[i].GetTag(TagSourceType) != SpanSourceWeb {
					continue
				}
				for j := range statusCodeList {
					if statusCode := dktrace[i].Get(TagHttpStatusCode); statusCode == statusCodeList[j] {
						log.Debugf("omit trace with status code: %s", statusCode)

						return nil, true
					}
				}
			}

			return dktrace, false
		}
	}
}

func PenetrateErrorTracing(log *logger.Logger, dktrace DatakitTrace) (DatakitTrace, bool) {
	if len(dktrace) == 0 {
		return nil, true
	}

	for i := range dktrace {
		switch dktrace[i].GetTag(TagSpanStatus) {
		case StatusErr, StatusCritical:
			log.Debugf("penetrate error trace tid: %s service: %s resource: %s",
				dktrace[i].GetTag(TagService), dktrace[i].GetFiledToString(FieldResource), dktrace[i].Get(TagSourceType))

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

func (kprres *KeepRareResource) Keep(log *logger.Logger, dktrace DatakitTrace) (DatakitTrace, bool) {
	if len(dktrace) == 0 {
		return nil, true
	}
	if !kprres.Open {
		return dktrace, false
	}

	var skip bool
	for i := range dktrace {
		if dktrace[i].Get(TagSpanType) == SpanTypeEntry {
			sed := fmt.Sprintf("%s%s%s",
				dktrace[i].GetTag(TagService), dktrace[i].GetFiledToString(FieldResource), dktrace[i].Get(TagSourceType))
			if len(sed) == 0 {
				break
			}

			checksum := hashcode.GenStringsHash(sed)
			if v, ok := kprres.presentMap.Load(checksum); !ok || time.Since(v.(time.Time)) >= kprres.Duration {
				log.Debugf("got rare trace from service: %s resource: %s send by %s",
					dktrace[i].GetTag(TagService), dktrace[i].GetFiledToString(FieldResource), dktrace[i].Get(TagSourceType))
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

type Sampler struct {
	Priority           int     `toml:"priority" json:"priority"` // deprecated
	SamplingRateGlobal float64 `toml:"sampling_rate" json:"sampling_rate"`
	threshold          uint64
}

func (smp *Sampler) Sample(log *logger.Logger, dktrace DatakitTrace) (DatakitTrace, bool) {
	if len(dktrace) == 0 {
		return nil, true
	}
	traceID := UnifyToUint64ID(dktrace[0].GetFiledToString(FieldTraceID))
	f := traceID%10000 <= smp.threshold
	if f {
		return dktrace, false
	} else {
		return nil, true
	}
}

func (smp *Sampler) Init() *Sampler {
	smp.threshold = uint64(float64(10000) * smp.SamplingRateGlobal)
	return smp
}
