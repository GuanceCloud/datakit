package trace

import "strconv"

var (
	DefSampleHandler = func(traceId uint64, rate, scope int) bool {
		return (traceId % uint64(scope)) < uint64(rate)
	}
	ErrMapper = map[string]int32{
		STATUS_OK:       0,
		STATUS_INFO:     0,
		STATUS_WARN:     2,
		STATUS_ERR:      3,
		STATUS_CRITICAL: 4,
	}
	DefErrCheckHandler = func(status string) bool {
		stat, ok := ErrMapper[status]

		return ok && (stat != 0)
	}
	DefIgnoreTagsHandler = func(source map[string]string, ignoreList []string) bool {
		for _, v := range ignoreList {
			if _, ok := source[v]; ok {
				return true
			}
		}

		return false
	}
)

type TraceSampleConfig struct {
	Rate           int      `toml:"rate"`
	Scope          int      `toml:"scope"`
	IgnoreTagsList []string `toml:ignore_tags_list`
}

// sample if true, not sample if false
func (this *TraceSampleConfig) SampleFilter(status string, tags map[string]string, traceId string) bool {
	if this != nil {
		if !DefErrCheckHandler(status) && !DefIgnoreTagsHandler(tags, this.IgnoreTagsList) {
			traceId, err := strconv.ParseInt(traceId, 10, 64)
			if err != nil {
				log.Error(err)

				return true
			}

			return DefSampleHandler(uint64(traceId), this.Rate, this.Scope)
		}
	}

	return true
}
