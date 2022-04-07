package trace

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hashcode"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/worker"
)

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

var _ worker.Task = &DkTracePiplineTask{}

type DkTracePiplineTask struct {
	pipFileName string
	DatakitTrace
}

func (pt *DkTracePiplineTask) GetSource() string {
	if len(pt.DatakitTrace) != 0 {
		return pt.DatakitTrace[0].Source
	} else {
		return ""
	}
}

func (pt *DkTracePiplineTask) GetScriptName() string {
	return pt.pipFileName
}

func (pt *DkTracePiplineTask) GetMaxMessageLen() int {
	return 0
}

func (pt *DkTracePiplineTask) ContentType() string {
	return worker.ContentString
}

func (pt *DkTracePiplineTask) ContentEncode() string {
	return ""
}

func (pt *DkTracePiplineTask) GetContent() interface{} {
	var content []string
	for i := range pt.DatakitTrace {
		content = append(content, pt.DatakitTrace[i].Content)
	}

	return content
}

func (pt *DkTracePiplineTask) Callback(rslt []*pipeline.Result) error {
	if len(pt.DatakitTrace) != len(rslt) {
		return fmt.Errorf("result count is less than input")
	}

	for i := range rslt {
		if len(rslt[i].Err) != 0 {
			log.Debugf("encounter error when traversing pipline results", rslt[i].Err)
			continue
		}

		for k, v := range rslt[i].Output.Tags {
			pt.DatakitTrace[i].Tags[k] = v
		}
		for k, v := range rslt[i].Output.Fields {
			pt.DatakitTrace[i].Metrics[k] = v
		}
	}

	return nil
}

func PiplineFilterWrapper(piplines map[string]string) FilterFunc {
	return func(dktrace DatakitTrace) (DatakitTrace, bool) {
		if len(dktrace) == 0 {
			return dktrace, true
		}

		for k, v := range piplines {
			if k == dktrace[0].Service {
				task := &DkTracePiplineTask{
					pipFileName:  v,
					DatakitTrace: dktrace,
				}
				if err := worker.FeedPipelineTaskBlock(task); err != nil {
					log.Debugf("run pipline error: %s", err)
				}
			}
		}

		return dktrace, false
	}
}
