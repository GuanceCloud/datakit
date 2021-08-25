package trace

var (
	DefSampleFunc SampleHandler = sampleHandlerFunc
	DefErrMapper                = map[string]int32{
		STATUS_OK:       0,
		STATUS_INFO:     0,
		STATUS_WARN:     2,
		STATUS_ERR:      3,
		STATUS_CRITICAL: 4,
	}
)

// will be sampled if true, not if false
type SampleHandler func(traceId uint64, rate, scope int) bool

func sampleHandlerFunc(traceId uint64, rate, scope int) bool {
	return (traceId % uint64(scope)) < uint64(rate)
}

type TraceSampleConfig struct {
	Target         map[string]string `toml:"target"`
	Rate           int               `toml:"rate"`
	Scope          int               `toml:"scope"`
	IgnoreTagsList []string          `toml:"ignore_tags_list"`
}

func TraceSampleMatcher(confs []*TraceSampleConfig, tags map[string]string) *TraceSampleConfig {
	var conf *TraceSampleConfig
	for _, conf = range confs {
		for k, v := range conf.Target {
			if tags[k] == v {
				return conf
			}
			// only match once
			break
		}
	}

	if conf != nil && len(conf.Target) == 0 {
		return conf
	} else {
		return nil
	}
}

func IgnoreErrSampleMW(status string, sampleFunc SampleHandler) SampleHandler {
	return func(traceId uint64, rate, scope int) bool {
		if stat, ok := DefErrMapper[status]; ok && (stat != 0) {
			return true
		} else {
			return sampleFunc(traceId, rate, scope)
		}
	}
}

func IgnoreTagsSampleMW(source map[string]string, ignores []string, sampleFunc SampleHandler) SampleHandler {
	return func(traceId uint64, rate, scope int) bool {
		for _, v := range ignores {
			if _, ok := source[v]; ok {
				return true
			}
		}

		return sampleFunc(traceId, rate, scope)
	}
}

func IgnoreKVPairsSampleMW(source map[string]string, ignores map[string]string, sampleFunc SampleHandler) SampleHandler {
	return func(traceId uint64, rate, scope int) bool {
		if len(source) != 0 {
			for k, v := range ignores {
				if source[k] == v {
					return true
				}
			}
		}

		return sampleFunc(traceId, rate, scope)
	}
}
