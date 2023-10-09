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

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
)

var _ inputs.ElectionInput = (*Input)(nil)

const (
	inputName               = "clickhousev1"
	catalog                 = "clickhouse"
	defaultIntervalDuration = time.Second * 30

	// defaultMaxFileSize is the default max response body size, in bytes.
	// This field is used only when metrics are written to file, i.e. Output is configured.
	// If the size of response body is over defaultMaxFileSize, metrics will be discarded.
	// 32 MB.
	defaultMaxFileSize int64 = 32 * 1024 * 1024
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	Source           string        `toml:"source"`
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
	AsLogging   *iprom.AsLogging    `toml:"as_logging"`
	IgnoreTagKV map[string][]string `toml:"ignore_tag_kv_match"`
	HTTPHeaders map[string]string   `toml:"http_headers"`

	Tags               map[string]string `toml:"tags"`
	DisableHostTag     bool              `toml:"disable_host_tag"`
	DisableInstanceTag bool              `toml:"disable_instance_tag"`
	DisableInfoTag     bool              `toml:"disable_info_tag"`

	Auth map[string]string `toml:"auth"`

	pm     *iprom.Prom
	Feeder io.Feeder

	Election bool `toml:"election"`
	chPause  chan bool
	pause    bool

	Tagger dkpt.GlobalTagger

	urls []*url.URL

	semStop *cliutils.Sem // start stop signal

	isInitialized bool

	urlTags map[string]urlTags

	l *logger.Logger
}

type urlTags map[string]string

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&AsyncMetricsMeasurement{},
		&MetricsMeasurement{},
		&ProfileEventsMeasurement{},
		&StatusInfoMeasurement{},
	}
}

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) Catalog() string { return catalog }

func (i *Input) SetTags(m map[string]string) {
	if i.Tags == nil {
		i.Tags = make(map[string]string)
	}

	for k, v := range m {
		if _, ok := i.Tags[k]; !ok {
			i.Tags[k] = v
		}
	}
}

func (i *Input) ElectionEnabled() bool {
	return i.Election
}

func (i *Input) Run() {
	if i.setup() {
		return
	}

	tick := time.NewTicker(i.Interval)
	defer tick.Stop()

	i.l.Info("clickhouse start")

	for {
		if i.pause {
			i.l.Debug("clickhouse paused")
		} else {
			if err := i.collect(); err != nil {
				i.l.Warn(err)
			}
		}

		select {
		case <-datakit.Exit.Wait():
			i.l.Info("clickhouse exit")
			return

		case <-i.semStop.Wait():
			i.l.Info("clickhouse return")
			return

		case <-tick.C:

		case i.pause = <-i.chPause:
			// nil
		}
	}
}

func (i *Input) collect() error {
	if !i.isInitialized {
		if err := i.Init(); err != nil {
			return err
		}
	}

	start := time.Now()
	pts, err := i.doCollect()
	if err != nil {
		return err
	}
	if pts == nil {
		return fmt.Errorf("points got nil from doCollect")
	}

	if i.AsLogging != nil && i.AsLogging.Enable {
		// Feed measurement as logging.
		for _, pt := range pts {
			// We need to feed each point separately because
			// each point might have different measurement name.
			if err := i.Feeder.Feed(string(pt.Name()), point.Logging, []*point.Point{pt},
				&io.Option{CollectCost: time.Since(start)}); err != nil {
				i.Feeder.FeedLastError(err.Error(),
					io.WithLastErrorInput(inputName),
					io.WithLastErrorSource(i.Source),
				)
			}
		}
	} else {
		err := i.Feeder.Feed(inputName, point.Metric, pts,
			&io.Option{CollectCost: time.Since(start)})
		if err != nil {
			i.l.Errorf("Feed: %s", err)
			i.Feeder.FeedLastError(err.Error(),
				io.WithLastErrorInput(inputName),
				io.WithLastErrorSource(i.Source),
			)
		}
	}
	return nil
}

