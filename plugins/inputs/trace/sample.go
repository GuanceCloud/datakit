package trace

var (
	DefSampleHandler = func(traceId uint64, rate, scope int) bool {
		return (traceId % uint64(scope)) < uint64(rate)
	}
	ErrMapper = map[string]int32{
		// STATUS_INFO:     1,
		STATUS_WARN:     2,
		STATUS_ERR:      3,
		STATUS_CRITICAL: 4,
	}
	DefErrCheckHandler = func(err int32) bool {
		return err != 0
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
