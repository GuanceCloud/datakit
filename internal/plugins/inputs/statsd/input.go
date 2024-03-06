// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package statsd serve a UDP/TCP(not used) server to handle statsd metrics.
package statsd

import (
	"net"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	istatsd "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/statsd"
)

const (
	minInterval                = time.Second
	maxInterval                = time.Minute
	defaultProtocol            = "udp"
	defaultAllowPendingMessage = 128
	inputName                  = "statsd"
	catalog                    = "statsd"
	defaultIOName              = "statsd/-/-"
)

// Input statsd allows the importing of statsd and dogstatsd data.
type Input struct {
	Source   string        `toml:"source"`
	Interval time.Duration `toml:"interval"`
	// Protocol used on listener - udp or tcp
	Protocol string `toml:"protocol"`

	// Address & Port to serve from
	ServiceAddress string `toml:"service_address"`

	// Tag request metric. Used for distinguish feed metric name.
	StatsdSourceKey string `toml:"statsd_source_key"`
	StatsdHostKey   string `toml:"statsd_host_key"`
	SaveAboveKey    bool   `toml:"save_above_key"`

	// Number of messages allowed to queue up in between calls to Gather. If this
	// fills up, packets will get dropped until the next Gather interval is ran.
	AllowedPendingMessages int `toml:"allowed_pending_messages"`

	// Percentiles specifies the percentiles that will be calculated for timing
	// and histogram stats.
	Percentiles     []float64 `toml:"percentiles"`
	PercentileLimit int       `toml:"percentile_limit"`

	DeleteGauges   bool `toml:"delete_gauges"`
	DeleteCounters bool `toml:"delete_counters"`
	DeleteSets     bool `toml:"delete_sets"`
	DeleteTimings  bool `toml:"delete_timings"`
	ConvertNames   bool
	// Counter metric is float in new Datakit version, set true if want be int.
	SetCounterInt bool `toml:"set_counter_int"`

	// MetricSeparator is the separator between parts of the metric name.
	MetricSeparator string `toml:"metric_separator"`
	// This flag enables parsing of tags in the dogstatsd extension to the
	// statsd protocol (http://docs.datadoghq.com/guides/dogstatsd/)
	ParseDataDogTags bool // depreciated in 1.10; use datadog_extensions

	// Parses extensions to statsd in the datadog statsd format
	// currently supports metrics and datadog tags.
	// http://docs.datadoghq.com/guides/dogstatsd/
	DataDogExtensions bool `toml:"datadog_extensions"`

	// Parses distribution metrics in the datadog statsd format.
	// Requires the DataDogExtension flag to be enabled.
	// https://docs.datadoghq.com/developers/metrics/types/?tab=distribution#definition
	DataDogDistributions bool `toml:"datadog_distributions"`

	// UDPPacketSize is deprecated, it's only here for legacy support
	// we now always create 1 max size buffer and then copy only what we need
	// into the in channel
	// see https://github.com/influxdata/telegraf/pull/992
	UDPPacketSize int `toml:"udp_packet_size"`

	ReadBufferSize    int               `toml:"read_buffer_size"`
	DropTags          []string          `toml:"drop_tags"`
	MetricMapping     []string          `toml:"metric_mapping"`
	Tags              map[string]string `toml:"tags"`
	MaxTCPConnections int               `toml:"max_tcp_connections"`
	TCPKeepAlive      bool              `toml:"tcp_keep_alive"`

	// Max duration for each metric to stay cached without being updated.
	MaxTTL time.Duration `toml:"max_ttl"`

	// Protocol listeners
	UDPlistener *net.UDPConn
	TCPlistener *net.TCPListener

	semStop    *cliutils.Sem // start stop signal
	Feeder     dkio.Feeder
	Tagger     datakit.GlobalTagger
	taggerTags map[string]string
	Col        *istatsd.Collector // The real collector

	isInitialized bool
	l             *logger.Logger
}

func (ipt *Input) SampleConfig() string {
	return sampleConfig
}

func (ipt *Input) Catalog() string {
	return catalog
}

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return nil
}

