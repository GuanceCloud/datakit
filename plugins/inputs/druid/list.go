package druid

type MetricType uint8

const (
	Normal MetricType = iota
	Count
	ConvertRange
)

var metricsTemplate = map[string]MetricType{
	"query/time":       Normal,
	"query/bytes":      Count,
	"query/node/bytes": Count,

	"query/success/Count":     Count,
	"query/interrupted/Count": Count,
	"query/failed/Count":      Count,

	"query/node/time":          Normal,
	"query/node/ttfb":          Normal,
	"query/intervalChunk/time": Normal,

	"query/segment/time":         Normal,
	"query/wait/time":            Normal,
	"segment/scan/pending":       Normal,
	"query/segmentAndCache/time": Normal,
	"query/cpu/time":             Normal,

	"query/cache/delta/numEntries":   Count,
	"query/cache/delta/sizeBytes":    Count,
	"query/cache/delta/hits":         Count,
	"query/cache/delta/misses":       Count,
	"query/cache/delta/evictions":    Count,
	"query/cache/delta/hitRate":      ConvertRange,
	"query/cache/delta/averageBytes": Count,
	"query/cache/delta/timeouts":     Count,
	"query/cache/delta/errors":       Count,

	"query/cache/total/numEntries":   Normal,
	"query/cache/total/sizeBytes":    Normal,
	"query/cache/total/hits":         Normal,
	"query/cache/total/misses":       Normal,
	"query/cache/total/evictions":    Normal,
	"query/cache/total/hitRate":      ConvertRange,
	"query/cache/total/averageBytes": Normal,
	"query/cache/total/timeouts":     Normal,
	"query/cache/total/errors":       Normal,

	"ingest/events/thrownAway":    Count,
	"ingest/events/unparseable":   Count,
	"ingest/events/processed":     Count,
	"ingest/rows/output":          Count,
	"ingest/persist/Count":        Count,
	"ingest/persist/time":         Normal,
	"ingest/persist/cpu":          Normal,
	"ingest/persist/backPressure": Normal,
	"ingest/persist/failed":       Count,
	"ingest/handoff/failed":       Count,
	"ingest/merge/time":           Normal,
	"ingest/merge/cpu":            Normal,

	"task/run/time":       Normal,
	"segment/added/bytes": Count,
	"segment/moved/bytes": Count,
	"segment/nuked/bytes": Count,

	"segment/assigned/Count":        Count,
	"segment/moved/Count":           Count,
	"segment/dropped/Count":         Count,
	"segment/deleted/Count":         Count,
	"segment/unneeded/Count":        Count,
	"segment/cost/raw":              Count,
	"segment/cost/normalization":    Count,
	"segment/cost/normalized":       Count,
	"segment/loadQueue/size":        Normal,
	"segment/loadQueue/failed":      Normal,
	"segment/loadQueue/Count":       Normal,
	"segment/dropQueue/Count":       Normal,
	"segment/size":                  Normal,
	"segment/overShadowed/Count":    Normal,
	"segment/underReplicated/Count": Normal,
	"segment/Count":                 Normal,
	"segment/unavailable/Count":     Normal,

	"segment/max":         Normal,
	"segment/used":        Normal,
	"segment/usedPercent": ConvertRange,

	"jvm/pool/committed":      Normal,
	"jvm/pool/init":           Normal,
	"jvm/pool/max":            Normal,
	"jvm/pool/used":           Normal,
	"jvm/bufferpool/Count":    Normal,
	"jvm/bufferpool/used":     Normal,
	"jvm/bufferpool/capacity": Normal,
	"jvm/mem/init":            Normal,
	"jvm/mem/max":             Normal,
	"jvm/mem/used":            Normal,
	"jvm/mem/committed":       Normal,
	"jvm/gc/Count":            Count,
	"jvm/gc/time":             Normal,

	"ingest/events/buffered": Normal,

	"sys/swap/free":        Normal,
	"sys/swap/max":         Normal,
	"sys/swap/pageIn":      Normal,
	"sys/swap/pageOut":     Normal,
	"sys/disk/write/Count": Count,
	"sys/disk/read/Count":  Count,
	"sys/disk/write/size":  Count,
	"sys/disk/read/size":   Count,
	"sys/net/write/size":   Count,
	"sys/net/read/size":    Count,
	"sys/fs/used":          Normal,
	"sys/fs/max":           Normal,
	"sys/mem/used":         Normal,
	"sys/mem/max":          Normal,
	"sys/storage/used":     Normal,
	"sys/cpu":              Normal,

	"coordinator-segment/Count": Normal,
	"historical-segment/Count":  Normal,

	"jetty/numopenconnections": Normal,
}
