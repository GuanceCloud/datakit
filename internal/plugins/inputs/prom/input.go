// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package prom scrape prometheus exporter metrics.
package prom

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
)

var _ inputs.ElectionInput = (*Input)(nil)

const (
	inputName               = "prom"
	catalog                 = "prom"
	defaultIntervalDuration = time.Second * 30

	// defaultMaxFileSize is the default max response body size, in bytes.
	// This field is used only when metrics are written to file, i.e. Output is configured.
	// If the size of response body is over defaultMaxFileSize, metrics will be discarded.
	// 32 MB.
	defaultMaxFileSize int64 = 32 * 1024 * 1024
)

type Input struct {
	Source           string        `toml:"source"`
	Interval         time.Duration `toml:"interval"`
	Timeout          time.Duration `toml:"timeout"`
	ConnectKeepAlive time.Duration `toml:"-"`

	URL                    string       `toml:"url,omitempty"` // Deprecated
	URLs                   []string     `toml:"urls"`
	StreamSize             int          `toml:"stream_size"`
	IgnoreReqErr           bool         `toml:"ignore_req_err"`
	MetricTypes            []string     `toml:"metric_types"`
	MetricNameFilter       []string     `toml:"metric_name_filter"`
	MetricNameFilterIgnore []string     `toml:"metric_name_filter_ignore"`
	MeasurementPrefix      string       `toml:"measurement_prefix"`
	MeasurementName        string       `toml:"measurement_name"`
	Measurements           []iprom.Rule `toml:"measurements"`
	KeepExistMetricName    bool         `toml:"keep_exist_metric_name"`
	Output                 string       `toml:"output"`
	MaxFileSize            int64        `toml:"max_file_size"`

	TLSOpen    bool   `toml:"tls_open"` // Deprecated
	UDSPath    string `toml:"uds_path"`
	CacertFile string `toml:"tls_ca"`   // Deprecated
	CertFile   string `toml:"tls_cert"` // Deprecated
	KeyFile    string `toml:"tls_key"`  // Deprecated
	*dknet.TLSClientConfig

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
	Feeder dkio.Feeder

	Election bool `toml:"election"`
	chPause  chan bool
	pause    bool

	Tagger datakit.GlobalTagger

	urls []*url.URL

	semStop *cliutils.Sem // start stop signal

	isInitialized bool

	urlTags map[string]urlTags

	// Input holds logger because prom have different types of instances.
	l *logger.Logger

	start        time.Time
	currentURL   string
	callbackFunc func([]*point.Point) error

	upStates map[string]int
}

type urlTags []struct {
	key,
	value string
}

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) SampleMeasurement() []inputs.Measurement { return nil }

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) Catalog() string { return catalog }

func (i *Input) ElectionEnabled() bool {
	return i.Election
}

func (i *Input) Run() {
	i.l = logger.SLogger(inputName + "/" + i.Source)

	tick := time.NewTicker(i.Interval)
	defer tick.Stop()

	i.l.Info("prom start")
	i.start = time.Now()

	for {
		if i.pause {
			i.l.Debug("prom paused")
		} else {
			i.tryInit()
			if err := i.collect(); err != nil {
				i.l.Warn(err)
			}
		}

		select {
		case <-datakit.Exit.Wait():
			i.l.Info("prom exit")
			return

		case <-i.semStop.Wait():
			i.l.Info("prom return")
			return

		case tt := <-tick.C:
			i.start = time.UnixMilli(inputs.AlignTimeMillSec(tt, i.start.UnixMilli(), i.Interval.Milliseconds()))

		case i.pause = <-i.chPause:
			// nil
		}
	}
}

func (i *Input) tryInit() {
	if i.isInitialized {
		return
	}

	if err := i.Init(); err != nil {
		i.l.Errorf("Init %s", err)
		return
	}
}

