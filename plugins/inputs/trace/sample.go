package trace

import "fmt"

var (
	DefSampleFunc = sample
	DefErrMapper  = map[string]int32{
		STATUS_OK:       0,
		STATUS_INFO:     1,
		STATUS_WARN:     2,
		STATUS_ERR:      3,
		STATUS_CRITICAL: 4,
	}
)

type TraceSampleConfig struct {
	Target         map[string]string `toml:"target,omitempty"` // deprecated in new issue
	Rate           int               `toml:"rate"`
	Scope          int               `toml:"scope"`
	IgnoreTagsList []string          `toml:"ignore_tags_list"`
}

// TraceSampleMatcher return sample config assgined by ddtrace config.
func SampleConfMatcher(confs []*TraceSampleConfig, tags map[string]string) *TraceSampleConfig {
	var conf *TraceSampleConfig
	for _, conf = range confs {
		for k, v := range conf.Target {
			if tags[k] == v {
				return conf
			}
			// only match once, since we dont allow multiple target config
			break
		}
	}

	if conf != nil && len(conf.Target) == 0 {
		// return the last sample config as default sample config
		return conf
	} else {
		return nil
	}
}

func SampleIgnoreErrStatus(status string) bool {
	stat, ok := DefErrMapper[status]

	return ok && stat > DefErrMapper[STATUS_INFO]
}

func SampleIgnoreKeys(tags map[string]string, ignore []string) bool {
	if len(tags) != 0 {
		for _, v := range ignore {
			if _, ok := tags[v]; ok {
				return true
			}
		}
	}

	return false
}

func SampleIgnoreTags(tags map[string]string, ignore map[string]string) bool {
	if len(tags) != 0 {
		for k, v := range ignore {
			if tags[k] == v {
				return true
			}
		}
	}

	return false
}

func sample(traceId uint64, rate, scope int) bool {
	return (traceId % uint64(scope)) < uint64(rate)
}

func MergeTags(data ...map[string]string) map[string]string {
	merged := map[string]string{}
	for _, tags := range data {
		for k, v := range tags {
			merged[k] = v
		}
	}

	return merged
}

func GetTraceId(high, low int64) int64 {
	temp := low
	for temp != 0 {
		high *= 10
		temp /= 10
	}

	return high + low
}

func GetStringTraceId(high, low int64) string {
	return fmt.Sprintf("%d%d", high, low)
}
