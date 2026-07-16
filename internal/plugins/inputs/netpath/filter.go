// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type compiledFilter struct {
	name             string
	sourceCIDRs      []*net.IPNet
	destCIDRs        []*net.IPNet
	sourceHosts      map[string]struct{}
	destHosts        map[string]struct{}
	namespaces       map[string]struct{}
	sourceServices   map[string]struct{}
	sourceProcesses  map[string]struct{}
	sourceContainers map[string]struct{}
	origins          map[string]struct{}
	protocols        map[string]struct{}
	ports            map[uint16]struct{}
	hasCriteria      bool
}

func compileFilters(filters []FilterConfig) []compiledFilter {
	out := make([]compiledFilter, 0, len(filters))
	for index, cfg := range filters {
		f := compiledFilter{
			name:             strings.TrimSpace(cfg.Name),
			sourceHosts:      hostnameSet(cfg.SourceHosts),
			destHosts:        hostnameSet(cfg.DestHosts),
			namespaces:       stringSet(cfg.Namespaces, false),
			sourceServices:   stringSet(cfg.SourceServices, false),
			sourceProcesses:  stringSet(cfg.SourceProcesses, false),
			sourceContainers: stringSet(cfg.SourceContainers, false),
			origins:          stringSet(cfg.Origins, false),
			protocols:        stringSet(cfg.Protocols, true),
			ports:            portSet(cfg.Ports),
		}
		var err error
		f.sourceCIDRs, err = parseCIDRs(cfg.SourceCIDRs)
		if err != nil {
			l.Errorf("disable netpath filter %q: invalid source_cidrs: %s", filterName(f.name, index), err.Error())
			continue
		}
		f.destCIDRs, err = parseCIDRs(cfg.DestCIDRs)
		if err != nil {
			l.Errorf("disable netpath filter %q: invalid dest_cidrs: %s", filterName(f.name, index), err.Error())
			continue
		}
		f.hasCriteria = len(f.sourceCIDRs) > 0 ||
			len(f.destCIDRs) > 0 ||
			len(f.sourceHosts) > 0 ||
			len(f.destHosts) > 0 ||
			len(f.namespaces) > 0 ||
			len(f.sourceServices) > 0 ||
			len(f.sourceProcesses) > 0 ||
			len(f.sourceContainers) > 0 ||
			len(f.origins) > 0 ||
			len(f.protocols) > 0 ||
			len(f.ports) > 0
		if f.hasCriteria {
			out = append(out, f)
		}
	}
	return out
}

func filterName(name string, index int) string {
	if name != "" {
		return name
	}
	return fmt.Sprintf("#%d", index+1)
}

func candidateDropReason(req candidateRequest, spec candidateSpec, cfg DynamicConfig) string {
	target, err := normalizeCandidateTarget(spec)
	if err != nil {
		// Let task construction preserve the specific validation error and its
		// API drop reason instead of allowing a filter to mask it.
		return ""
	}
	if target.hostname == "" && target.targetIP != "" && !cfg.MonitorIPWithoutDomain {
		return "ip_without_domain"
	}

	for _, f := range cfg.compiledFilters {
		if f.match(req, spec, target, cfg.Protocol) {
			return f.dropReason()
		}
	}
	return ""
}

func (f compiledFilter) match(req candidateRequest, spec candidateSpec, target candidateTarget, configProtocol string) bool {
	return f.matchValues(newCandidateFilterValues(req, spec, target, configProtocol))
}

func newCandidateFilterValues(
	req candidateRequest,
	spec candidateSpec,
	target candidateTarget,
	configProtocol string,
) candidateFilterValues {
	source := spec.Source
	return candidateFilterValues{
		sourceHost:      valueOrDefault(source.Hostname, req.Host),
		destHostname:    target.hostname,
		destTargetIP:    target.targetIP,
		destResolvedIP:  target.targetIP,
		namespace:       valueOrDefault(spec.Namespace, source.NetNS),
		sourceService:   source.ServiceName,
		sourceProcess:   source.ProcessName,
		sourceContainer: valueOrDefault(source.ContainerID, spec.SourceContainerID),
		origin:          valueOrDefault(spec.Origin, req.Source),
		protocol:        effectiveDynamicProtocol(spec.Protocol, configProtocol, spec.Port),
		port:            spec.Port,
		sourceIP:        source.IP,
	}
}

type candidateFilterValues struct {
	sourceHost      string
	destHostname    string
	destTargetIP    string
	destResolvedIP  string
	namespace       string
	sourceService   string
	sourceProcess   string
	sourceContainer string
	origin          string
	protocol        string
	port            uint16
	sourceIP        string
}

type targetResolver func(host string, timeout time.Duration) (net.IP, error)

type contextTargetResolver func(ctx context.Context, host string, timeout time.Duration) (net.IP, error)

