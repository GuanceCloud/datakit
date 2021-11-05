// Package statsd serve a UDP/TCP(not used) server to handle statsd metrics.
package statsd

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf/plugins/parsers/graphite"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	// UDPMaxPacketSize is the UDP packet limit, see
	// https://en.wikipedia.org/wiki/User_Datagram_Protocol#Packet_structure
	UDPMaxPacketSize int = 64 * 1024

	defaultFieldName = "value"

	defaultProtocol = "udp"

	defaultSeparator           = "_"
	defaultAllowPendingMessage = 10000

	parserGoRoutines = 5

	inputName = "statsd"
	catalog   = "statsd"
)

var l = logger.DefaultSLogger("statsd")

// Statsd allows the importing of statsd and dogstatsd data.
type input struct {
	// Protocol used on listener - udp or tcp
	Protocol string `toml:"protocol"`

	// Address & Port to serve from
	ServiceAddress string

	// Number of messages allowed to queue up in between calls to Gather. If this
	// fills up, packets will get dropped until the next Gather interval is ran.
	AllowedPendingMessages int

	// Percentiles specifies the percentiles that will be calculated for timing
	// and histogram stats.
	Percentiles     []float64
	PercentileLimit int

	DeleteGauges   bool
	DeleteCounters bool
	DeleteSets     bool
	DeleteTimings  bool
	ConvertNames   bool

	// MetricSeparator is the separator between parts of the metric name.
	MetricSeparator string
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

	ReadBufferSize int `toml:"read_buffer_size"`

	DropTags      []string `toml:"drop_tags"`
	MetricMapping []string `toml:"metric_mapping"`
	mmap          map[string]string

	Tags map[string]string `toml:"tags"`

	sync.Mutex
	// Lock for preventing a data race during resource cleanup
	cleanup sync.Mutex
	wg      sync.WaitGroup
	// accept channel tracks how many active connections there are, if there
	// is an available bool in accept, then we are below the maximum and can
	// accept the connection
	accept chan bool
	// drops tracks the number of dropped metrics.
	drops int

	// Channel for all incoming statsd packets
	in   chan job
	done chan struct{}

	// Cache gauges, counters & sets so they can be aggregated as they arrive
	// gauges and counters map measurement/tags hash -> field name -> metrics
	// sets and timings map measurement/tags hash -> metrics
	// distributions aggregate measurement/tags and are published directly
	gauges        map[string]cachedgauge
	counters      map[string]cachedcounter
	sets          map[string]cachedset
	timings       map[string]cachedtimings
	distributions []cacheddistributions

	// bucket -> influx templates
	Templates []string // NOTE: Deprecated

	// Protocol listeners
	UDPlistener *net.UDPConn
	TCPlistener *net.TCPListener

	// track current connections so we can close them in Stop()
	conns map[string]*net.TCPConn

	MaxTCPConnections int `toml:"max_tcp_connections"`

	TCPKeepAlive       bool              `toml:"tcp_keep_alive"`
	TCPKeepAlivePeriod *datakit.Duration `toml:"tcp_keep_alive_period"`

	// Max duration for each metric to stay cached without being updated.
	MaxTTL datakit.Duration `toml:"max_ttl"`

	graphiteParser *graphite.GraphiteParser

	acc *accumulator

	// A pool of byte slices to handle parsing
	bufPool sync.Pool

	semStop          *cliutils.Sem // start stop signal
	semStopCompleted *cliutils.Sem // stop completed signal
}

type job struct {
	*bytes.Buffer
	time.Time
	Addr string
}

// One statsd metric, form is <bucket>:<value>|<mtype>|@<samplerate>.
type metric struct {
	name       string
	field      string
	bucket     string
	hash       string
	intvalue   int64
	floatvalue float64
	strvalue   string
	mtype      string
	additive   bool
	samplerate float64
	tags       map[string]string
}