func (ipt *Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (ipt *Input) setup() error {
	if ipt.isInitialized {
		return nil
	}

	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)
	ipt.l = logger.SLogger(ipt.Source)

	if ipt.ParseDataDogTags {
		ipt.DataDogExtensions = true
		ipt.l.Warn("'parse_data_dog_tags' config option is deprecated, please use 'datadog_extensions' instead")
	}

	opts := []istatsd.CollectorOption{
		istatsd.WithLogger(ipt.l),
		istatsd.WithProtocol(ipt.Protocol),
		istatsd.WithServiceAddress(ipt.ServiceAddress),
		istatsd.WithStatsdSourceKey(ipt.StatsdSourceKey),
		istatsd.WithStatsdHostKey(ipt.StatsdHostKey),
		istatsd.WithSaveAboveKey(ipt.SaveAboveKey),
		istatsd.WithAllowedPendingMessages(ipt.AllowedPendingMessages),
		istatsd.WithPercentiles(ipt.Percentiles),
		istatsd.WithPercentileLimit(ipt.PercentileLimit),
		istatsd.WithDeleteGauges(ipt.DeleteGauges),
		istatsd.WithDeleteCounters(ipt.DeleteCounters),
		istatsd.WithSetCounterInt(ipt.SetCounterInt),
		istatsd.WithDeleteSets(ipt.DeleteSets),
		istatsd.WithDeleteTimings(ipt.DeleteTimings),
		istatsd.WithConvertNames(ipt.ConvertNames),
		istatsd.WithMetricSeparator(ipt.MetricSeparator),
		istatsd.WithDataDogExtensions(ipt.DataDogExtensions),
		istatsd.WithDataDogDistributions(ipt.DataDogDistributions),
		istatsd.WithUDPPacketSize(ipt.UDPPacketSize),
		istatsd.WithReadBufferSize(ipt.ReadBufferSize),
		istatsd.WithDropTags(ipt.DropTags),
		istatsd.WithMetricMapping(ipt.MetricMapping),
		istatsd.WithTags(ipt.Tags),
		istatsd.WithMaxTCPConnections(ipt.MaxTCPConnections),
		istatsd.WithTCPKeepAlive(ipt.TCPKeepAlive),
		istatsd.WithMaxTTL(ipt.MaxTTL),
	}

	col, err := istatsd.NewCollector(ipt.UDPlistener, ipt.TCPlistener, opts...)
	if err != nil {
		return err
	}

	ipt.Col = col
	ipt.isInitialized = true
	return nil
}

func (ipt *Input) Collect() error {
	points, err := ipt.Col.GetPoints()
	if err != nil {
		ipt.Feeder.FeedLastError(err.Error(),
			dkio.WithLastErrorInput(inputName),
			dkio.WithLastErrorSource(ipt.Source),
		)
		ipt.l.Errorf("GetPoints: %v", err)
	}

	if len(points) > 0 {
		ipt.feedBatch(points)
	} else {
		ipt.l.Debug("GetPoints 0 pts")
	}

	return nil
}

func (ipt *Input) feedBatch(points []*point.Point) {
	start := time.Now()

	pts := []*point.Point{}
	for i, v := range points {
		for kk, vv := range ipt.taggerTags {
			v.AddTag(kk, vv)
		}
		pts = append(pts, v)

		// i >= len(points)-1 --> last batch
		// len(pts) >= 1024 --> 1024 pts per batch
		if i >= len(points)-1 || len(pts) >= 1024 {
			if err := ipt.Feeder.Feed(ipt.Source, point.Metric, pts,
				&dkio.Option{CollectCost: time.Since(start)}); err != nil {
				ipt.l.Errorf("Feed: %v", err)
				ipt.Feeder.FeedLastError(err.Error(),
					dkio.WithLastErrorInput(inputName),
					dkio.WithLastErrorSource(ipt.Source),
				)
			}

			pts = []*point.Point{}
		}
	}
}

func (ipt *Input) Run() {
	for {
		select {
		case <-datakit.Exit.Wait():
			return
		default:
		}

		if err := ipt.setup(); err != nil {
			ipt.Feeder.FeedLastError(err.Error(),
				dkio.WithLastErrorInput(inputName),
			)
			time.Sleep(time.Second * 5)
			continue
		}
		break
	}

	ipt.l.Infof("Started the statsd service on %q", ipt.ServiceAddress)

	ipt.taggerTags = inputs.MergeTags(ipt.Tagger.HostTags(), ipt.taggerTags, "")

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			ipt.l.Info("statsd exited")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			ipt.l.Info("statsd return")
			return

		case <-tick.C:
			ipt.l.Debugf("try gathering...")
			if err := ipt.Collect(); err != nil {
				ipt.l.Errorf("Collect: %s", err)
			}
		}
	}
}

func (ipt *Input) exit() {
	ipt.Col.Exit()
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func DefaultInput() *Input {
	return &Input{
		Source:                 defaultIOName,
		Interval:               time.Second * 10,
		Protocol:               defaultProtocol,
		ServiceAddress:         ":8125",
		MaxTCPConnections:      250,
		TCPKeepAlive:           false,
		MetricSeparator:        "_",
		AllowedPendingMessages: defaultAllowPendingMessage,
		DeleteCounters:         true,
		DeleteGauges:           true,
		DeleteSets:             true,
		DeleteTimings:          true,
		MaxTTL:                 0,

		semStop: cliutils.NewSem(),
		Feeder:  dkio.DefaultFeeder(),
		Tagger:  datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return DefaultInput()
	})
}
