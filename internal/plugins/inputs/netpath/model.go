// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

import (
	"crypto/sha1" //nolint:gosec
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

const (
	protocolAuto = "auto"
	protocolTCP  = "tcp"
	protocolUDP  = "udp"
	protocolICMP = "icmp"

	sourceLocal   = "local"
	sourceDynamic = "dynamic"

	runTypeScheduled = "scheduled"
	runTypeDynamic   = "dynamic"
)

type TargetConfig struct {
	Name              string            `toml:"name" json:"name"`
	Target            string            `toml:"target" json:"target"`
	IP                string            `toml:"ip" json:"ip"`
	Port              uint16            `toml:"port" json:"port"`
	Protocol          string            `toml:"protocol" json:"protocol"`
	Interval          *datakit.Duration `toml:"interval" json:"interval"`
	Timeout           *datakit.Duration `toml:"timeout" json:"timeout"`
	MaxTTL            int               `toml:"max_ttl" json:"max_ttl"`
	TracerouteQueries int               `toml:"traceroute_queries" json:"traceroute_queries"`
	E2EQueries        int               `toml:"e2e_queries" json:"e2e_queries"`
	Tags              map[string]string `toml:"tags" json:"tags"`
}

type DynamicConfig struct {
	Enabled                bool              `toml:"enabled" json:"enabled"`
	Token                  string            `toml:"token" json:"token"`
	Protocol               string            `toml:"protocol" json:"protocol"`
	Tags                   map[string]string `toml:"tags" json:"tags"`
	ContextsLimit          int               `toml:"contexts_limit" json:"contexts_limit"`
	ContextsBytesLimit     int64             `toml:"contexts_bytes_limit" json:"contexts_bytes_limit"`
	TTL                    *datakit.Duration `toml:"ttl" json:"ttl"`
	Interval               *datakit.Duration `toml:"interval" json:"interval"`
	FlushInterval          *datakit.Duration `toml:"flush_interval" json:"flush_interval"`
	MaxPerMinute           int               `toml:"max_per_minute" json:"max_per_minute"`
	Workers                int               `toml:"workers" json:"workers"`
	Timeout                *datakit.Duration `toml:"timeout" json:"timeout"`
	MaxTTL                 int               `toml:"max_ttl" json:"max_ttl"`
	Queries                int               `toml:"traceroute_queries" json:"traceroute_queries"`
	E2EQueries             int               `toml:"e2e_queries" json:"e2e_queries"`
	InputQueue             int               `toml:"input_queue" json:"input_queue"`
	ProcessQueue           int               `toml:"process_queue" json:"process_queue"`
	MonitorIPWithoutDomain bool              `toml:"monitor_ip_without_domain" json:"monitor_ip_without_domain"`
	MaxTestsPerRequest     int               `toml:"max_tests_per_request" json:"max_tests_per_request"`
	MaxBodyBytes           int64             `toml:"max_body_bytes" json:"max_body_bytes"`
	Filters                []FilterConfig    `toml:"filters" json:"filters"`

	compiledFilters []compiledFilter
}

type ReverseDNSConfig struct {
	Enabled   bool              `toml:"enabled" json:"enabled"`
	Timeout   *datakit.Duration `toml:"timeout" json:"timeout"`
	CacheTTL  *datakit.Duration `toml:"cache_ttl" json:"cache_ttl"`
	CacheSize int               `toml:"cache_size" json:"cache_size"`
}

type FilterConfig struct {
	Name             string   `toml:"name" json:"name"`
	SourceCIDRs      []string `toml:"source_cidrs" json:"source_cidrs"`
	DestCIDRs        []string `toml:"dest_cidrs" json:"dest_cidrs"`
	SourceHosts      []string `toml:"source_hosts" json:"source_hosts"`
	DestHosts        []string `toml:"dest_hosts" json:"dest_hosts"`
	Namespaces       []string `toml:"namespaces" json:"namespaces"`
	SourceServices   []string `toml:"source_services" json:"source_services"`
	SourceProcesses  []string `toml:"source_processes" json:"source_processes"`
	SourceContainers []string `toml:"source_containers" json:"source_containers"`
	Origins          []string `toml:"origins" json:"origins"`
	Protocols        []string `toml:"protocols" json:"protocols"`
	Ports            []uint16 `toml:"ports" json:"ports"`
}

type candidateRequest struct {
	Source   string            `json:"source"`
	Host     string            `json:"host,omitempty"`
	AgentID  string            `json:"agent_id,omitempty"`
	Sequence uint64            `json:"sequence,omitempty"`
	SentAt   int64             `json:"sent_at,omitempty"`
	Tags     map[string]string `json:"tags,omitempty"`
	Tests    []candidateSpec   `json:"tests"`
}

type candidateSpec struct {
	Hostname          string            `json:"hostname"`
	TargetIP          string            `json:"target_ip,omitempty"`
	IP                string            `json:"ip,omitempty"`
	Port              uint16            `json:"port,omitempty"`
	DstIP             string            `json:"dst_ip,omitempty"`
	DstPort           uint16            `json:"dst_port,omitempty"`
	Protocol          string            `json:"protocol"`
	Origin            string            `json:"origin"`
	Namespace         string            `json:"namespace,omitempty"`
	SourceContainerID string            `json:"source_container_id,omitempty"`
	Source            candidateSource   `json:"source,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
}

type candidateSource struct {
	Hostname    string `json:"hostname,omitempty"`
	IP          string `json:"ip,omitempty"`
	Port        uint16 `json:"port,omitempty"`
	NetNS       string `json:"netns,omitempty"`
	ContainerID string `json:"container_id,omitempty"`
	PID         uint32 `json:"pid,omitempty"`
	ProcessName string `json:"process_name,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
}

type candidateResponse struct {
	Accepted    int            `json:"accepted"`
	Dropped     int            `json:"dropped"`
	QueueSize   int            `json:"queue_size"`
	DropReasons map[string]int `json:"drop_reasons,omitempty"`
}

type task struct {
	ID                string
	Name              string
	Source            string
	Origin            string
	RunType           string
	Target            string
	Hostname          string
	TargetIP          string
	Port              uint16
	Protocol          string
	Namespace         string
	SourceContainerID string
	SourceHost        string
	SourceIP          string
	SourcePort        uint16
	SourcePID         uint32
	SourceProcess     string
	SourceService     string
	DstIP             string
	DstPort           uint16
	NetNS             string
	// ConfigTags and RequestTags are immutable maps shared by queued dynamic tasks.
	ConfigTags        map[string]string
	RequestTags       map[string]string
	Tags              map[string]string
	ScheduleKey       string
	Interval          time.Duration
	TTL               time.Duration
	Timeout           time.Duration
	MaxTTL            int
	TracerouteQueries int
	E2EQueries        int
	ScheduledAt       time.Time
	resolvedFilters   []compiledFilter
	filterValues      candidateFilterValues
}

var errIPv6TargetUnsupported = errors.New("netpath IPv6 targets are unsupported; only IPv4 targets are supported")

func taskFromTargetConfig(target TargetConfig, defaults *Input) (task, error) {
	if err := rejectIPv6TargetLiteral(target.Target); err != nil {
		return task{}, err
	}
	if err := rejectIPv6TargetLiteral(target.IP); err != nil {
		return task{}, err
	}

	dest := strings.TrimSpace(valueOrDefault(target.Target, target.IP))
	if dest == "" {
		return task{}, errors.New("missing target")
	}
	if net.ParseIP(dest) == nil && strings.TrimSpace(target.Target) == "" {
		return task{}, fmt.Errorf("invalid target %q", dest)
	}

	protocol := normalizeProtocol(valueOrDefault(target.Protocol, defaults.Protocol))
	if protocol == protocolAuto {
		if target.Port > 0 {
			protocol = protocolTCP
		} else {
			protocol = protocolICMP
		}
	}
	if err := validateProtocol(protocol); err != nil {
		return task{}, err
	}
	if err := validateProtocolForPlatform(protocol); err != nil {
		return task{}, err
	}
	if (protocol == protocolTCP || protocol == protocolUDP) && target.Port == 0 {
		return task{}, fmt.Errorf("%s netpath target missing port", protocol)
	}

	name := valueOrDefault(target.Name, dest)
	hostname := strings.TrimSpace(target.Target)
	if net.ParseIP(hostname) != nil {
		hostname = ""
	}
	interval := defaultStaticInterval
	if defaults.Interval != nil && defaults.Interval.Duration > 0 {
		interval = defaults.Interval.Duration
	}
	if target.Interval != nil && target.Interval.Duration > 0 {
		interval = target.Interval.Duration
	}
	timeout := defaultTimeout
	if defaults.Timeout != nil && defaults.Timeout.Duration > 0 {
		timeout = defaults.Timeout.Duration
	}
	if target.Timeout != nil && target.Timeout.Duration > 0 {
		timeout = target.Timeout.Duration
	}
	timeout = normalizeProbeTimeout(timeout)
	maxTTL := normalizeMaxTTL(firstPositive(target.MaxTTL, defaults.MaxTTL, defaultMaxTTL), protocol)
	queries := normalizeTracerouteQueries(firstPositive(target.TracerouteQueries, defaults.TracerouteQueries, defaultTracerouteQueries))
	e2eQueries := normalizeE2EQueries(firstPositive(target.E2EQueries, defaults.E2EQueries, defaultE2EQueries))

	tags := cloneTags(defaults.Tags)
	mergeTags(tags, target.Tags)
	scheduleKey := makeScheduleKey(sourceLocal, name, dest, strconv.Itoa(int(target.Port)), protocol)

	return task{
		ID:                "local-" + shortHash(scheduleKey),
		Name:              name,
		Source:            sourceLocal,
		Origin:            "config",
		RunType:           runTypeScheduled,
		Target:            dest,
		Hostname:          hostname,
		TargetIP:          strings.TrimSpace(target.IP),
		Port:              target.Port,
		DstIP:             strings.TrimSpace(target.IP),
		DstPort:           target.Port,
		Protocol:          protocol,
		Tags:              tags,
		ScheduleKey:       scheduleKey,
		Interval:          interval,
		Timeout:           timeout,
		MaxTTL:            maxTTL,
		TracerouteQueries: queries,
		E2EQueries:        e2eQueries,
	}, nil
}

func taskFromCandidate(req candidateRequest, spec candidateSpec, cfg DynamicConfig) (task, error) {
	target, err := normalizeCandidateTarget(spec)
	if err != nil {
		return task{}, err
	}
	targetIP := target.targetIP
	hostname := target.hostname

	source := spec.Source
	if source.Hostname == "" {
		source.Hostname = req.Host
	}
	if source.ContainerID == "" {
		source.ContainerID = spec.SourceContainerID
	}
	namespace := strings.TrimSpace(spec.Namespace)
	sourceNetNS := strings.TrimSpace(source.NetNS)

	protocol := effectiveDynamicProtocol(spec.Protocol, cfg.Protocol, spec.Port)
	if err := validateProtocol(protocol); err != nil {
		return task{}, err
	}
	if err := validateProtocolForPlatform(protocol); err != nil {
		return task{}, err
	}
	if (protocol == protocolTCP || protocol == protocolUDP) && spec.Port == 0 {
		return task{}, fmt.Errorf("%s netpath candidate missing port", protocol)
	}

	sourceKey := valueOrDefault(source.ServiceName, source.ProcessName)
	if sourceKey == "" {
		sourceKey = source.ContainerID
	}
	if sourceKey == "" {
		sourceKey = source.Hostname
	}
	displayHost := valueOrDefault(hostname, target.target)
	dstIP := valueOrDefault(spec.DstIP, targetIP)
	hasDestinationTranslation := candidateHasDestinationTranslation(spec, targetIP)
	probeTarget := displayHost
	if hasDestinationTranslation && targetIP != "" {
		probeTarget = targetIP
	}
	scheduleDstIP := ""
	scheduleDstPort := ""
	if hasDestinationTranslation {
		scheduleDstIP = dstIP
		scheduleDstPort = strconv.Itoa(int(spec.DstPort))
	}
	scheduleKey := makeScheduleKey(
		sourceDynamic,
		source.Hostname, sourceKey, namespace, sourceNetNS,
		displayHost, strconv.Itoa(int(spec.Port)), protocol,
		scheduleDstIP, scheduleDstPort,
	)

	t := task{
		ID:                "dynamic-" + shortHash(scheduleKey),
		Name:              "dynamic " + displayHost,
		Source:            sourceDynamic,
		Origin:            valueOrDefault(spec.Origin, req.Source),
		RunType:           runTypeDynamic,
		Target:            probeTarget,
		Hostname:          hostname,
		TargetIP:          targetIP,
		Port:              spec.Port,
		Protocol:          protocol,
		Namespace:         namespace,
		SourceContainerID: valueOrDefault(spec.SourceContainerID, source.ContainerID),
		SourceHost:        source.Hostname,
		SourceIP:          source.IP,
		SourcePort:        source.Port,
		SourcePID:         source.PID,
		SourceProcess:     source.ProcessName,
		SourceService:     source.ServiceName,
		DstIP:             dstIP,
		DstPort:           spec.DstPort,
		NetNS:             sourceNetNS,
		ConfigTags:        cfg.Tags,
		RequestTags:       req.Tags,
		Tags:              spec.Tags,
		ScheduleKey:       scheduleKey,
		Interval:          cfg.Interval.Duration,
		TTL:               cfg.TTL.Duration,
		Timeout:           normalizeProbeTimeout(cfg.Timeout.Duration),
		MaxTTL:            normalizeMaxTTL(cfg.MaxTTL, protocol),
		TracerouteQueries: cfg.Queries,
		E2EQueries:        cfg.E2EQueries,
		resolvedFilters:   cfg.compiledFilters,
		filterValues:      newCandidateFilterValues(req, spec, target, cfg.Protocol),
	}
	detachDynamicTaskStrings(&t)
	return t, nil
}

type candidateTarget struct {
	target   string
	hostname string
	targetIP string
}

func normalizeCandidateTarget(spec candidateSpec) (candidateTarget, error) {
	hostname := strings.TrimSpace(spec.Hostname)
	targetIP := strings.TrimSpace(valueOrDefault(spec.TargetIP, spec.IP))
	if err := rejectIPv6TargetLiteral(hostname); err != nil {
		return candidateTarget{}, err
	}
	if err := rejectIPv6TargetLiteral(targetIP); err != nil {
		return candidateTarget{}, err
	}
	if ip := net.ParseIP(hostname); ip != nil {
		hostname = ""
		if targetIP == "" {
			targetIP = ip.String()
		}
	}

	target := targetIP
	if target == "" {
		target = hostname
	}
	if target == "" {
		return candidateTarget{}, errors.New("missing target")
	}
	if target == targetIP && net.ParseIP(target) == nil {
		return candidateTarget{}, fmt.Errorf("invalid target %q", target)
	}
	return candidateTarget{target: target, hostname: hostname, targetIP: targetIP}, nil
}

func rejectIPv6TargetLiteral(value string) error {
	value = strings.TrimSpace(value)
	address := value
	if strings.HasPrefix(address, "[") && strings.HasSuffix(address, "]") {
		address = strings.TrimSpace(address[1 : len(address)-1])
	}
	if ip, err := netip.ParseAddr(address); err == nil && !ip.Is4() {
		return fmt.Errorf("%w: %q", errIPv6TargetUnsupported, value)
	}
	return nil
}

func candidateHasDestinationTranslation(spec candidateSpec, targetIP string) bool {
	if targetIP == "" {
		return false
	}
	destinationIP := strings.TrimSpace(spec.DstIP)
	if destinationIP != "" && !sameIP(destinationIP, targetIP) {
		return true
	}
	return spec.DstPort > 0 && spec.Port != spec.DstPort
}

func sameIP(left, right string) bool {
	leftIP := net.ParseIP(strings.TrimSpace(left))
	rightIP := net.ParseIP(strings.TrimSpace(right))
	if leftIP != nil && rightIP != nil {
		return leftIP.Equal(rightIP)
	}
	return strings.EqualFold(strings.TrimSpace(left), strings.TrimSpace(right))
}

func normalizeProtocol(protocol string) string {
	protocol = strings.ToLower(strings.TrimSpace(protocol))
	if protocol == "" {
		return protocolAuto
	}
	return protocol
}

func validateProtocol(protocol string) error {
	switch protocol {
	case protocolTCP, protocolUDP, protocolICMP:
		return nil
	default:
		return fmt.Errorf("unsupported netpath protocol %q", protocol)
	}
}

func effectiveDynamicProtocol(targetProtocol, configProtocol string, port uint16) string {
	configProtocol = normalizeProtocol(configProtocol)
	if configProtocol == protocolTCP || configProtocol == protocolUDP || configProtocol == protocolICMP {
		return configProtocol
	}
	targetProtocol = normalizeProtocol(targetProtocol)
	switch targetProtocol {
	case protocolTCP, protocolUDP, protocolICMP:
		return targetProtocol
	}
	if port > 0 {
		return protocolTCP
	}
	return protocolICMP
}

func valueOrDefault(v, fallback string) string {
	if strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return strings.TrimSpace(fallback)
}

func firstPositive(values ...int) int {
	for _, v := range values {
		if v > 0 {
			return v
		}
	}
	return 0
}

func shortHash(v string) string {
	sum := sha1.Sum([]byte(v)) //nolint:gosec
	return hex.EncodeToString(sum[:])[:12]
}

func makeScheduleKey(parts ...string) string {
	var key strings.Builder
	for _, part := range parts {
		key.WriteString(strconv.Itoa(len(part)))
		key.WriteByte(':')
		key.WriteString(part)
	}
	return key.String()
}

// detachDynamicTaskStrings prevents normalized substrings from retaining the
// larger JSON input strings they were sliced from. Reuse the detached task
// strings for filter values whenever both views have the same value.
func detachDynamicTaskStrings(t *task) {
	if t == nil {
		return
	}
	t.Origin = strings.Clone(t.Origin)
	t.Hostname = strings.Clone(t.Hostname)
	t.TargetIP = strings.Clone(t.TargetIP)
	switch t.Target {
	case t.TargetIP:
		t.Target = t.TargetIP
	case t.Hostname:
		t.Target = t.Hostname
	default:
		t.Target = strings.Clone(t.Target)
	}
	t.Protocol = strings.Clone(t.Protocol)
	t.Namespace = strings.Clone(t.Namespace)
	t.SourceContainerID = strings.Clone(t.SourceContainerID)
	if t.DstIP == t.TargetIP {
		t.DstIP = t.TargetIP
	} else {
		t.DstIP = strings.Clone(t.DstIP)
	}
	t.NetNS = strings.Clone(t.NetNS)
	t.ScheduleKey = strings.Clone(t.ScheduleKey)

	values := &t.filterValues
	if values.sourceHost == t.SourceHost {
		values.sourceHost = t.SourceHost
	} else {
		values.sourceHost = strings.Clone(values.sourceHost)
	}
	values.destHostname = t.Hostname
	values.destTargetIP = t.TargetIP
	values.destResolvedIP = t.TargetIP
	switch values.namespace {
	case t.Namespace:
		values.namespace = t.Namespace
	case t.NetNS:
		values.namespace = t.NetNS
	default:
		values.namespace = strings.Clone(values.namespace)
	}
	values.sourceService = t.SourceService
	values.sourceProcess = t.SourceProcess
	if values.sourceContainer == t.SourceContainerID {
		values.sourceContainer = t.SourceContainerID
	} else {
		values.sourceContainer = strings.Clone(values.sourceContainer)
	}
	values.origin = t.Origin
	values.protocol = t.Protocol
	values.sourceIP = t.SourceIP
}

func cloneTags(in map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range in {
		out[k] = v
	}
	return out
}

func mergeTags(dst, src map[string]string) {
	for k, v := range src {
		dst[k] = v
	}
}

func firstPositivePort(values ...uint16) uint16 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}