type cachedset struct {
	name      string
	fields    map[string]map[string]bool
	tags      map[string]string
	expiresAt time.Time
}

type cachedgauge struct {
	name      string
	fields    map[string]interface{}
	tags      map[string]string
	expiresAt time.Time
}

type cachedcounter struct {
	name      string
	fields    map[string]interface{}
	tags      map[string]string
	expiresAt time.Time
}

type cachedtimings struct {
	name      string
	fields    map[string]RunningStats
	tags      map[string]string
	expiresAt time.Time
}

type cacheddistributions struct {
	name  string
	value float64
	tags  map[string]string
}

const sampleConfig = `
[[inputs.statsd]]
  protocol = "udp"

  ## Address and port to host UDP listener on
  service_address = ":8125"

  delete_gauges = true
  delete_counters = true
  delete_sets = true
  delete_timings = true

  ## Percentiles to calculate for timing & histogram stats
  percentiles = [50.0, 90.0, 99.0, 99.9, 99.95, 100.0]

  ## separator to use between elements of a statsd metric
  metric_separator = "_"

  ## Parses tags in the datadog statsd format
  ## http://docs.datadoghq.com/guides/dogstatsd/
  parse_data_dog_tags = true

  ## Parses datadog extensions to the statsd format
  datadog_extensions = true

  ## Parses distributions metric as specified in the datadog statsd format
  ## https://docs.datadoghq.com/developers/metrics/types/?tab=distribution#definition
  datadog_distributions = true

	## We do not need following tags(they may create tremendous of time-series under influxdb's logic)
	# Examples:
	# "runtime-id", "metric-type"
  drop_tags = [ ]

  # All metric-name prefixed with 'jvm_' are set to influxdb's measurement 'jvm'
  # All metric-name prefixed with 'stats_' are set to influxdb's measurement 'stats'
  # Examples:
  # "stats_:stats", "jvm_:jvm"
	metric_mapping = [ ]

  ## Number of UDP messages allowed to queue up, once filled,
  ## the statsd server will start dropping packets
  allowed_pending_messages = 10000

  ## Number of timing/histogram values to track per-measurement in the
  ## calculation of percentiles. Raising this limit increases the accuracy
  ## of percentiles but also increases the memory usage and cpu time.
  percentile_limit = 1000

  ## Max duration (TTL) for each metric to stay cached/reported without being updated.
  #max_ttl = "1000h"

  [inputs.statsd.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`

func (ipt *input) SampleConfig() string {
	return sampleConfig
}

func (ipt *input) Catalog() string {
	return catalog
}

func (ipt *input) SampleMeasurement() []inputs.Measurement {
	return nil
}

func (ipt *input) AvailableArchs() []string {
	return datakit.AllOS
}