func (i *Input) collect() error {
	if !i.isInitialized {
		return fmt.Errorf("un initialized")
	}

	i.start = time.Now()
	i.l.Debugf("collect URLs %v", i.URLs)

	// If Output is configured, data is written to local file specified by Output.
	// Data will no more be written to datakit io.
	if i.Output != "" {
		err := i.WriteMetricText2File()
		if err != nil {
			i.l.Errorf("WriteMetricText2File: %s", err.Error())
		}
		return nil
	}

	err := i.collectFormURLs()
	if err != nil {
		i.l.Errorf("Collect: %s", err)

		ioname := inputName + "/" + i.Source
		i.Feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorSource(ioname),
		)

		// Try testing the connect
		for _, u := range i.urls {
			if err := dknet.RawConnect(u.Hostname(), u.Port(), time.Second*3); err != nil {
				i.l.Errorf("failed to connect to %s:%s, %s", u.Hostname(), u.Port(), err)
			}
		}

		return err
	}

	return nil
}

// Collect collects metrics from all URLs.
func (i *Input) collectFormURLs() error {
	if i.pm == nil {
		return fmt.Errorf("i.pm is nil")
	}

	for _, u := range i.URLs {
		i.setUpState(u)
		pts, err := i.collectFormSource(u)
		if err != nil {
			i.l.Errorf("failed to get pts from %s, %s", u, err)
			i.setErrUpState(u)
			i.FeedUpMetric(u)
			continue
		}

		if len(pts) > 0 {
			i.FeedUpMetric(u)

			if i.AsLogging != nil && i.AsLogging.Enable {
				// Feed measurement as logging.
				for _, pt := range pts {
					// We need to feed each point separately because
					// each point might have different measurement name.
					if err := i.Feeder.FeedV2(point.Logging, []*point.Point{pt},
						dkio.WithCollectCost(time.Since(i.start)),
						dkio.WithElection(i.Election),
						dkio.WithInputName(pt.Name()),
					); err != nil {
						i.Feeder.FeedLastError(err.Error(),
							metrics.WithLastErrorInput(inputName),
							metrics.WithLastErrorSource(inputName+"/"+i.Source),
						)
						i.l.Errorf("feed logging: %s", err)
					}
				}
			} else if err := i.Feeder.FeedV2(point.Metric, pts,
				dkio.WithCollectCost(time.Since(i.start)),
				dkio.WithElection(i.Election),
				dkio.WithInputName(inputName+"/"+i.Source),
			); err != nil {
				i.Feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(inputName),
					metrics.WithLastErrorSource(inputName+"/"+i.Source),
				)
				i.l.Errorf("feed measurement: %s", err)
			}
		}
	}
	return nil
}

func (i *Input) collectFormSource(u string) (pts []*point.Point, err error) {
	uu, err := url.Parse(u)
	if err != nil {
		return
	}

	i.currentURL = u

	if uu.Scheme != "http" && uu.Scheme != "https" {
		pts, err = i.CollectFromFile(u)
	} else {
		pts, err = i.CollectFromHTTP(u)
	}
	if err != nil {
		i.l.Errorf("failed to get pts from %s, %s", u, err)
		return
	}

	// Append tags to points
	for _, v := range i.urlTags[u] {
		for _, pt := range pts {
			pt.AddTag(v.key, v.value)
		}
	}

	return
}

