// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package traceroute discovers IPv4 network paths with ICMP, UDP, or TCP SYN probes.
package traceroute

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"time"
)

const (
	// MaxHops is the maximum hop count accepted for ICMP and TCP traces.
	MaxHops = 60
	// MaxUDPHops is the maximum hop count accepted for UDP traces.
	MaxUDPHops = 255
	// MaxAttempts is the maximum number of probes sent for each hop.
	MaxAttempts = 10
	// MaxUDPProbePayload keeps IPv4 UDP probes below common path MTUs.
	MaxUDPProbePayload = 1200
	// MaxTimeout is the maximum wait time for one probe.
	MaxTimeout = 30 * time.Second
)

const defaultTimeout = 500 * time.Millisecond

// Protocol identifies the packet type used to discover a path.
type Protocol string

const (
	ProtocolICMP Protocol = "icmp"
	ProtocolUDP  Protocol = "udp"
	ProtocolTCP  Protocol = "tcp"
)

// Options controls a single traceroute run.
type Options struct {
	Protocol Protocol
	Port     uint16
	MaxTTL   int
	Attempts int
	Timeout  time.Duration
}

// RouteItem is one probe response at a hop. A timed-out probe uses IP "*".
type RouteItem struct {
	IP           string  `json:"ip"`
	ResponseTime float64 `json:"response_time"`
}

// Route summarizes all probe responses at one hop. Durations are microseconds.
type Route struct {
	Total   int          `json:"total"`
	Failed  int          `json:"failed"`
	Loss    float64      `json:"loss"`
	AvgCost float64      `json:"avg_cost"`
	MinCost float64      `json:"min_cost"`
	MaxCost float64      `json:"max_cost"`
	StdCost float64      `json:"std_cost"`
	Items   []*RouteItem `json:"items"`
}

// Result contains one route summary per TTL and whether the destination replied.
type Result struct {
	Routes  []*Route
	Reached bool
}

// ProbeResult describes one protocol probe.
type ProbeResult struct {
	IP       string
	RTT      time.Duration
	Reached  bool
	TimedOut bool
	Sent     bool
}

type probeReply struct {
	ip       net.IP
	rtt      time.Duration
	reached  bool
	timedOut bool
	sent     bool
}

type options struct {
	protocol Protocol
	port     uint16
	maxTTL   int
	attempts int
	timeout  time.Duration
}

// Trace discovers the path to an IPv4 destination.
func Trace(ctx context.Context, target net.IP, opts Options) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	target = target.To4()
	if target == nil {
		return Result{}, errors.New("traceroute target must be IPv4")
	}
	if target.IsUnspecified() || target.IsMulticast() || target.Equal(net.IPv4bcast) {
		return Result{}, fmt.Errorf("traceroute target %s must be a unicast address", target.String())
	}
	cfg, err := normalizeOptions(opts)
	if err != nil {
		return Result{}, err
	}
	return trace(ctx, target, cfg)
}

func normalizeOptions(opts Options) (options, error) {
	cfg := options{
		protocol: opts.Protocol,
		port:     opts.Port,
		maxTTL:   opts.MaxTTL,
		attempts: opts.Attempts,
		timeout:  opts.Timeout,
	}
	if cfg.protocol == "" {
		cfg.protocol = ProtocolICMP
	}
	if cfg.protocol != ProtocolICMP && cfg.protocol != ProtocolUDP && cfg.protocol != ProtocolTCP {
		return options{}, fmt.Errorf("unsupported traceroute protocol %q", cfg.protocol)
	}
	if (cfg.protocol == ProtocolUDP || cfg.protocol == ProtocolTCP) && cfg.port == 0 {
		return options{}, fmt.Errorf("%s traceroute requires a destination port", cfg.protocol)
	}
	if cfg.maxTTL <= 0 {
		cfg.maxTTL = 30
	}
	maxTTL := MaxHops
	if cfg.protocol == ProtocolUDP {
		maxTTL = MaxUDPHops
	}
	if cfg.maxTTL > maxTTL {
		cfg.maxTTL = maxTTL
	}
	if cfg.attempts <= 0 {
		cfg.attempts = 1
	}
	if cfg.attempts > MaxAttempts {
		cfg.attempts = MaxAttempts
	}
	if cfg.protocol == ProtocolUDP && cfg.maxTTL*cfg.attempts > MaxUDPProbePayload {
		return options{}, fmt.Errorf("UDP traceroute max TTL times attempts must not exceed %d",
			MaxUDPProbePayload)
	}
	if cfg.timeout <= 0 {
		cfg.timeout = defaultTimeout
	}
	if cfg.timeout > MaxTimeout {
		cfg.timeout = MaxTimeout
	}
	return cfg, nil
}

func repliesToRoute(replies []probeReply) *Route {
	route := &Route{Total: len(replies), Items: make([]*RouteItem, 0, len(replies))}
	latencies := make([]float64, 0, len(replies))
	for _, reply := range replies {
		item := &RouteItem{IP: "*"}
		if reply.timedOut || reply.ip == nil {
			route.Failed++
		} else {
			item.IP = reply.ip.String()
			item.ResponseTime = float64(reply.rtt.Microseconds())
			latencies = append(latencies, item.ResponseTime)
		}
		route.Items = append(route.Items, item)
	}
	if route.Total > 0 {
		route.Loss = float64(route.Failed) * 100 / float64(route.Total)
	}
	if len(latencies) == 0 {
		return route
	}
	route.MinCost, route.MaxCost = latencies[0], latencies[0]
	for _, latency := range latencies {
		route.AvgCost += latency
		if latency < route.MinCost {
			route.MinCost = latency
		}
		if latency > route.MaxCost {
			route.MaxCost = latency
		}
	}
	route.AvgCost /= float64(len(latencies))
	if len(latencies) > 1 {
		for _, latency := range latencies {
			delta := latency - route.AvgCost
			route.StdCost += delta * delta
		}
		route.StdCost = math.Sqrt(route.StdCost / float64(len(latencies)-1))
	}
	return route
}