func (f compiledFilter) matchValues(values candidateFilterValues) bool {
	if !matchHostnameSet(f.sourceHosts, values.sourceHost) {
		return false
	}
	if !matchHostnameSet(f.destHosts, values.destHostname, values.destTargetIP, values.destResolvedIP) {
		return false
	}
	if !matchSet(f.namespaces, values.namespace) {
		return false
	}
	if !matchSet(f.sourceServices, values.sourceService) {
		return false
	}
	if !matchSet(f.sourceProcesses, values.sourceProcess) {
		return false
	}
	if !matchSet(f.sourceContainers, values.sourceContainer) {
		return false
	}
	if !matchSet(f.origins, values.origin) {
		return false
	}
	if !matchSet(f.protocols, values.protocol) {
		return false
	}
	if !matchPortSet(f.ports, values.port) {
		return false
	}
	if !matchCIDRs(f.sourceCIDRs, values.sourceIP) {
		return false
	}
	if !matchCIDRs(f.destCIDRs, values.destResolvedIP) {
		return false
	}
	return true
}

func (f compiledFilter) dropReason() string {
	if f.name != "" {
		return "filtered:" + f.name
	}
	return "filtered"
}

func resolvedDestinationDropReason(t task, destinationIP string) string {
	values := t.filterValues
	values.destResolvedIP = destinationIP
	for _, f := range t.resolvedFilters {
		if len(f.destCIDRs) == 0 && len(f.destHosts) == 0 {
			continue
		}
		if f.matchValues(values) {
			return f.dropReason()
		}
	}
	return ""
}

func resolveTaskTarget(t task, host string, timeout time.Duration) (net.IP, error) {
	ip, err := resolveIPv4Fn(host, timeout)
	if err != nil {
		return nil, err
	}
	if reason := resolvedDestinationDropReason(t, ip.String()); reason != "" {
		return nil, fmt.Errorf("resolved target %q is blocked by %s", host, reason)
	}
	return ip, nil
}

func resolveTaskTargetContext(ctx context.Context, t task, host string, timeout time.Duration) (net.IP, error) {
	ip, err := resolveIPv4Context(ctx, host, timeout)
	if err != nil {
		return nil, err
	}
	if reason := resolvedDestinationDropReason(t, ip.String()); reason != "" {
		return nil, fmt.Errorf("resolved target %q is blocked by %s", host, reason)
	}
	return ip, nil
}

func resolverForTask(t task) targetResolver {
	return func(host string, timeout time.Duration) (net.IP, error) {
		return resolveTaskTarget(t, host, timeout)
	}
}

func contextResolverForTask(t task) contextTargetResolver {
	return func(ctx context.Context, host string, timeout time.Duration) (net.IP, error) {
		return resolveTaskTargetContext(ctx, t, host, timeout)
	}
}

func parseCIDRs(values []string) ([]*net.IPNet, error) {
	out := []*net.IPNet{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if ip := net.ParseIP(value); ip != nil {
			bits := 32
			if ip.To4() == nil {
				bits = 128
			}
			_, cidr, err := net.ParseCIDR(value + "/" + strconv.Itoa(bits))
			if err != nil {
				return nil, fmt.Errorf("parse %q: %w", value, err)
			}
			out = append(out, cidr)
			continue
		}
		_, cidr, err := net.ParseCIDR(value)
		if err != nil {
			return nil, fmt.Errorf("parse %q: %w", value, err)
		}
		out = append(out, cidr)
	}
	return out, nil
}

func stringSet(values []string, lower bool) map[string]struct{} {
	out := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if lower {
			value = strings.ToLower(value)
		}
		out[value] = struct{}{}
	}
	return out
}

func hostnameSet(values []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, value := range values {
		if value = normalizeHostname(value); value != "" {
			out[value] = struct{}{}
		}
	}
	return out
}

func normalizeHostname(value string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ".")
}

func portSet(values []uint16) map[uint16]struct{} {
	out := map[uint16]struct{}{}
	for _, value := range values {
		if value > 0 {
			out[value] = struct{}{}
		}
	}
	return out
}

func matchSet(set map[string]struct{}, values ...string) bool {
	if len(set) == 0 {
		return true
	}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := set[value]; ok {
			return true
		}
		if _, ok := set[strings.ToLower(value)]; ok {
			return true
		}
	}
	return false
}

func matchHostnameSet(set map[string]struct{}, values ...string) bool {
	if len(set) == 0 {
		return true
	}
	for _, value := range values {
		if value = normalizeHostname(value); value == "" {
			continue
		}
		if _, ok := set[value]; ok {
			return true
		}
	}
	return false
}

func matchPortSet(set map[uint16]struct{}, port uint16) bool {
	if len(set) == 0 {
		return true
	}
	_, ok := set[port]
	return ok
}

func matchCIDRs(cidrs []*net.IPNet, value string) bool {
	if len(cidrs) == 0 {
		return true
	}
	ip := net.ParseIP(strings.TrimSpace(value))
	if ip == nil {
		return false
	}
	for _, cidr := range cidrs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}