func (i *Input) doCollect() ([]*point.Point, error) {
	i.l.Debugf("collect URLs %v", i.URLs)

	// If Output is configured, data is written to local file specified by Output.
	// Data will no more be written to datakit io.
	if i.Output != "" {
		err := i.WriteMetricText2File()
		if err != nil {
			i.l.Errorf("WriteMetricText2File: %s", err.Error())
		}
		return nil, nil
	}

	pts, err := i.Collect()
	if err != nil {
		i.l.Errorf("Collect: %s", err)
		i.Feeder.FeedLastError(err.Error(),
			io.WithLastErrorInput(inputName),
			io.WithLastErrorSource(i.Source),
		)

		// Try testing the connect
		for _, u := range i.urls {
			if err := net.RawConnect(u.Hostname(), u.Port(), time.Second*3); err != nil {
				i.l.Errorf("failed to connect to %s:%s, %s", u.Hostname(), u.Port(), err)
			}
		}

		return nil, err
	}

	if pts == nil {
		return nil, fmt.Errorf("points got nil from Collect")
	}

	return pts, nil
}

// Collect collects metrics from all URLs.
func (i *Input) Collect() ([]*point.Point, error) {
	if i.pm == nil {
		return nil, fmt.Errorf("i.pm is nil")
	}
	var points []*point.Point
	for _, u := range i.URLs {
		uu, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		var pts []*point.Point
		if uu.Scheme != "http" && uu.Scheme != "https" {
			pts, err = i.CollectFromFile(u)
		} else {
			pts, err = i.CollectFromHTTP(u)
		}
		if err != nil {
			return nil, err
		}

		// append tags to points
		for k, v := range i.urlTags[u] {
			for _, pt := range pts {
				pt.AddTag([]byte(k), []byte(v))
			}
		}

		points = append(points, pts...)
	}

	return i.formatPointSuffixes(points), nil
}

const clickHouseAsyncMetrics string = "ClickHouseAsyncMetrics"

