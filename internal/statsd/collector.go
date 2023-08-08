// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package statsd serve a UDP/TCP(not used) server to handle statsd metrics.
package statsd

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/influxdata/telegraf/plugins/parsers/graphite"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
)

const (
	// UDPMaxPacketSize is the UDP packet limit, see
	// https://en.wikipedia.org/wiki/User_Datagram_Protocol#Packet_structure
	UDPMaxPacketSize int = 64 * 1024
	defaultFieldName     = "value"
	defaultSeparator     = "_"
	parserGoRoutines     = 5
)

var g = goroutine.NewGroup(goroutine.Option{Name: "inputs_statsd"})

type Collector struct {
	opts *option
	mmap map[string]string

	sync.Mutex
	// Lock for preventing a data race during resource cleanup
	cleanup sync.Mutex
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

	graphiteParser *graphite.GraphiteParser

	acc *accumulator

	// A pool of byte slices to handle parsingPromOption
	bufPool sync.Pool

	opt point.Option
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

func (col *Collector) setupMmap() {
	col.mmap = map[string]string{}

	for _, mm := range col.opts.metricMapping {
		arr := strings.SplitN(mm, ":", 2)
		if len(arr) != 2 {
			col.opts.l.Warnf("invalid MetricMapping: %s, ignored", mm)
			continue
		}

		col.mmap[arr[0]] = arr[1]
	}
}

func (col *Collector) Exit() {
	col.stop()
}

// aggregate takes in a metric. It then
// aggregates and caches the current value(s). It does not deal with the
// Delete* options, because those are dealt with in the Gather function.
func (col *Collector) aggregate(m metric) {
	col.Lock()
	defer col.Unlock()

	switch m.mtype {
	case "d":
		if col.opts.dataDogExtensions && col.opts.dataDogDistributions {
			cached := cacheddistributions{
				name:  m.name,
				value: m.floatvalue,
				tags:  m.tags,
			}
			col.distributions = append(col.distributions, cached)
		}
	case "ms", "h":
		// Check if the measurement exists
		cached, ok := col.timings[m.hash]
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
				PercLimit: col.opts.percentileLimit,
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
		cached.expiresAt = time.Now().Add(col.opts.maxTTL)
		col.timings[m.hash] = cached
	case "c":
		// check if the measurement exists
		cached, ok := col.counters[m.hash]
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
		cached.expiresAt = time.Now().Add(col.opts.maxTTL)
		col.counters[m.hash] = cached
	case "g":
		// check if the measurement exists
		cached, ok := col.gauges[m.hash]
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

		cached.expiresAt = time.Now().Add(col.opts.maxTTL)
		col.gauges[m.hash] = cached
	case "s":
		// check if the measurement exists
		cached, ok := col.sets[m.hash]
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
		cached.expiresAt = time.Now().Add(col.opts.maxTTL)
		col.sets[m.hash] = cached
	}
}

func (col *Collector) stop() {
	col.Lock()
	col.opts.l.Infof("Stopping the statsd service")
	close(col.done)
	if col.isUDP() && col.UDPlistener != nil {
		// Ignore the returned error as we cannot do anything about it anyway
		//nolint:errcheck,revive
		if err := col.UDPlistener.Close(); err != nil {
			col.opts.l.Warnf("Close: %s, ignored", err)
		}
	} else if col.TCPlistener != nil {
		// Ignore the returned error as we cannot do anything about it anyway
		//nolint:errcheck,revive
		if err := col.TCPlistener.Close(); err != nil {
			col.opts.l.Warnf("Close: %s, ignored", err)
		}

		// Close all open TCP connections
		//  - get all conns from the s.conns map and put into slice
		//  - this is so the forget() function doesnt conflict with looping
		//    over the s.conns map
		var conns []*net.TCPConn
		col.cleanup.Lock()
		for _, conn := range col.conns {
			conns = append(conns, conn)
		}
		col.cleanup.Unlock()
		for _, conn := range conns {
			// Ignore the returned error as we cannot do anything about it anyway
			//nolint:errcheck,revive
			if err := conn.Close(); err != nil {
				col.opts.l.Warnf("Close: %s, ignored", err)
			}
		}
	}

	col.Unlock()

	_ = g.Wait()

	col.Lock()
	close(col.in)
	col.opts.l.Infof("Stopped listener service on %s", col.opts.serviceAddress)
	col.Unlock()
}

// IsUDP returns true if the protocol is UDP, false otherwise.
func (col *Collector) isUDP() bool {
	return strings.HasPrefix(col.opts.protocol, "udp")
}

func NewCollector(udplistener *net.UDPConn, tcplistener *net.TCPListener, collectorOpts ...CollectorOption) (*Collector, error) {
	opt := option{}
	for idx := range collectorOpts {
		if collectorOpts[idx] != nil {
			collectorOpts[idx](&opt)
		}
	}

	if opt.l == nil {
		opt.l = logger.DefaultSLogger("statsd")
	}

	col := &Collector{
		opts:        &opt,
		UDPlistener: udplistener,
		TCPlistener: tcplistener,
		mmap:        map[string]string{},
	}

	// no election.
	col.opt = point.WithExtraTags(dkpt.GlobalHostTags())

	// Make data structures
	col.gauges = make(map[string]cachedgauge)
	col.counters = make(map[string]cachedcounter)
	col.sets = make(map[string]cachedset)
	col.timings = make(map[string]cachedtimings)
	col.distributions = make([]cacheddistributions, 0)

	col.Lock()
	defer col.Unlock()

	col.in = make(chan job, col.opts.allowedPendingMessages)
	col.done = make(chan struct{})
	col.accept = make(chan bool, col.opts.maxTCPConnections)
	col.conns = make(map[string]*net.TCPConn)
	col.bufPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	for i := 0; i < col.opts.maxTCPConnections; i++ {
		col.accept <- true
	}

	if col.opts.convertNames {
		col.opts.l.Warn("'convert_names' config option is deprecated, please use 'metric_separator' instead")
	}

	if col.opts.metricSeparator == "" {
		col.opts.metricSeparator = defaultSeparator
	}

	if len(col.opts.metricMapping) > 0 {
		col.setupMmap()
	}

	if col.isUDP() {
		if err := col.setupUDPServer(); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("TCP not supported")
		// TODO: not testing
		// s.setupTCPServer()
	}

	col.acc = &accumulator{
		ref: col,
		l:   opt.l,
	}

	col.opts.l.Infof("starting %d parser worker...", parserGoRoutines)
	for i := 1; i <= parserGoRoutines; i++ {
		// Start the line parser
		func(idx int) {
			g.Go(func(ctx context.Context) error {
				col.parser(idx)
				return nil
			})
		}(i)
	}

	return col, nil
}
