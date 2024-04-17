// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package clickhousev1 scrape clickhouse exporter metrics.
package clickhousev1

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
)

const (
	minInterval             = time.Second
	maxInterval             = time.Minute
	inputName               = "clickhousev1"
	source                  = "clickhouse"
	defaultIntervalDuration = time.Second * 30

	// defaultMaxFileSize is the default max response body size, in bytes.
	// This field is used only when metrics are written to file, ipt.e. Output is configured.
	// If the size of response body is over defaultMaxFileSize, metrics will be discarded.
	// 32 MB.
	defaultMaxFileSize int64 = 32 * 1024 * 1024

	clickHouseAsyncMetrics string = "ClickHouseAsyncMetrics"
)

var (
	l = logger.DefaultSLogger(inputName)

	_ inputs.ElectionInput = (*Input)(nil)
)

type urlTags map[string]string

type Input struct {
	Interval         time.Duration `toml:"interval"`
	Timeout          time.Duration `toml:"timeout"`
	ConnectKeepAlive time.Duration `toml:"-"`

	URLs                   []string     `toml:"urls"`
	IgnoreReqErr           bool         `toml:"ignore_req_err"`
	MetricTypes            []string     `toml:"metric_types"`
	MetricNameFilter       []string     `toml:"metric_name_filter"`
	MetricNameFilterIgnore []string     `toml:"metric_name_filter_ignore"`
	MeasurementPrefix      string       `toml:"measurement_prefix"`
	MeasurementName        string       `toml:"measurement_name"`
	Measurements           []iprom.Rule `toml:"measurements"`
	Output                 string       `toml:"output"`
	MaxFileSize            int64        `toml:"max_file_size"`

	TLSOpen    bool   `toml:"tls_open"`
	UDSPath    string `toml:"uds_path"`
	CacertFile string `toml:"tls_ca"`
	CertFile   string `toml:"tls_cert"`
	KeyFile    string `toml:"tls_key"`

	TagsIgnore  []string            `toml:"tags_ignore"`
	TagsRename  *iprom.RenameTags   `toml:"tags_rename"`
	IgnoreTagKV map[string][]string `toml:"ignore_tag_kv_match"`
	HTTPHeaders map[string]string   `toml:"http_headers"`

	Tags               map[string]string `toml:"tags"`
	DisableHostTag     bool              `toml:"disable_host_tag"`
	DisableInstanceTag bool              `toml:"disable_instance_tag"`
	DisableInfoTag     bool              `toml:"disable_info_tag"`

	Auth map[string]string `toml:"auth"`

	semStop    *cliutils.Sem
	feeder     dkio.Feeder
	pm         *iprom.Prom
	mergedTags map[string]urlTags
	tagger     datakit.GlobalTagger

	Election bool `toml:"election"`
	pauseCh  chan bool
	pause    bool

	urls []*url.URL

	l *logger.Logger
}

func (ipt *Input) Run() {
	if err := ipt.setup(); err != nil {
		l.Errorf("setup err: %v", err)
		return
	}

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	for {
		if ipt.pause {
			l.Debug("%s election paused", inputName)
		} else {
			if err := ipt.collect(); err != nil {
				ipt.l.Warn(err)
			}
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Infof("%s input exit", inputName)
			return
		case <-ipt.semStop.Wait():
			l.Infof("%s input return", inputName)
			return
		case ipt.pause = <-ipt.pauseCh:
		}
	}
}

func (ipt *Input) setup() error {
	l = logger.SLogger(inputName)

	l.Infof("%s input started", inputName)
	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)
	// ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")
	// l.Debugf("merged tags: %+#v", ipt.mergedTags)

	ipt.l = l

	for _, u := range ipt.URLs {
		uu, err := url.Parse(u)
		if err != nil {
			return err
		}
		ipt.urls = append(ipt.urls, uu)

		// add extra `instance' tag, the tag take higher priority
		// over global tags.
		if !ipt.DisableInstanceTag {
			if _, ok := ipt.Tags["instance"]; !ok {
				ipt.Tags["instance"] = uu.Host
			}
		}

		if ipt.Election {
			ipt.mergedTags[u] = inputs.MergeTags(ipt.tagger.ElectionTags(), ipt.Tags, u)
		} else {
			ipt.mergedTags[u] = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, u)
		}
	}

	opts := []iprom.PromOption{
		iprom.WithLogger(ipt.l), // WithLogger must in the first
		iprom.WithSource(source),
		iprom.WithTimeout(ipt.Timeout),
		iprom.WithKeepAlive(ipt.ConnectKeepAlive),
		iprom.WithIgnoreReqErr(ipt.IgnoreReqErr),
		iprom.WithMetricTypes(ipt.MetricTypes),
		iprom.WithMetricNameFilter(ipt.MetricNameFilter),
		iprom.WithMetricNameFilterIgnore(ipt.MetricNameFilterIgnore),
		iprom.WithMeasurementPrefix(ipt.MeasurementPrefix),
		iprom.WithMeasurementName(ipt.MeasurementName),
		iprom.WithMeasurements(ipt.Measurements),
		iprom.WithOutput(ipt.Output),
		iprom.WithMaxFileSize(ipt.MaxFileSize),
		iprom.WithTLSOpen(ipt.TLSOpen),
		iprom.WithUDSPath(ipt.UDSPath),
		iprom.WithCacertFile(ipt.CacertFile),
		iprom.WithCertFile(ipt.CertFile),
		iprom.WithKeyFile(ipt.KeyFile),
		iprom.WithTagsIgnore(ipt.TagsIgnore),
		iprom.WithTagsRename(ipt.TagsRename),
		iprom.WithIgnoreTagKV(ipt.IgnoreTagKV),
		iprom.WithHTTPHeaders(ipt.HTTPHeaders),
		iprom.WithDisableInfoTag(ipt.DisableInfoTag),
		iprom.WithAuth(ipt.Auth),
	}

	pm, err := iprom.NewProm(opts...)
	if err != nil {
		// ipt.l.Warnf("clickhouse.NewProm: %s, ignored", err)
		return err
	}
	ipt.pm = pm

	return nil
}

