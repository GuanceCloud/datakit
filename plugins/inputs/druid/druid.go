package druid

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

// druid config
//
// druid.monitoring.emissionPeriod=PT10s
// druid.monitoring.monitors=["com.metamx.metrics.JvmMonitor"]
// druid.emitter=none
// druid.emitter.http.flushMillis=10000
// druid.emitter.http.recipientBaseUrl=http://ADDR_TO_THIS_SERVICE:8424

func init() {
	inputs.Add("druid", func() inputs.Input {
		return &Druid{}
	})
}

const configSample = `
# [druid]
#       path = "/druid"
#       measurement = "druid"
`

const (
	Normal MetricType = iota
	Count
	ConvertRange
)

var l *zap.SugaredLogger

type MetricType uint8

type Druid struct {
	Config struct {
		Path        string `toml:"path"`
		Measurement string `toml:"measurement"`
	} `toml:"druid"`
}

func (d *Druid) Run() {
	l = logger.SLogger("druid")
	io.RegisterRoute(d.Config.Path, d.handle)
}

func (d *Druid) handle(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.Errorf("failed of read body, err: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	fields := extract(body)
	if fields == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pt, err := influxdb.NewPoint(d.Config.Measurement, nil, fields, time.Now())
	if err != nil {
		l.Errorf("build point err, %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := io.Feed([]byte(pt.PrecisionString("ns")), io.Metric); err != nil {
		l.Errorf("failed of io send, err: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (d *Druid) SampleConfig() string {
	return configSample
}

func (d *Druid) Catalog() string {
	return "druid"
}

type druidMetric []struct {
	Service string  `json:"service"`
	Metric  string  `json:"metric"`
	Value   float64 `json:"value"`
	// discard others..
}

func extract(body []byte) map[string]interface{} {

	var metrics druidMetric
	if err := json.Unmarshal(body, &metrics); err != nil {
		l.Errorf("failed of paras data, err: %s", err.Error())
		return nil
	}

	var fields = make(map[string]interface{}, len(metricsTemplate))

	for _, metric := range metrics {

		metricType, ok := metricsTemplate[metric.Metric]
		if !ok {
			continue
		}

		if metric.Service == "druid/peon" {
			// Skipping all metrics from peon. These are task specific and need some
			// thinking before sending to DataDog.
			continue
		}

		metricKey := strings.Replace(metric.Service+"."+metric.Metric, "/", ".", -1)

		switch metricType {
		case Normal:
			fields[metricKey] = metric.Value
		case Count:
			fields[metricKey] = int64(metric.Value)
		case ConvertRange:
			fields[metricKey] = metric.Value * 100
		default:
			l.Info("Unknown metric type ", metricType)
		}
	}

	return fields
}

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
