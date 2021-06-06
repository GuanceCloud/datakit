package statsd

import (
	"bytes"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf/plugins/parsers/graphite"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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

var (
	l = logger.DefaultSLogger("statsd")
)

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
}

type job struct {
	*bytes.Buffer
	time.Time
	Addr string
}

// One statsd metric, form is <bucket>:<value>|<mtype>|@<samplerate>
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

  drop_tags = [
	## Examples:
	## we do not need following tags(they may create tremendous of time-series under influxdb's logic)
	# "runtime-id", "metric-type"
	]

  metric_mapping = [
	  # Examples:
    # all metric-name prefixed with 'jvm_' are set to influxdb's measurement 'jvm'
    # "jvm_:jvm",

    # all metric-name prefixed with 'stats_' are set to influxdb's measurement 'stats'
    # "stats_:stats",
  ]

  ## Number of UDP messages allowed to queue up, once filled,
  ## the statsd server will start dropping packets
  allowed_pending_messages = 10000

  ## Number of timing/histogram values to track per-measurement in the
  ## calculation of percentiles. Raising this limit increases the accuracy
  ## of percentiles but also increases the memory usage and cpu time.
  percentile_limit = 1000

  ## Max duration (TTL) for each metric to stay cached/reported without being updated.
  #max_ttl = "1000h"

  [input.statsd.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`

func (s *input) SampleConfig() string {
	return sampleConfig
}

func (s *input) Catalog() string {
	return catalog
}

func (s *input) SampleMeasurement() []inputs.Measurement {
	return nil
}

func (s *input) AvailableArchs() []string {
	return datakit.AllOS
}

func (s *input) setup() {
	if s.ParseDataDogTags {
		s.DataDogExtensions = true
		l.Warn("'parse_data_dog_tags' config option is deprecated, please use 'datadog_extensions' instead")
	}

	s.acc = &accumulator{ref: s}

	// Make data structures
	s.gauges = make(map[string]cachedgauge)
	s.counters = make(map[string]cachedcounter)
	s.sets = make(map[string]cachedset)
	s.timings = make(map[string]cachedtimings)
	s.distributions = make([]cacheddistributions, 0)

	s.Lock()
	defer s.Unlock()

	s.in = make(chan job, s.AllowedPendingMessages)
	s.done = make(chan struct{})
	s.accept = make(chan bool, s.MaxTCPConnections)
	s.conns = make(map[string]*net.TCPConn)
	s.bufPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	for i := 0; i < s.MaxTCPConnections; i++ {
		s.accept <- true
	}

	if s.ConvertNames {
		l.Warn("'convert_names' config option is deprecated, please use 'metric_separator' instead")
	}

	if s.MetricSeparator == "" {
		s.MetricSeparator = defaultSeparator
	}

	if len(s.MetricMapping) > 0 {
		s.setupMmap()
	}

	if s.isUDP() {
		s.setupUDPServer()
	} else {
		// TODO: not testing
		// s.setupTCPServer()
	}

	for i := 1; i <= parserGoRoutines; i++ {
		// Start the line parser
		s.wg.Add(1)
		go func(idx int) {
			defer s.wg.Done()
			s.parser(idx)
		}(i)
	}
}

func (s *input) setupMmap() {
	s.mmap = map[string]string{}

	for _, mm := range s.MetricMapping {
		arr := strings.SplitN(mm, ":", 2)
		if len(arr) != 2 {
			l.Warnf("invalid MetricMapping: %s, ignored", mm)
			continue
		}

		s.mmap[arr[0]] = arr[1]
	}
}

func (s *input) Run() {
	l = logger.SLogger(inputName)

	s.setup()

	l.Infof("Started the statsd service on %q", s.ServiceAddress)
	tick := time.NewTicker(time.Second * 10)
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			s.stop()
			l.Info("statsd exited")
			return
		case <-tick.C:
			l.Debugf("try gathering...")
			s.gather()
			// TODO: feedIO
		}
	}
}

// aggregate takes in a metric. It then
// aggregates and caches the current value(s). It does not deal with the
// Delete* options, because those are dealt with in the Gather function.
func (s *input) aggregate(m metric) {
	s.Lock()
	defer s.Unlock()

	switch m.mtype {
	case "d":
		if s.DataDogExtensions && s.DataDogDistributions {
			cached := cacheddistributions{
				name:  m.name,
				value: m.floatvalue,
				tags:  m.tags,
			}
			s.distributions = append(s.distributions, cached)
		}
	case "ms", "h":
		// Check if the measurement exists
		cached, ok := s.timings[m.hash]
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
				PercLimit: s.PercentileLimit,
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
		cached.expiresAt = time.Now().Add(s.MaxTTL.Duration)
		s.timings[m.hash] = cached
	case "c":
		// check if the measurement exists
		cached, ok := s.counters[m.hash]
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
		cached.expiresAt = time.Now().Add(s.MaxTTL.Duration)
		s.counters[m.hash] = cached
	case "g":
		// check if the measurement exists
		cached, ok := s.gauges[m.hash]
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

		cached.expiresAt = time.Now().Add(s.MaxTTL.Duration)
		s.gauges[m.hash] = cached
	case "s":
		// check if the measurement exists
		cached, ok := s.sets[m.hash]
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
		cached.expiresAt = time.Now().Add(s.MaxTTL.Duration)
		s.sets[m.hash] = cached
	}
}

func (s *input) stop() {
	s.Lock()
	l.Infof("Stopping the statsd service")
	close(s.done)
	if s.isUDP() {
		// Ignore the returned error as we cannot do anything about it anyway
		//nolint:errcheck,revive
		s.UDPlistener.Close()
	} else {
		// Ignore the returned error as we cannot do anything about it anyway
		//nolint:errcheck,revive
		s.TCPlistener.Close()
		// Close all open TCP connections
		//  - get all conns from the s.conns map and put into slice
		//  - this is so the forget() function doesnt conflict with looping
		//    over the s.conns map
		var conns []*net.TCPConn
		s.cleanup.Lock()
		for _, conn := range s.conns {
			conns = append(conns, conn)
		}
		s.cleanup.Unlock()
		for _, conn := range conns {
			// Ignore the returned error as we cannot do anything about it anyway
			//nolint:errcheck,revive
			conn.Close()
		}
	}
	s.Unlock()

	s.wg.Wait()

	s.Lock()
	close(s.in)
	l.Infof("Stopped listener service on %q", s.ServiceAddress)
	s.Unlock()
}

// IsUDP returns true if the protocol is UDP, false otherwise.
func (s *input) isUDP() bool {
	return strings.HasPrefix(s.Protocol, "udp")
}

func (s *input) expireCachedMetrics() {
	// If Max TTL wasn't configured, skip expiration.
	if s.MaxTTL.Duration == 0 {
		return
	}

	now := time.Now()

	for key, cached := range s.gauges {
		if now.After(cached.expiresAt) {
			delete(s.gauges, key)
		}
	}

	for key, cached := range s.sets {
		if now.After(cached.expiresAt) {
			delete(s.sets, key)
		}
	}

	for key, cached := range s.timings {
		if now.After(cached.expiresAt) {
			delete(s.timings, key)
		}
	}

	for key, cached := range s.counters {
		if now.After(cached.expiresAt) {
			delete(s.counters, key)
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
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