func (ipt *input) setup() error {
	if ipt.ParseDataDogTags {
		ipt.DataDogExtensions = true
		l.Warn("'parse_data_dog_tags' config option is deprecated, please use 'datadog_extensions' instead")
	}

	ipt.acc = &accumulator{ref: ipt}

	// Make data structures
	ipt.gauges = make(map[string]cachedgauge)
	ipt.counters = make(map[string]cachedcounter)
	ipt.sets = make(map[string]cachedset)
	ipt.timings = make(map[string]cachedtimings)
	ipt.distributions = make([]cacheddistributions, 0)

	ipt.Lock()
	defer ipt.Unlock()

	ipt.in = make(chan job, ipt.AllowedPendingMessages)
	ipt.done = make(chan struct{})
	ipt.accept = make(chan bool, ipt.MaxTCPConnections)
	ipt.conns = make(map[string]*net.TCPConn)
	ipt.bufPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	for i := 0; i < ipt.MaxTCPConnections; i++ {
		ipt.accept <- true
	}

	if ipt.ConvertNames {
		l.Warn("'convert_names' config option is deprecated, please use 'metric_separator' instead")
	}

	if ipt.MetricSeparator == "" {
		ipt.MetricSeparator = defaultSeparator
	}

	if len(ipt.MetricMapping) > 0 {
		ipt.setupMmap()
	}

	if ipt.isUDP() {
		if err := ipt.setupUDPServer(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("TCP not supported")
		// TODO: not testing
		// s.setupTCPServer()
	}

	l.Infof("starting %d parser worker...", parserGoRoutines)
	for i := 1; i <= parserGoRoutines; i++ {
		// Start the line parser
		ipt.wg.Add(1)
		go func(idx int) {
			defer ipt.wg.Done()
			ipt.parser(idx)
		}(i)
	}

	return nil
}

func (ipt *input) setupMmap() {
	ipt.mmap = map[string]string{}

	for _, mm := range ipt.MetricMapping {
		arr := strings.SplitN(mm, ":", 2)
		if len(arr) != 2 {
			l.Warnf("invalid MetricMapping: %s, ignored", mm)
			continue
		}

		ipt.mmap[arr[0]] = arr[1]
	}
}

func (ipt *input) Run() {
	l = logger.SLogger(inputName)

	for {
		select {
		case <-datakit.Exit.Wait():
			return
		default:
		}

		if err := ipt.setup(); err != nil {
			io.FeedLastError(inputName, err.Error())
			time.Sleep(time.Second * 5)
			continue
		}
		break
	}

	l.Infof("Started the statsd service on %q", ipt.ServiceAddress)
	tick := time.NewTicker(time.Second * 10)
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Info("statsd exited")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Info("statsd return")

			if ipt.semStopCompleted != nil {
				ipt.semStopCompleted.Close()
			}
			return

		case <-tick.C:
			l.Debugf("try gathering...")
			ipt.gather()
		}
	}
}

func (ipt *input) exit() {
	ipt.stop()
}

func (ipt *input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()

		// wait stop completed
		if ipt.semStopCompleted != nil {
			for range ipt.semStopCompleted.Wait() {
				return
			}
		}
	}
}

// aggregate takes in a metric. It then
// aggregates and caches the current value(s). It does not deal with the
// Delete* options, because those are dealt with in the Gather function.
func (ipt *input) aggregate(m metric) {
	ipt.Lock()
	defer ipt.Unlock()

	switch m.mtype {
	case "d":
		if ipt.DataDogExtensions && ipt.DataDogDistributions {
			cached := cacheddistributions{
				name:  m.name,
				value: m.floatvalue,
				tags:  m.tags,
			}
			ipt.distributions = append(ipt.distributions, cached)
		}
	case "ms", "h":
		// Check if the measurement exists
		cached, ok := ipt.timings[m.hash]
		if !ok {
			cached = cachedtimings{
				name:   m.name,
				fields: make(map[string]RunningStats),
				tags:   m.tags,
			}
		}
		// Check if the field exists. If we've not enabled multiple fields per timer
		// this will be the default field name, eg. "value"
		field, ok := cached.fields[m.field]
		if !ok {
			field = RunningStats{
				PercLimit: ipt.PercentileLimit,
			}
		}
		if m.samplerate > 0 {
			for i := 0; i < int(1.0/m.samplerate); i++ {
				field.AddValue(m.floatvalue)
			}
		} else {
			field.AddValue(m.floatvalue)
		}
		cached.fields[m.field] = field
		cached.expiresAt = time.Now().Add(ipt.MaxTTL.Duration)
		ipt.timings[m.hash] = cached
	case "c":
		// check if the measurement exists
		cached, ok := ipt.counters[m.hash]
		if !ok {
			cached = cachedcounter{
				name:   m.name,
				fields: make(map[string]interface{}),
				tags:   m.tags,
			}
		}
		// check if the field exists
		_, ok = cached.fields[m.field]
		if !ok {
			cached.fields[m.field] = int64(0)
		}
		cached.fields[m.field] = cached.fields[m.field].(int64) + m.intvalue
		cached.expiresAt = time.Now().Add(ipt.MaxTTL.Duration)
		ipt.counters[m.hash] = cached
	case "g":
		// check if the measurement exists
		cached, ok := ipt.gauges[m.hash]
		if !ok {
			cached = cachedgauge{
				name:   m.name,
				fields: make(map[string]interface{}),
				tags:   m.tags,
			}
		}
		// check if the field exists
		_, ok = cached.fields[m.field]
		if !ok {
			cached.fields[m.field] = float64(0)
		}
		if m.additive {
			cached.fields[m.field] = cached.fields[m.field].(float64) + m.floatvalue
		} else {
			cached.fields[m.field] = m.floatvalue
		}

		cached.expiresAt = time.Now().Add(ipt.MaxTTL.Duration)
		ipt.gauges[m.hash] = cached
	case "s":
		// check if the measurement exists
		cached, ok := ipt.sets[m.hash]
		if !ok {
			cached = cachedset{
				name:   m.name,
				fields: make(map[string]map[string]bool),
				tags:   m.tags,
			}
		}
		// check if the field exists
		_, ok = cached.fields[m.field]
		if !ok {
			cached.fields[m.field] = make(map[string]bool)
		}
		cached.fields[m.field][m.strvalue] = true
		cached.expiresAt = time.Now().Add(ipt.MaxTTL.Duration)
		ipt.sets[m.hash] = cached
	}
}