// formatPointSuffixes modify all points who have suffix.
func (i *Input) formatPointSuffixes(pts []*point.Point) []*point.Point {
	if len(pts) < 1 {
		return pts
	}
	// All suffix info, store average/total message, every loop is new.
	suffixes := suffixInfos()
	// The points timestamp, will use in average/total point.
	ts := pts[0].Time()

	instance := string(pts[0].GetTag([]byte("instance")))

	for j := 0; j < len(pts); j++ {
		// Only some ClickHouseAsyncMetrics metrics set have suffix.
		if string(pts[j].Name()) == clickHouseAsyncMetrics {
			i.formatPointSuffix(suffixes, pts[j])
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
		pt := point.NewPointV2([]byte(clickHouseAsyncMetrics),
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
func (i *Input) formatPointSuffix(suffixes []suffixInfo, pt *point.Point) {
	// Step 1, traverse to find the field(KV) to modify.
	willModifyKVs := []willModifyKV{}
	// kvs := pt.Fields()
	for _, kv := range pt.Fields() {
		key := string(kv.Key)
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
				if value, ok := tag.Val.(*point.Field_D); ok {
					tags[string(tag.Key)] = string(value.D)
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
		pt.MustAddTag([]byte(tagKey), []byte(tagVal))

		// Add field(KV).
		kv := &point.Field{
			Key: []byte(suffixes[v.suffixID].name),
			Val: pt.Fields().Get([]byte(v.key)).Val,
		}
		pt.MustAddKV(kv)

		// Delete old field(KV).
		pt.Del([]byte(v.key))
	}
}

func (i *Input) CollectFromHTTP(u string) ([]*point.Point, error) {
	if i.pm == nil {
		return nil, nil
	}
	return i.pm.CollectFromHTTPV2(u)
}

func (i *Input) CollectFromFile(filepath string) ([]*point.Point, error) {
	if i.pm == nil {
		return nil, nil
	}
	return i.pm.CollectFromFileV2(filepath)
}

// WriteMetricText2File collects from all URLs and then
// directly writes them to file specified by field Output.
func (i *Input) WriteMetricText2File() error {
	// Remove if file already exists.
	if _, err := os.Stat(i.Output); err == nil {
		if err := os.Remove(i.Output); err != nil {
			return err
		}
	}
	for _, u := range i.URLs {
		if err := i.pm.WriteMetricText2File(u); err != nil {
			return err
		}
		stat, err := os.Stat(i.Output)
		if err != nil {
			return err
		}
		if stat.Size() > i.MaxFileSize {
			return fmt.Errorf("file size is too large, max: %d, got: %d", i.MaxFileSize, stat.Size())
		}
	}
	return nil
}

func (i *Input) Terminate() {
	if i.semStop != nil {
		i.semStop.Close()
	}
}

func (i *Input) setup() bool {
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}
		time.Sleep(1 * time.Second) // sleep a while
		if err := i.Init(); err != nil {
			continue
		} else {
			break
		}
	}

	return false
}

func (i *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	select {
	case i.chPause <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (i *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	select {
	case i.chPause <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func (i *Input) Init() error {
	i.l = logger.SLogger(inputName)

	for _, u := range i.URLs {
		uu, err := url.Parse(u)
		if err != nil {
			return err
		}
		i.urls = append(i.urls, uu)

		// add extra `instance' tag, the tag take higher priority
		// over global tags.
		if !i.DisableInstanceTag {
			if _, ok := i.Tags["instance"]; !ok {
				i.Tags["instance"] = uu.Host
			}
		}

		if i.Election {
			i.urlTags[u] = inputs.MergeTags(i.Tagger.ElectionTags(), i.Tags, u)
		} else {
			i.urlTags[u] = inputs.MergeTags(i.Tagger.HostTags(), i.Tags, u)
		}
	}

	opts := []iprom.PromOption{
		iprom.WithLogger(i.l), // WithLogger must in the first
		iprom.WithSource(i.Source),
		iprom.WithTimeout(i.Timeout),
		iprom.WithKeepAlive(i.ConnectKeepAlive),
		iprom.WithIgnoreReqErr(i.IgnoreReqErr),
		iprom.WithMetricTypes(i.MetricTypes),
		iprom.WithMetricNameFilter(i.MetricNameFilter),
		iprom.WithMetricNameFilterIgnore(i.MetricNameFilterIgnore),
		iprom.WithMeasurementPrefix(i.MeasurementPrefix),
		iprom.WithMeasurementName(i.MeasurementName),
		iprom.WithMeasurements(i.Measurements),
		iprom.WithOutput(i.Output),
		iprom.WithMaxFileSize(i.MaxFileSize),
		iprom.WithTLSOpen(i.TLSOpen),
		iprom.WithUDSPath(i.UDSPath),
		iprom.WithCacertFile(i.CacertFile),
		iprom.WithCertFile(i.CertFile),
		iprom.WithKeyFile(i.KeyFile),
		iprom.WithTagsIgnore(i.TagsIgnore),
		iprom.WithTagsRename(i.TagsRename),
		iprom.WithAsLogging(i.AsLogging),
		iprom.WithIgnoreTagKV(i.IgnoreTagKV),
		iprom.WithHTTPHeaders(i.HTTPHeaders),
		iprom.WithDisableInfoTag(i.DisableInfoTag),
		iprom.WithAuth(i.Auth),
	}

	pm, err := iprom.NewProm(opts...)
	if err != nil {
		i.l.Warnf("clickhouse.NewProm: %s, ignored", err)
		return err
	}
	i.pm = pm
	i.isInitialized = true

	return nil
}

var maxPauseCh = inputs.ElectionPauseChannelLength

func NewProm() *Input {
	return &Input{
		chPause:     make(chan bool, maxPauseCh),
		MaxFileSize: defaultMaxFileSize,
		Source:      "clickhouse",
		Interval:    defaultIntervalDuration,
		Timeout:     time.Second * 30,
		Election:    true,
		Tags:        make(map[string]string),

		urlTags: map[string]urlTags{},

		semStop: cliutils.NewSem(),
		Feeder:  io.DefaultFeeder(),
		Tagger:  dkpt.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return NewProm()
	})
}