func (ipt *Input) collect() error {
	start := time.Now()
	pts, err := ipt.doCollect()
	if err != nil {
		return err
	}
	if len(pts) < 1 {
		return fmt.Errorf("points got nil from doCollect")
	}

	if err := ipt.feeder.FeedV2(point.Metric, pts,
		dkio.WithCollectCost(time.Since(start)),
		dkio.WithElection(ipt.Election),
		dkio.WithInputName(inputName)); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			dkio.WithLastErrorInput(inputName),
			dkio.WithLastErrorCategory(point.Metric),
		)
		l.Errorf("feed measurement: %s", err)
	}

	return nil
}

func (ipt *Input) doCollect() ([]*point.Point, error) {
	ipt.l.Debugf("collect URLs %v", ipt.URLs)

	// If Output is configured, data is written to local file specified by Output.
	// Data will no more be written to datakit io.
	if ipt.Output != "" {
		err := ipt.WriteMetricText2File()
		if err != nil {
			ipt.l.Errorf("WriteMetricText2File: %s", err.Error())
		}
		return nil, nil
	}

	pts, err := ipt.getPts()
	if err != nil {
		ipt.l.Errorf("getPts: %s", err)
		ipt.feeder.FeedLastError(err.Error(),
			dkio.WithLastErrorInput(inputName),
			dkio.WithLastErrorSource(source),
		)

		// Try testing the connect
		for _, u := range ipt.urls {
			if err := net.RawConnect(u.Hostname(), u.Port(), time.Second*3); err != nil {
				ipt.l.Errorf("failed to connect to %s:%s, %s", u.Hostname(), u.Port(), err)
			}
		}

		return nil, err
	}

	if pts == nil {
		return nil, fmt.Errorf("points got nil from Collect")
	}

	return pts, nil
}

// get points from all URLs.
func (ipt *Input) getPts() ([]*point.Point, error) {
	if ipt.pm == nil {
		return nil, fmt.Errorf("ipt.pm is nil")
	}
	var points []*point.Point
	for _, u := range ipt.URLs {
		uu, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		var pts []*point.Point
		if uu.Scheme != "http" && uu.Scheme != "https" {
			pts, err = ipt.CollectFromFile(u)
		} else {
			pts, err = ipt.CollectFromHTTP(u)
		}
		if err != nil {
			return nil, err
		}

		// append tags to points
		for k, v := range ipt.mergedTags[u] {
			for _, pt := range pts {
				pt.AddTag(k, v)
			}
		}

		points = append(points, pts...)
	}

	return ipt.formatPointSuffixes(points), nil
}

// formatPointSuffixes modify all points who have suffix.
func (ipt *Input) formatPointSuffixes(pts []*point.Point) []*point.Point {
	if len(pts) < 1 {
		return pts
	}
	// All suffix info, store average/total message, every loop is new.
	suffixes := suffixInfos()
	// The points timestamp, will use in average/total point.
	ts := pts[0].Time()

	instance := pts[0].GetTag("instance")

	for j := 0; j < len(pts); j++ {
		// Only some ClickHouseAsyncMetrics metrics set have suffix.
		if pts[j].Name() == clickHouseAsyncMetrics {
			ipt.formatPointSuffix(suffixes, pts[j])
		}
	}

	// Add average/total point.
	for j := 0; j < len(suffixes); j++ {
		if suffixes[j].count <= 0 {
			continue
		}

		// Create tags.
		if suffixes[j].wantAverage {
			suffixes[j].total /= float64(suffixes[j].count)
			suffixes[j].tags[suffixes[j].tagKey] = "average"
		} else {
			suffixes[j].tags[suffixes[j].tagKey] = "total"
		}
		if instance != "" {
			suffixes[j].tags["instance"] = instance
		}

		// Create fields.
		fields := make(map[string]interface{})
		switch suffixes[j].totalType {
		case "F":
			fields[suffixes[j].name] = suffixes[j].total
		case "I":
			value := int64(suffixes[j].total)
			fields[suffixes[j].name] = value
		case "U":
			value := uint64(suffixes[j].total)
			fields[suffixes[j].name] = value
		default:
			continue
		}

		// Add average/total point.
		pt := point.NewPointV2(clickHouseAsyncMetrics,
			append(point.NewTags(suffixes[j].tags), point.NewKVs(fields)...),
			append(point.DefaultMetricOptions(), point.WithTime(ts))...)
		pts = append(pts, pt)
	}
	return pts
}