func (ipt *input) stop() {
	ipt.Lock()
	l.Infof("Stopping the statsd service")
	close(ipt.done)
	if ipt.isUDP() && ipt.UDPlistener != nil {
		// Ignore the returned error as we cannot do anything about it anyway
		//nolint:errcheck,revive
		if err := ipt.UDPlistener.Close(); err != nil {
			l.Warnf("Close: %s, ignored", err)
		}
	} else if ipt.TCPlistener != nil {
		// Ignore the returned error as we cannot do anything about it anyway
		//nolint:errcheck,revive
		if err := ipt.TCPlistener.Close(); err != nil {
			l.Warnf("Close: %s, ignored", err)
		}

		// Close all open TCP connections
		//  - get all conns from the s.conns map and put into slice
		//  - this is so the forget() function doesnt conflict with looping
		//    over the s.conns map
		var conns []*net.TCPConn
		ipt.cleanup.Lock()
		for _, conn := range ipt.conns {
			conns = append(conns, conn)
		}
		ipt.cleanup.Unlock()
		for _, conn := range conns {
			// Ignore the returned error as we cannot do anything about it anyway
			//nolint:errcheck,revive
			if err := conn.Close(); err != nil {
				l.Warnf("Close: %s, ignored", err)
			}
		}
	}

	ipt.Unlock()

	ipt.wg.Wait()

	ipt.Lock()
	close(ipt.in)
	l.Infof("Stopped listener service on %q", ipt.ServiceAddress)
	ipt.Unlock()
}

// IsUDP returns true if the protocol is UDP, false otherwise.
func (ipt *input) isUDP() bool {
	return strings.HasPrefix(ipt.Protocol, "udp")
}

func (ipt *input) expireCachedMetrics() {
	// If Max TTL wasn't configured, skip expiration.
	if ipt.MaxTTL.Duration == 0 {
		return
	}

	now := time.Now()

	for key, cached := range ipt.gauges {
		if now.After(cached.expiresAt) {
			delete(ipt.gauges, key)
		}
	}

	for key, cached := range ipt.sets {
		if now.After(cached.expiresAt) {
			delete(ipt.sets, key)
		}
	}

	for key, cached := range ipt.timings {
		if now.After(cached.expiresAt) {
			delete(ipt.timings, key)
		}
	}

	for key, cached := range ipt.counters {
		if now.After(cached.expiresAt) {
			delete(ipt.counters, key)
		}
	}
}

func defaultInput() *input {
	return &input{
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

		semStop:          cliutils.NewSem(),
		semStopCompleted: cliutils.NewSem(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
