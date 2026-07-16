// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	tr "github.com/GuanceCloud/cliutils/traceroute"
)

// tracerouteProbe owns NetPath orchestration and result formatting. Packet IO
// is implemented by cliutils/traceroute.
type tracerouteProbe struct {
	host            string
	port            uint16
	protocol        tr.Protocol
	timeout         time.Duration
	queries         int
	maxTTL          int
	tags            map[string]string
	resolver        targetResolver
	contextResolver contextTargetResolver

	destIP               string
	multipleDestinations bool
	payload              traceroutePayload
	reached              bool
	err                  error
}

var runTraceroute = tr.Trace
var resolveIPv4Fn = resolveIPv4

func newTracerouteProbe(t task, resolver targetResolver) *tracerouteProbe {
	return &tracerouteProbe{
		host:     t.Target,
		port:     t.Port,
		protocol: tr.Protocol(t.Protocol),
		timeout:  normalizeProbeTimeout(t.Timeout),
		queries:  normalizeTracerouteQueries(t.TracerouteQueries),
		maxTTL:   normalizeMaxTTL(t.MaxTTL, t.Protocol),
		tags:     map[string]string{},
		resolver: resolver,
	}
}

func (p *tracerouteProbe) Check() error {
	if p == nil || p.host == "" {
		return errors.New("netpath target missing host")
	}
	if (p.protocol == tr.ProtocolTCP || p.protocol == tr.ProtocolUDP) && p.port == 0 {
		return fmt.Errorf("%s netpath target missing port", p.protocol)
	}
	return nil
}

func (p *tracerouteProbe) Run() error {
	return p.RunContext(context.Background())
}

func (p *tracerouteProbe) RunContext(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	options := tr.Options{
		Protocol: p.protocol,
		Port:     p.port,
		MaxTTL:   p.maxTTL,
		Attempts: 1,
		Timeout:  p.timeout,
	}
	for runID := 1; runID <= p.queries; runID++ {
		if err := ctx.Err(); err != nil {
			p.err = err
			return err
		}
		ip, err := p.resolve(ctx)
		if err != nil {
			p.err = fmt.Errorf("resolve %s traceroute run %d target: %w", p.protocol, runID, err)
			return p.err
		}
		destinationIP := ip.String()
		p.recordDestination(destinationIP)

		result, err := runTraceroute(ctx, ip, options)
		if len(result.Routes) > 0 {
			run := routesToRun(result.Routes, runID)
			run.Destination = newTracerouteDestination(p.host, destinationIP, p.port)
			p.payload.Runs = append(p.payload.Runs, run)
			p.reached = p.reached || result.Reached
		}
		if err != nil {
			p.err = fmt.Errorf("%s traceroute run %d: %w", p.protocol, runID, err)
			return p.err
		}
		if len(result.Routes) == 0 {
			p.err = fmt.Errorf("%s traceroute run %d returned no hops", p.protocol, runID)
			return p.err
		}
	}
	p.payload.HopCount = buildHopCount(p.payload.Runs)
	return nil
}

func (p *tracerouteProbe) resolve(ctx context.Context) (net.IP, error) {
	if p.contextResolver != nil {
		return p.contextResolver(ctx, p.host, p.timeout)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	resolver := p.resolver
	if resolver == nil {
		resolver = resolveIPv4Fn
	}
	ip, err := resolver(p.host, p.timeout)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return ip, nil
}

func (p *tracerouteProbe) GetResults() (map[string]string, map[string]interface{}) {
	p.payload.HopCount = buildHopCount(p.payload.Runs)
	tags := cloneTags(p.tags)
	if p.destIP != "" {
		tags["probe_dest_ip"] = p.destIP
	}
	if p.multipleDestinations {
		tags["probe_dest_ip_multiple"] = "true"
	}
	tags["traceroute_protocol"] = string(p.protocol)

	fields := map[string]interface{}{"hops": 0, "traceroute": "[]"}
	if len(p.payload.Runs) > 0 {
		if data, err := json.Marshal(p.payload); err == nil {
			fields["traceroute"] = string(data)
			fields["hops"] = p.payload.HopCount.Max
		}
	}
	setTracerouteResultStatus(tags, fields, p.payload, p.reached, p.err)
	return tags, fields
}

func (p *tracerouteProbe) recordDestination(ip string) {
	if p.multipleDestinations {
		return
	}
	if p.destIP == "" {
		p.destIP = ip
		return
	}
	if p.destIP != ip {
		p.destIP = ""
		p.multipleDestinations = true
	}
}

func resolveIPv4(host string, timeout time.Duration) (net.IP, error) {
	return resolveIPv4Context(context.Background(), host, timeout)
}

func resolveIPv4Context(ctx context.Context, host string, timeout time.Duration) (net.IP, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			return validateProbeTargetIPv4(host, ip4)
		}
		return nil, fmt.Errorf("netpath target %q is not IPv4", host)
	}

	timeout = normalizeProbeTimeout(timeout)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		if ip4 := addr.IP.To4(); ip4 != nil {
			if validated, err := validateProbeTargetIPv4(host, ip4); err == nil {
				return validated, nil
			}
		}
	}
	return nil, fmt.Errorf("netpath target %q has no usable unicast IPv4 address", host)
}

func validateProbeTargetIPv4(host string, ip net.IP) (net.IP, error) {
	if ip.IsUnspecified() || ip.IsMulticast() || ip.Equal(net.IPv4bcast) {
		return nil, fmt.Errorf("netpath target %q resolved to non-unicast address %s", host, ip.String())
	}
	return ip, nil
}