type willModifyKV struct {
	key      string
	suffixID int
}

// formatPointSuffix modify the point who have suffix.
// A point can have multiple fields, some of which may match the suffix.
// If a field(KV) matches:
// - Sum the V, store in suffixes slice.
// - Delete this field(KV).
// - Add a New field(KV), but the K needs to be cut off the suffix.
// - The suffix will be a new tag's value.
// Average/total will be a new point.
func (ipt *Input) formatPointSuffix(suffixes []suffixInfo, pt *point.Point) {
	// Step 1, traverse to find the field(KV) to modify.
	willModifyKVs := []willModifyKV{}
	// kvs := pt.Fields()
	for _, kv := range pt.Fields() {
		key := kv.Key
		for j := 0; j < len(suffixes); j++ {
			if key == suffixes[j].name || !strings.HasPrefix(key, suffixes[j].name) {
				continue
			}

			// Add a modification amount.
			willModifyKVs = append(willModifyKVs, willModifyKV{key, j})

			// After assertion, the data is converted to flo64 format and accumulated in the suffixes table.
			switch v := kv.Val.(type) {
			case *point.Field_F:
				suffixes[j].count++
				suffixes[j].total += v.F
				suffixes[j].totalType = "F"
			case *point.Field_I:
				suffixes[j].count++
				suffixes[j].total += float64(v.I)
				suffixes[j].totalType = "I"
			case *point.Field_U:
				suffixes[j].count++
				suffixes[j].total += float64(v.U)
				suffixes[j].totalType = "U"
			default:
				break
			}

			// Tags only need add one times.
			if suffixes[j].count > 1 {
				break
			}

			tags := make(map[string]string)
			for _, tag := range pt.Tags() {
				if value, ok := tag.Val.(*point.Field_S); ok {
					tags[tag.Key] = value.S
				}
			}
			suffixes[j].tags = tags
			break
		}
	}

	// Step 2, add tag, add/delete field(KV).
	for _, v := range willModifyKVs {
		// Add tag.
		tagKey := suffixes[v.suffixID].tagKey
		tagVal := strings.TrimPrefix(v.key, suffixes[v.suffixID].name)
		tagVal = strings.TrimPrefix(tagVal, "_") // some suffix like "_coretemp_Core_0".
		pt.MustAddTag(tagKey, tagVal)

		// Add field(KV).
		kv := &point.Field{
			Key: suffixes[v.suffixID].name,
			Val: pt.Fields().Get(v.key).Val,
		}
		pt.MustAddKVs(kv)

		// Delete old field(KV).
		pt.Del(v.key)
	}
}

func (ipt *Input) CollectFromHTTP(u string) ([]*point.Point, error) {
	if ipt.pm == nil {
		return nil, nil
	}
	return ipt.pm.CollectFromHTTPV2(u)
}

func (ipt *Input) CollectFromFile(filepath string) ([]*point.Point, error) {
	if ipt.pm == nil {
		return nil, nil
	}
	return ipt.pm.CollectFromFileV2(filepath)
}

// WriteMetricText2File collects from all URLs and then
// directly writes them to file specified by field Output.
func (ipt *Input) WriteMetricText2File() error {
	// Remove if file already exists.
	if _, err := os.Stat(ipt.Output); err == nil {
		if err := os.Remove(ipt.Output); err != nil {
			return err
		}
	}
	for _, u := range ipt.URLs {
		if err := ipt.pm.WriteMetricText2File(u); err != nil {
			return err
		}
		stat, err := os.Stat(ipt.Output)
		if err != nil {
			return err
		}
		if stat.Size() > ipt.MaxFileSize {
			return fmt.Errorf("file size is too large, max: %d, got: %d", ipt.MaxFileSize, stat.Size())
		}
	}
	return nil
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}
func (*Input) Catalog() string          { return inputName }
func (*Input) SampleConfig() string     { return sampleCfg }
func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }
func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&asyncMetricsMeasurement{},
		&metricsMeasurement{},
		&profileEventsMeasurement{},
		&statusInfoMeasurement{},
	}
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

var maxPauseCh = inputs.ElectionPauseChannelLength

func NewProm() *Input {
	return &Input{
		pauseCh:     make(chan bool, maxPauseCh),
		MaxFileSize: defaultMaxFileSize,
		Interval:    defaultIntervalDuration,
		Timeout:     time.Second * 30,
		Election:    true,
		Tags:        make(map[string]string),

		mergedTags: map[string]urlTags{},

		semStop: cliutils.NewSem(),
		feeder:  dkio.DefaultFeeder(),
		tagger:  datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return NewProm()
	})
}