func (i *Input) CollectFromHTTP(u string) ([]*point.Point, error) {
	if i.pm == nil {
		return nil, nil
	}
	return i.pm.CollectFromHTTPV2(u, iprom.WithTimestamp(i.start.UnixNano()))
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

type promHandleCallback func([]*point.Point) error

func (i *Input) defaultHandleCallback() promHandleCallback {
	return func(pts []*point.Point) error {
		// Append tags to points
		for _, v := range i.urlTags[i.currentURL] {
			for _, pt := range pts {
				pt.AddTag(v.key, v.value)
			}
		}

		if len(pts) < 1 {
			return nil
		}

		if i.AsLogging != nil && i.AsLogging.Enable {
			// Feed measurement as logging.
			for _, pt := range pts {
				// We need to feed each point separately because
				// each point might have different measurement name.
				if err := i.Feeder.FeedV2(point.Logging, []*point.Point{pt},
					dkio.WithCollectCost(time.Since(i.start)),
					dkio.WithElection(i.Election),
					dkio.WithInputName(pt.Name()),
				); err != nil {
					i.Feeder.FeedLastError(err.Error(),
						metrics.WithLastErrorInput(inputName),
						metrics.WithLastErrorSource(inputName+"/"+i.Source),
					)
					i.l.Errorf("feed logging: %s", err)
				}
			}
		} else if err := i.Feeder.FeedV2(point.Metric, pts,
			dkio.WithCollectCost(time.Since(i.start)),
			dkio.WithElection(i.Election),
			dkio.WithInputName(inputName+"/"+i.Source),
		); err != nil {
			i.Feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorSource(inputName+"/"+i.Source),
			)
			i.l.Errorf("feed measurement: %s", err)
		}
		i.FeedUpMetric(i.currentURL)
		return nil
	}
}

func (i *Input) Init() error {
	i.l = logger.SLogger(inputName + "/" + i.Source)

	if i.URL != "" && len(i.URLs) == 0 {
		i.URLs = append(i.URLs, i.URL)
	}

	i.urls = make([]*url.URL, 0)

	for _, u := range i.URLs {
		uu, err := url.Parse(u)
		if err != nil {
			return err
		}
		i.urls = append(i.urls, uu)

		var globalTags map[string]string
		if i.Election {
			globalTags = i.Tagger.ElectionTags()
			i.l.Infof("add global election tags %q", globalTags)
		} else {
			globalTags = i.Tagger.HostTags()
			i.l.Infof("add global host tags %q", globalTags)
		}

		temp := inputs.MergeTags(globalTags, i.Tags, u)
		// Add extra `instance` tag, from url
		if !i.DisableInstanceTag {
			if _, ok := temp["instance"]; !ok {
				temp["instance"] = uu.Host
			}
		}
		tempTags := urlTags{}
		for k, v := range temp {
			tempTags = append(tempTags, struct {
				key, value string
			}{key: k, value: v})
		}
		i.urlTags[u] = tempTags
	}

	if i.StreamSize > 0 && i.callbackFunc == nil { // set callback on streamming-mode
		i.callbackFunc = i.defaultHandleCallback()
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
		iprom.KeepExistMetricName(i.KeepExistMetricName),
		iprom.WithOutput(i.Output),
		iprom.WithMaxFileSize(i.MaxFileSize),
		iprom.WithTLSOpen(i.TLSOpen),
		iprom.WithUDSPath(i.UDSPath),
		iprom.WithCacertFiles([]string{i.CacertFile}),
		iprom.WithCertFile(i.CertFile),
		iprom.WithKeyFile(i.KeyFile),
		iprom.WithTLSClientConfig(i.TLSClientConfig),
		iprom.WithTagsIgnore(i.TagsIgnore),
		iprom.WithTagsRename(i.TagsRename),
		iprom.WithAsLogging(i.AsLogging),
		iprom.WithIgnoreTagKV(i.IgnoreTagKV),
		iprom.WithHTTPHeaders(i.HTTPHeaders),
		iprom.WithDisableInfoTag(i.DisableInfoTag),
		iprom.WithMaxBatchCallback(i.StreamSize, i.callbackFunc),
		iprom.WithAuth(i.Auth),
	}

	pm, err := iprom.NewProm(opts...)
	if err != nil {
		i.l.Warnf("prom.NewProm: %s, ignored", err)
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
		Source:      "prom",
		Interval:    defaultIntervalDuration,
		Timeout:     time.Second * 30,
		StreamSize:  1,
		Election:    true,
		Tags:        make(map[string]string),

		urlTags: map[string]urlTags{},

		semStop:  cliutils.NewSem(),
		Feeder:   dkio.DefaultFeeder(),
		Tagger:   datakit.DefaultGlobalTagger(),
		upStates: make(map[string]int),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return NewProm()
	})
}
