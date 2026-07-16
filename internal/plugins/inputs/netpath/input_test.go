// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	tr "github.com/GuanceCloud/cliutils/traceroute"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

type captureNetpathFeeder struct {
	category point.Category
	points   []*point.Point
	err      error
}

func (f *captureNetpathFeeder) Feed(category point.Category, points []*point.Point, _ ...dkio.FeedOption) error {
	f.category = category
	f.points = append(f.points[:0], points...)
	return f.err
}

func (f *captureNetpathFeeder) FeedLastError(string, ...metrics.LastErrorOption) {
}

func TestNormalizeDefaults(t *testing.T) {
	ipt := &Input{}
	ipt.normalize()

	require.NotNil(t, ipt.Interval)
	require.NotNil(t, ipt.Timeout)
	require.NotNil(t, ipt.Dynamic)
	require.NotNil(t, ipt.ReverseDNS)
	assert.Equal(t, protocolTCP, ipt.Protocol)
	assert.Equal(t, defaultStaticInterval, ipt.Interval.Duration)
	assert.Equal(t, defaultTimeout, ipt.Timeout.Duration)
	assert.Equal(t, defaultMaxTTL, ipt.MaxTTL)
	assert.Equal(t, defaultTracerouteQueries, ipt.TracerouteQueries)
	assert.Equal(t, defaultE2EQueries, ipt.E2EQueries)
	assert.Equal(t, protocolAuto, ipt.Dynamic.Protocol)
	assert.Equal(t, defaultE2EQueries, ipt.Dynamic.E2EQueries)
	assert.Equal(t, defaultDynamicInterval, ipt.Dynamic.Interval.Duration)
	assert.Equal(t, defaultDynamicTTL, ipt.Dynamic.TTL.Duration)
	assert.Equal(t, 20*time.Minute, ipt.Dynamic.Interval.Duration)
	assert.Equal(t, 50*time.Minute, ipt.Dynamic.TTL.Duration)
	assert.Equal(t, defaultMaxTestsPerRequest, ipt.Dynamic.MaxTestsPerRequest)
	assert.Equal(t, defaultMaxBodyBytes, ipt.Dynamic.MaxBodyBytes)
	assert.False(t, ipt.Dynamic.MonitorIPWithoutDomain)
	assert.False(t, ipt.ReverseDNS.Enabled)
	assert.Equal(t, defaultReverseDNSTimeout, ipt.ReverseDNS.Timeout.Duration)
	assert.Equal(t, defaultReverseDNSCacheTTL, ipt.ReverseDNS.CacheTTL.Duration)
	assert.Equal(t, defaultReverseDNSCacheSize, ipt.ReverseDNS.CacheSize)
}

func TestDynamicEnabledDefaultsToTrue(t *testing.T) {
	ipt := defaultInput()
	assert.True(t, ipt.Dynamic.Enabled)

	ipt.Dynamic.Enabled = false
	ipt.normalize()
	assert.False(t, ipt.Dynamic.Enabled)
}

func TestEffectiveDynamicProtocolUsesConfiguredUDP(t *testing.T) {
	assert.Equal(t, protocolUDP, effectiveDynamicProtocol(protocolTCP, protocolUDP, 443))
	assert.Equal(t, protocolUDP, effectiveDynamicProtocol("", protocolUDP, 443))
}

func TestProbeStatus(t *testing.T) {
	tests := []struct {
		name string
		res  probeResult
		want string
	}{
		{name: "reached", res: probeResult{tags: map[string]string{"traceroute_status": "reached"}}, want: "reached"},
		{name: "partial normalized", res: probeResult{tags: map[string]string{"traceroute_status": "PARTIAL"}}, want: "partial"},
		{name: "failed tag", res: probeResult{tags: map[string]string{"traceroute_status": "failed"}}, want: "failed"},
		{name: "probe error", res: probeResult{err: errors.New("probe failed")}, want: "fail"},
		{name: "missing status", res: probeResult{tags: map[string]string{}}, want: "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, probeStatus(tt.res))
		})
	}
}

func TestMeasurementUsesNetworkCategory(t *testing.T) {
	assert.Equal(t, point.Network, (&netpathMeasurement{}).Info().Cat)
}

func TestFeedResultUsesNetworkCategory(t *testing.T) {
	feeder := &captureNetpathFeeder{}
	ipt := &Input{feeder: feeder}

	ipt.feedResult(probeResult{
		task: task{
			ID:                "task-1",
			Name:              "external-api",
			Source:            sourceLocal,
			RunType:           runTypeScheduled,
			Protocol:          protocolTCP,
			Target:            "203.0.113.10",
			TargetIP:          "203.0.113.10",
			Port:              443,
			MaxTTL:            30,
			TracerouteQueries: 1,
		},
		tags:       map[string]string{"status": "OK"},
		fields:     map[string]interface{}{"hops": int64(1)},
		finishedAt: time.Unix(1, 0),
	})

	assert.Equal(t, point.Network, feeder.category)
	require.Len(t, feeder.points, 1)
	assert.Contains(t, feeder.points[0].LineProto(), "netpath")
}

func TestPointDataIncludesGlobalHostTags(t *testing.T) {
	datakit.ClearGlobalTags()
	t.Cleanup(datakit.ClearGlobalTags)
	datakit.SetGlobalHostTags("env", "global")
	datakit.SetGlobalHostTags("host", "node-a")
	datakit.SetGlobalHostTags("cloud_provider", "aws")

	tags, _ := pointData(probeResult{
		task: task{
			ID:       "task-1",
			Name:     "api",
			Source:   sourceLocal,
			RunType:  runTypeScheduled,
			Protocol: protocolTCP,
			Target:   "api.example.com",
			Tags:     map[string]string{"env": "input"},
		},
	})

	assert.Equal(t, "input", tags["env"])
	assert.Equal(t, "node-a", tags["host"])
	assert.Equal(t, "node-a", tags["source_host"])
	assert.Equal(t, "aws", tags["src_cloud_provider"])
}

func TestPointDataKeepsEndpointsWithoutHops(t *testing.T) {
	tags, fields := pointData(probeResult{
		task: task{
			ID:       "task-1",
			Name:     "external-api",
			Source:   sourceLocal,
			Origin:   "config",
			RunType:  runTypeScheduled,
			Target:   "203.0.113.10",
			Port:     443,
			Protocol: protocolTCP,
		},
		err: errors.New("operation not permitted"),
	})

	assert.Equal(t, "203.0.113.10", tags["dst_ip"])
	assert.Equal(t, "443", tags["dst_port"])
	assert.NotContains(t, tags, "target_ip")
	assert.Equal(t, "failed", tags["traceroute_status"])
	assert.Equal(t, "permission", tags["traceroute_fail_type"])
	assert.Equal(t, int64(0), fields["hop_count"])
}

func TestPointDataUsesProbeSourceIPAsEndpointFallback(t *testing.T) {
	tags, _ := pointData(probeResult{
		task: task{
			ID:       "task-1",
			Name:     "api",
			Source:   sourceLocal,
			RunType:  runTypeScheduled,
			Target:   "api.example.com",
			Protocol: protocolICMP,
		},
		tags: map[string]string{
			"probe_source_ip": "192.0.2.10",
			"probe_dest_ip":   "203.0.113.10",
		},
	})

	assert.Equal(t, "192.0.2.10", tags["src_ip"])
	assert.Equal(t, "*", tags["src_port"])
	assert.Equal(t, "203.0.113.10", tags["dst_ip"])
	assert.Equal(t, "*", tags["dst_port"])
}

func TestReadEnv(t *testing.T) {
	ipt := &Input{}
	ipt.ReadEnv(map[string]string{
		"ENV_INPUT_NETPATH_PROTOCOL":                          "icmp",
		"ENV_INPUT_NETPATH_INTERVAL":                          "2m",
		"ENV_INPUT_NETPATH_TIMEOUT":                           "3s",
		"ENV_INPUT_NETPATH_MAX_TTL":                           "40",
		"ENV_INPUT_NETPATH_TRACEROUTE_QUERIES":                "4",
		"ENV_INPUT_NETPATH_E2E_QUERIES":                       "12",
		"ENV_INPUT_NETPATH_TAGS":                              "env=production,team=network",
		"ENV_INPUT_NETPATH_DYNAMIC_ENABLED":                   "true",
		"ENV_INPUT_NETPATH_DYNAMIC_PROTOCOL":                  "udp",
		"ENV_INPUT_NETPATH_DYNAMIC_TOKEN":                     "test-token",
		"ENV_INPUT_NETPATH_DYNAMIC_TTL":                       "70m",
		"ENV_INPUT_NETPATH_DYNAMIC_INTERVAL":                  "30m",
		"ENV_INPUT_NETPATH_DYNAMIC_FLUSH_INTERVAL":            "5s",
		"ENV_INPUT_NETPATH_DYNAMIC_CONTEXTS_LIMIT":            "500",
		"ENV_INPUT_NETPATH_DYNAMIC_MAX_PER_MINUTE":            "120",
		"ENV_INPUT_NETPATH_DYNAMIC_WORKERS":                   "6",
		"ENV_INPUT_NETPATH_DYNAMIC_E2E_QUERIES":               "14",
		"ENV_INPUT_NETPATH_DYNAMIC_MONITOR_IP_WITHOUT_DOMAIN": "true",
	})

	assert.Equal(t, protocolICMP, ipt.Protocol)
	assert.Equal(t, 2*time.Minute, ipt.Interval.Duration)
	assert.Equal(t, 3*time.Second, ipt.Timeout.Duration)
	assert.Equal(t, 40, ipt.MaxTTL)
	assert.Equal(t, 4, ipt.TracerouteQueries)
	assert.Equal(t, 12, ipt.E2EQueries)
	assert.Equal(t, "production", ipt.Tags["env"])
	assert.Equal(t, "network", ipt.Tags["team"])
	require.NotNil(t, ipt.Dynamic)
	assert.True(t, ipt.Dynamic.Enabled)
	assert.Equal(t, protocolUDP, ipt.Dynamic.Protocol)
	assert.Equal(t, "test-token", ipt.Dynamic.Token)
	assert.Equal(t, 70*time.Minute, ipt.Dynamic.TTL.Duration)
	assert.Equal(t, 30*time.Minute, ipt.Dynamic.Interval.Duration)
	assert.Equal(t, 5*time.Second, ipt.Dynamic.FlushInterval.Duration)
	assert.Equal(t, 500, ipt.Dynamic.ContextsLimit)
	assert.Equal(t, 120, ipt.Dynamic.MaxPerMinute)
	assert.Equal(t, 6, ipt.Dynamic.Workers)
	assert.Equal(t, 14, ipt.Dynamic.E2EQueries)
	assert.True(t, ipt.Dynamic.MonitorIPWithoutDomain)
}

func TestTaskFromTargetConfig(t *testing.T) {
	ipt := defaultInput()
	target := TargetConfig{
		Name:       "api",
		Target:     "example.com",
		Port:       443,
		Protocol:   protocolTCP,
		E2EQueries: 7,
		Tags:       map[string]string{"service": "api"},
	}

	task, err := taskFromTargetConfig(target, ipt)
	require.NoError(t, err)
	assert.Equal(t, sourceLocal, task.Source)
	assert.Equal(t, runTypeScheduled, task.RunType)
	assert.Equal(t, protocolTCP, task.Protocol)
	assert.Equal(t, uint16(443), task.Port)
	assert.Equal(t, "api", task.Name)
	assert.Equal(t, "api", task.Tags["service"])
	assert.Equal(t, 7, task.E2EQueries)
	assert.NotEmpty(t, task.ID)
	assert.NotEmpty(t, task.ScheduleKey)
}

func TestTaskFromTargetConfigCapsTracerouteQueries(t *testing.T) {
	ipt := defaultInput()
	task, err := taskFromTargetConfig(TargetConfig{
		Target: "203.0.113.10", Protocol: protocolICMP, MaxTTL: 1_000_000, TracerouteQueries: 1_000_000,
	}, ipt)
	require.NoError(t, err)
	assert.Equal(t, maxTracerouteQueries, task.TracerouteQueries)
	assert.Equal(t, tr.MaxHops, task.MaxTTL)
	assert.Empty(t, task.Hostname)

	tags, _ := pointData(probeResult{task: task})
	assert.NotContains(t, tags, "dst_domain")

	udpTask, err := taskFromTargetConfig(TargetConfig{
		Target: "203.0.113.10", Port: 53, Protocol: protocolUDP, MaxTTL: 1_000_000,
	}, ipt)
	require.NoError(t, err)
	assert.Equal(t, maxTTL, udpTask.MaxTTL)
}

func TestTaskFromCandidate(t *testing.T) {
	cfg := normalizeDynamicConfig(&DynamicConfig{
		Tags: map[string]string{"global": "tag"},
	})
	req := candidateRequest{
		Source: "ebpf_netflow",
		Host:   "node-a",
		Tags:   map[string]string{"request": "tag"},
	}
	spec := candidateSpec{
		Hostname: "db.example.com",
		TargetIP: "198.51.100.20",
		Port:     5432,
		DstIP:    "203.0.113.10",
		DstPort:  5432,
		Protocol: protocolTCP,
		Origin:   "ebpf",
		Source: candidateSource{
			IP:          "192.0.2.10",
			Port:        50000,
			ProcessName: "app",
			ServiceName: "frontend",
		},
		Tags: map[string]string{"candidate": "tag"},
	}

	task, err := taskFromCandidate(req, spec, cfg)
	require.NoError(t, err)
	assert.Equal(t, sourceDynamic, task.Source)
	assert.Equal(t, runTypeDynamic, task.RunType)
	assert.Equal(t, protocolTCP, task.Protocol)
	assert.Equal(t, "198.51.100.20", task.Target)
	assert.Equal(t, "db.example.com", task.Hostname)
	assert.Equal(t, "node-a", task.SourceHost)
	assert.Equal(t, "192.0.2.10", task.SourceIP)
	assert.Equal(t, uint16(50000), task.SourcePort)
	assert.Equal(t, "frontend", task.SourceService)
	assert.Equal(t, "203.0.113.10", task.DstIP)
	assert.Equal(t, uint16(5432), task.DstPort)
	assert.Equal(t, "tag", task.ConfigTags["global"])
	assert.Equal(t, "tag", task.RequestTags["request"])
	assert.Equal(t, "tag", task.Tags["candidate"])
	assert.NotContains(t, task.Tags, "dst_domain")
	assert.NotContains(t, task.Tags, "target_ip")
	assert.NotContains(t, task.Tags, "target_port")
	assert.NotContains(t, task.Tags, "direction")
	assert.NotEmpty(t, task.ScheduleKey)

	spec.TargetIP = "198.51.100.21"
	secondTask, err := taskFromCandidate(req, spec, cfg)
	require.NoError(t, err)
	assert.Equal(t, task.ScheduleKey, secondTask.ScheduleKey)
	assert.Equal(t, "198.51.100.21", secondTask.Target)

	differentDstIP := spec
	differentDstIP.DstIP = "203.0.113.11"
	differentDstIPTask, err := taskFromCandidate(req, differentDstIP, cfg)
	require.NoError(t, err)
	assert.NotEqual(t, secondTask.ScheduleKey, differentDstIPTask.ScheduleKey)

	differentDstPort := spec
	differentDstPort.DstPort = 15432
	differentDstPortTask, err := taskFromCandidate(req, differentDstPort, cfg)
	require.NoError(t, err)
	assert.NotEqual(t, secondTask.ScheduleKey, differentDstPortTask.ScheduleKey)

	spec.DstIP = spec.TargetIP
	nonNATTask, err := taskFromCandidate(req, spec, cfg)
	require.NoError(t, err)
	assert.Equal(t, "db.example.com", nonNATTask.Target)
	rotatedNonNATSpec := spec
	rotatedNonNATSpec.TargetIP = "198.51.100.22"
	rotatedNonNATSpec.DstIP = rotatedNonNATSpec.TargetIP
	rotatedNonNATTask, err := taskFromCandidate(req, rotatedNonNATSpec, cfg)
	require.NoError(t, err)
	assert.Equal(t, nonNATTask.ScheduleKey, rotatedNonNATTask.ScheduleKey)

	spec.TargetIP = "not-an-ip"
	spec.DstIP = "203.0.113.10"
	_, err = taskFromCandidate(req, spec, cfg)
	require.ErrorContains(t, err, "invalid target")
}

func TestTaskFromCandidateSharesRequestTags(t *testing.T) {
	cfg := normalizeDynamicConfig(&DynamicConfig{Tags: map[string]string{"config": "tag"}})
	req := candidateRequest{Tags: map[string]string{"shared": "before"}}
	spec := candidateSpec{Hostname: "example.com", Port: 443, Protocol: protocolTCP, Tags: map[string]string{"test": "tag"}}

	task, err := taskFromCandidate(req, spec, cfg)
	require.NoError(t, err)
	req.Tags["shared"] = "after"

	assert.Equal(t, "after", task.RequestTags["shared"])
	assert.Equal(t, "tag", task.ConfigTags["config"])
	assert.Equal(t, map[string]string{"test": "tag"}, task.Tags)
	assert.NotContains(t, task.Tags, "shared")
}

func TestScheduleKeyEncodingPreventsDelimiterCollisions(t *testing.T) {
	cfg := normalizeDynamicConfig(&DynamicConfig{})
	spec := candidateSpec{
		Hostname: "api.example.com",
		Port:     443,
		Protocol: protocolTCP,
		Source: candidateSource{
			ServiceName: "part",
		},
	}
	first, err := taskFromCandidate(candidateRequest{Host: "host|service"}, spec, cfg)
	require.NoError(t, err)

	spec.Source.ServiceName = "service|part"
	second, err := taskFromCandidate(candidateRequest{Host: "host"}, spec, cfg)
	require.NoError(t, err)

	assert.NotEqual(t, first.ScheduleKey, second.ScheduleKey)
	assert.NotEqual(t, makeScheduleKey(sourceLocal, "same"), makeScheduleKey(sourceDynamic, "same"))
}

func TestEnrichProbeGateway(t *testing.T) {
	ipt := &Input{
		gatewayLookup: func(destinationIP string) (probeGateway, error) {
			assert.Equal(t, "203.0.113.10", destinationIP)
			return probeGateway{
				sourceIP:      "192.0.2.20",
				gatewayIP:     "192.0.2.1",
				interfaceName: "eth0",
				interfaceMAC:  "02:42:ac:11:00:02",
				netNS:         "4026531993",
			}, nil
		},
	}
	res := probeResult{
		task: task{Target: "198.51.100.7", TargetIP: "198.51.100.7"},
		tags: map[string]string{"probe_dest_ip": "203.0.113.10"},
	}

	ipt.enrichProbeGateway(&res)

	assert.Equal(t, "192.0.2.20", res.tags["probe_source_ip"])
	assert.Equal(t, "192.0.2.1", res.tags["probe_gateway_ip"])
	assert.Equal(t, "eth0", res.tags["probe_interface"])
	assert.Equal(t, "02:42:ac:11:00:02", res.tags["probe_interface_mac"])
	assert.Equal(t, "4026531993", res.tags["probe_netns"])
}

func TestTaskFromCandidateAcceptsUDP(t *testing.T) {
	cfg := normalizeDynamicConfig(&DynamicConfig{MaxTTL: maxTTL})
	task, err := taskFromCandidate(candidateRequest{Host: "node-a"}, candidateSpec{
		TargetIP: "10.0.0.12",
		Port:     53,
		Protocol: "udp",
	}, cfg)
	require.NoError(t, err)
	assert.Equal(t, protocolUDP, task.Protocol)
	assert.Equal(t, maxTTL, task.MaxTTL)

	tcpTask, err := taskFromCandidate(candidateRequest{Host: "node-a"}, candidateSpec{
		Hostname: "example.com",
		Port:     443,
		Protocol: protocolTCP,
	}, cfg)
	require.NoError(t, err)
	assert.Equal(t, tr.MaxHops, tcpTask.MaxTTL)
}

func TestResolveIPv4RejectsNonUnicastTarget(t *testing.T) {
	for _, target := range []string{"0.0.0.0", "224.0.0.1", "255.255.255.255"} {
		_, err := resolveIPv4Context(context.Background(), target, time.Second)
		require.ErrorContains(t, err, "non-unicast")
	}
}

func TestRunProbeReportsEffectiveMaxTTL(t *testing.T) {
	previous := runTraceroute
	runTraceroute = func(_ context.Context, ip net.IP, opt tr.Options) (tr.Result, error) {
		assert.Equal(t, "203.0.113.10", ip.String())
		assert.Equal(t, tr.MaxHops, opt.MaxTTL)
		assert.Equal(t, tr.ProtocolTCP, opt.Protocol)
		return tr.Result{Reached: true, Routes: []*tr.Route{{
			Items: []*tr.RouteItem{{IP: ip.String(), ResponseTime: 1000}},
		}}}, nil
	}
	t.Cleanup(func() { runTraceroute = previous })

	result := runProbe(task{
		ID:                "effective-max-ttl",
		Target:            "203.0.113.10",
		Port:              443,
		Protocol:          protocolTCP,
		MaxTTL:            maxTTL,
		TracerouteQueries: 1,
	})
	require.NoError(t, result.err)
	assert.Equal(t, tr.MaxHops, result.task.MaxTTL)

	_, fields := pointData(result)
	assert.Equal(t, int64(tr.MaxHops), fields["max_ttl"])
}

func TestUDPProbeResult(t *testing.T) {
	previous := runTraceroute
	calls := 0
	runTraceroute = func(_ context.Context, target net.IP, opt tr.Options) (tr.Result, error) {
		calls++
		assert.Equal(t, "203.0.113.10", target.String())
		assert.Equal(t, uint16(53), opt.Port)
		assert.Equal(t, 1, opt.Attempts)
		assert.Equal(t, tr.ProtocolUDP, opt.Protocol)
		return tr.Result{Reached: true, Routes: []*tr.Route{{
			Total:   1,
			Items:   []*tr.RouteItem{{IP: "203.0.113.10", ResponseTime: float64(calls * 1200)}},
			AvgCost: 1200,
		}}}, nil
	}
	t.Cleanup(func() { runTraceroute = previous })

	probe, err := buildProbeTask(task{
		Target:            "203.0.113.10",
		Port:              53,
		Protocol:          protocolUDP,
		Timeout:           time.Second,
		MaxTTL:            2,
		TracerouteQueries: 3,
	})
	require.NoError(t, err)
	require.NoError(t, probe.Check())
	require.NoError(t, probe.Run())

	tags, fields := probe.GetResults()
	assert.Equal(t, "reached", tags["traceroute_status"])
	assert.Equal(t, protocolUDP, tags["traceroute_protocol"])
	assert.Equal(t, "203.0.113.10", tags["probe_dest_ip"])
	assert.Equal(t, 1, fields["hops"])
	assert.Equal(t, 3, calls)
	var payload traceroutePayload
	require.NoError(t, json.Unmarshal([]byte(fields["traceroute"].(string)), &payload))
	require.Len(t, payload.Runs, 3)
	assert.Equal(t, "1", payload.Runs[0].RunID)
	assert.Equal(t, 1.2, payload.Runs[0].Hops[0].RTT)
	assert.Equal(t, "3", payload.Runs[2].RunID)
	assert.Equal(t, 3.6, payload.Runs[2].Hops[0].RTT)
}

func TestRunProbeKeepsPartialUDPPathOnError(t *testing.T) {
	previous := runTraceroute
	runTraceroute = func(context.Context, net.IP, tr.Options) (tr.Result, error) {
		return tr.Result{Routes: []*tr.Route{{
			Total: 1,
			Items: []*tr.RouteItem{{IP: "192.0.2.1", ResponseTime: 1200}},
		}}}, context.DeadlineExceeded
	}
	t.Cleanup(func() { runTraceroute = previous })

	result := runProbe(task{
		ID:                "partial-udp",
		Name:              "partial udp",
		Source:            sourceDynamic,
		RunType:           runTypeDynamic,
		Target:            "203.0.113.10",
		Port:              53,
		Protocol:          protocolUDP,
		Timeout:           time.Second,
		MaxTTL:            2,
		TracerouteQueries: 1,
	})
	require.ErrorIs(t, result.err, context.DeadlineExceeded)
	assert.Equal(t, "partial", result.tags["traceroute_status"])
	assert.Equal(t, 1, result.fields["hops"])
	assert.Contains(t, result.fields["traceroute_fail_reason"], "udp traceroute run 1")

	tags, fields := pointData(result)
	assert.Equal(t, "partial", tags["traceroute_status"])
	assert.Equal(t, "timeout", tags["traceroute_fail_type"])
	assert.Equal(t, int64(1), fields["hop_count"])
	assert.Contains(t, fields["traceroute_fail_reason"], "udp traceroute run 1")
}

func TestTCPTracerouteProbeRunsAllQueriesWithoutEndpointQualityProbe(t *testing.T) {
	previous := runTraceroute
	calls := 0
	runTraceroute = func(_ context.Context, ip net.IP, opt tr.Options) (tr.Result, error) {
		calls++
		assert.Equal(t, "203.0.113.10", ip.String())
		assert.Equal(t, 1, opt.Attempts)
		assert.Equal(t, tr.ProtocolTCP, opt.Protocol)
		return tr.Result{Reached: true, Routes: []*tr.Route{{Items: []*tr.RouteItem{{
			IP:           fmt.Sprintf("192.0.2.%d", calls+1),
			ResponseTime: float64(calls * 100),
		}}}, {
			Items: []*tr.RouteItem{{IP: "203.0.113.10", ResponseTime: float64(calls * 200)}},
		}}}, nil
	}
	t.Cleanup(func() { runTraceroute = previous })

	probe, err := buildProbeTask(task{
		Target: "203.0.113.10", Port: 443, Protocol: protocolTCP,
		MaxTTL: 30, TracerouteQueries: 3, Timeout: time.Second,
	})
	require.NoError(t, err)
	require.IsType(t, &tracerouteProbe{}, probe)
	require.NoError(t, probe.Check())
	require.NoError(t, probe.Run())
	tags, fields := probe.GetResults()
	assert.Equal(t, 3, calls)
	assert.Equal(t, "reached", tags["traceroute_status"])
	assert.Equal(t, protocolTCP, tags["traceroute_protocol"])
	assert.Equal(t, "203.0.113.10", tags["probe_dest_ip"])
	for _, key := range []string{
		"response_time", "response_time_with_dns",
		"average_round_trip_time", "packet_loss_percent", "packets_sent", "packets_received",
	} {
		assert.NotContains(t, fields, key)
	}

	var payload traceroutePayload
	require.NoError(t, json.Unmarshal([]byte(fields["traceroute"].(string)), &payload))
	require.Len(t, payload.Runs, 3)
	assert.Equal(t, "1", payload.Runs[0].RunID)
	assert.Equal(t, "192.0.2.2", payload.Runs[0].Hops[0].IPAddress)
	assert.Equal(t, "3", payload.Runs[2].RunID)
	assert.Equal(t, "192.0.2.4", payload.Runs[2].Hops[0].IPAddress)
	assert.Equal(t, 2, payload.HopCount.Max)
}

func TestTCPTracerouteResolvesDomainPerRun(t *testing.T) {
	previousResolve := resolveIPv4Fn
	previousTrace := runTraceroute
	destinations := []string{"203.0.113.10", "203.0.113.11", "203.0.113.10"}
	resolveCalls := 0
	resolveIPv4Fn = func(host string, _ time.Duration) (net.IP, error) {
		assert.Equal(t, "api.example.com", host)
		ip := net.ParseIP(destinations[resolveCalls])
		resolveCalls++
		return ip, nil
	}
	traceCalls := 0
	runTraceroute = func(_ context.Context, ip net.IP, opt tr.Options) (tr.Result, error) {
		assert.Equal(t, destinations[traceCalls], ip.String())
		assert.Equal(t, 1, opt.Attempts)
		assert.Equal(t, tr.ProtocolTCP, opt.Protocol)
		traceCalls++
		return tr.Result{Reached: true, Routes: []*tr.Route{{
			Items: []*tr.RouteItem{{IP: ip.String(), ResponseTime: 1000}},
		}}}, nil
	}
	t.Cleanup(func() {
		resolveIPv4Fn = previousResolve
		runTraceroute = previousTrace
	})

	probe, err := buildProbeTask(task{
		Target: "api.example.com", Hostname: "api.example.com", Port: 443,
		Protocol: protocolTCP, MaxTTL: 30, TracerouteQueries: 3, Timeout: time.Second,
	})
	require.NoError(t, err)
	require.NoError(t, probe.Check())
	require.NoError(t, probe.Run())

	tags, fields := probe.GetResults()
	assert.Equal(t, 3, resolveCalls)
	assert.Equal(t, 3, traceCalls)
	assert.Equal(t, "reached", tags["traceroute_status"])
	assert.Empty(t, tags["probe_dest_ip"])
	assert.Equal(t, "true", tags["probe_dest_ip_multiple"])

	var payload traceroutePayload
	require.NoError(t, json.Unmarshal([]byte(fields["traceroute"].(string)), &payload))
	require.Len(t, payload.Runs, 3)
	for i, run := range payload.Runs {
		require.NotNil(t, run.Destination)
		assert.Equal(t, destinations[i], run.Destination.IPAddress)
		assert.Equal(t, uint16(443), run.Destination.Port)
		assert.Equal(t, []string{"api.example.com"}, run.Destination.ReverseDNS)
	}
}

func TestUDPTracerouteResolvesDomainPerRun(t *testing.T) {
	previousResolve := resolveIPv4Fn
	previousTrace := runTraceroute
	destinations := []string{"203.0.113.20", "203.0.113.21"}
	resolveCalls := 0
	resolveIPv4Fn = func(host string, _ time.Duration) (net.IP, error) {
		assert.Equal(t, "dns.example.com", host)
		ip := net.ParseIP(destinations[resolveCalls])
		resolveCalls++
		return ip, nil
	}
	traceCalls := 0
	runTraceroute = func(_ context.Context, target net.IP, opt tr.Options) (tr.Result, error) {
		assert.Equal(t, destinations[traceCalls], target.String())
		assert.Equal(t, uint16(53), opt.Port)
		assert.Equal(t, tr.ProtocolUDP, opt.Protocol)
		traceCalls++
		return tr.Result{Reached: true, Routes: []*tr.Route{{
			Items: []*tr.RouteItem{{IP: target.String(), ResponseTime: 1000}},
		}}}, nil
	}
	t.Cleanup(func() {
		resolveIPv4Fn = previousResolve
		runTraceroute = previousTrace
	})

	probe, err := buildProbeTask(task{
		Target: "dns.example.com", Hostname: "dns.example.com", Port: 53,
		Protocol: protocolUDP, MaxTTL: 30, TracerouteQueries: 2, Timeout: time.Second,
	})
	require.NoError(t, err)
	require.NoError(t, probe.Check())
	require.NoError(t, probe.Run())

	tags, fields := probe.GetResults()
	assert.Equal(t, 2, resolveCalls)
	assert.Equal(t, 2, traceCalls)
	assert.Empty(t, tags["probe_dest_ip"])
	assert.Equal(t, "true", tags["probe_dest_ip_multiple"])

	var payload traceroutePayload
	require.NoError(t, json.Unmarshal([]byte(fields["traceroute"].(string)), &payload))
	require.Len(t, payload.Runs, 2)
	for i, run := range payload.Runs {
		require.NotNil(t, run.Destination)
		assert.Equal(t, destinations[i], run.Destination.IPAddress)
		assert.Equal(t, uint16(53), run.Destination.Port)
	}
}

func TestBuildProbeTaskUsesTracerouteOnlyProbeForTCPAndICMP(t *testing.T) {
	for _, tt := range []struct {
		name     string
		protocol string
		port     uint16
	}{
		{name: "tcp", protocol: protocolTCP, port: 443},
		{name: "icmp", protocol: protocolICMP},
	} {
		t.Run(tt.name, func(t *testing.T) {
			probe, err := buildProbeTask(task{
				Target: "203.0.113.10", Port: tt.port, Protocol: tt.protocol,
				Timeout: time.Second, MaxTTL: 30, TracerouteQueries: 3,
			})
			require.NoError(t, err)
			typed, ok := probe.(*tracerouteProbe)
			require.Truef(t, ok, "unexpected probe type %T", probe)
			assert.Equal(t, tr.Protocol(tt.protocol), typed.protocol)
			assert.Equal(t, 30, typed.maxTTL)
			assert.Equal(t, 3, typed.queries)
		})
	}
}

func TestCandidateDropReason(t *testing.T) {
	cfg := normalizeDynamicConfig(&DynamicConfig{})
	req := candidateRequest{Host: "node-a", Source: "ebpf_netflow"}
	spec := candidateSpec{TargetIP: "10.0.0.12", Port: 443, Protocol: protocolTCP}
	assert.Equal(t, "ip_without_domain", candidateDropReason(req, spec, cfg))

	spec = candidateSpec{Hostname: "10.0.0.12", Port: 443, Protocol: protocolTCP}
	assert.Equal(t, "ip_without_domain", candidateDropReason(req, spec, cfg))

	cfg = normalizeDynamicConfig(&DynamicConfig{
		MonitorIPWithoutDomain: true,
		Filters:                []FilterConfig{{Name: "private", DestCIDRs: []string{"10.0.0.0/8"}}},
	})
	assert.Equal(t, "filtered:private", candidateDropReason(req, spec, cfg))

	cfg = normalizeDynamicConfig(&DynamicConfig{
		MonitorIPWithoutDomain: true,
		Filters: []FilterConfig{{
			Name:       "skip-system",
			Namespaces: []string{"kube-system"},
		}},
	})
	spec.Namespace = "kube-system"
	assert.Equal(t, "filtered:skip-system", candidateDropReason(req, spec, cfg))

	cfg = normalizeDynamicConfig(&DynamicConfig{
		MonitorIPWithoutDomain: true,
		Filters: []FilterConfig{{
			Name:      "private-db",
			DestCIDRs: []string{"10.0.0.0/8"},
			Ports:     []uint16{5432},
		}},
	})
	spec.Namespace = ""
	spec.Port = 5432
	assert.Equal(t, "filtered:private-db", candidateDropReason(req, spec, cfg))

	spec.Port = 443
	assert.Empty(t, candidateDropReason(req, spec, cfg))

	cfg = normalizeDynamicConfig(&DynamicConfig{
		Protocol:               protocolUDP,
		MonitorIPWithoutDomain: true,
		Filters: []FilterConfig{{
			Name:      "skip-udp",
			Protocols: []string{protocolUDP},
		}},
	})
	spec.Protocol = protocolTCP
	assert.Equal(t, "filtered:skip-udp", candidateDropReason(req, spec, cfg))
}

func TestResolvedDestinationFilterBlocksDNSRotationBeforeProbe(t *testing.T) {
	cfg := normalizeDynamicConfig(&DynamicConfig{
		MonitorIPWithoutDomain: true,
		Queries:                2,
		Filters: []FilterConfig{{
			Name:             "private-container",
			DestCIDRs:        []string{"10.0.0.0/8"},
			SourceContainers: []string{"container-a"},
			Namespaces:       []string{"production"},
			Ports:            []uint16{443},
		}},
	})
	req := candidateRequest{Host: "node-a", Source: "ebpf_netflow"}
	spec := candidateSpec{
		Hostname:          "api.example.com",
		Port:              443,
		Protocol:          protocolTCP,
		Namespace:         "production",
		SourceContainerID: "legacy-container",
		Source: candidateSource{
			ContainerID: "container-a",
		},
	}
	assert.Empty(t, candidateDropReason(req, spec, cfg))
	probeTask, err := taskFromCandidate(req, spec, cfg)
	require.NoError(t, err)

	previousResolve := resolveIPv4Fn
	previousTrace := runTraceroute
	destinations := []string{"203.0.113.10", "10.20.30.40"}
	resolveCalls := 0
	resolveIPv4Fn = func(host string, _ time.Duration) (net.IP, error) {
		assert.Equal(t, "api.example.com", host)
		ip := net.ParseIP(destinations[resolveCalls])
		resolveCalls++
		return ip, nil
	}
	traceCalls := 0
	runTraceroute = func(_ context.Context, ip net.IP, _ tr.Options) (tr.Result, error) {
		traceCalls++
		return tr.Result{Reached: true, Routes: []*tr.Route{{
			Items: []*tr.RouteItem{{IP: ip.String(), ResponseTime: 1000}},
		}}}, nil
	}
	t.Cleanup(func() {
		resolveIPv4Fn = previousResolve
		runTraceroute = previousTrace
	})

	probe, err := buildProbeTask(probeTask)
	require.NoError(t, err)
	err = probe.Run()
	require.ErrorContains(t, err, "filtered:private-container")
	assert.Equal(t, 2, resolveCalls)
	assert.Equal(t, 1, traceCalls, "the filtered destination must not be probed")
}

func TestResolvedDestinationFilterBlocksUDPAndE2E(t *testing.T) {
	cfg := normalizeDynamicConfig(&DynamicConfig{
		MonitorIPWithoutDomain: true,
		Queries:                1,
		Filters: []FilterConfig{{
			Name:      "private",
			DestCIDRs: []string{"10.0.0.0/8"},
		}},
	})
	req := candidateRequest{Host: "node-a"}
	spec := candidateSpec{Hostname: "api.example.com", Port: 443, Protocol: protocolUDP}
	probeTask, err := taskFromCandidate(req, spec, cfg)
	require.NoError(t, err)

	previousResolve := resolveIPv4Fn
	previousTrace := runTraceroute
	previousTCPDial := dialE2ETCP
	resolveIPv4Fn = func(string, time.Duration) (net.IP, error) {
		return net.ParseIP("10.20.30.40"), nil
	}
	traceCalls := 0
	runTraceroute = func(context.Context, net.IP, tr.Options) (tr.Result, error) {
		traceCalls++
		return tr.Result{}, nil
	}
	tcpDialCalls := 0
	dialE2ETCP = func(context.Context, string) (net.Conn, error) {
		tcpDialCalls++
		return nil, errors.New("unexpected dial")
	}
	t.Cleanup(func() {
		resolveIPv4Fn = previousResolve
		runTraceroute = previousTrace
		dialE2ETCP = previousTCPDial
	})

	probe, err := buildProbeTask(probeTask)
	require.NoError(t, err)
	require.ErrorContains(t, probe.Run(), "filtered:private")
	assert.Equal(t, 0, traceCalls)

	probeTask.Protocol = protocolTCP
	result := runTCPE2E(probeTask)
	require.ErrorContains(t, result.err, "filtered:private")
	assert.Equal(t, 0, tcpDialCalls)
}

func TestTaskStoreFlushAndTTL(t *testing.T) {
	store := newTaskStore(10)
	now := time.Unix(100, 0)
	targetTask := task{
		Source:      sourceDynamic,
		Target:      "example.com",
		Protocol:    protocolICMP,
		ScheduleKey: "schedule",
		Interval:    time.Minute,
		TTL:         2 * time.Minute,
	}

	ok, reason := store.add(now, targetTask)
	require.True(t, ok)
	assert.Empty(t, reason)
	assert.Equal(t, 1, store.len())

	flushed, expired := store.flush(now, 1, sourceDynamic)
	require.Len(t, flushed, 1)
	assert.Equal(t, 0, expired)
	flushedTask, accountedBytes, reason := store.acquireRun(flushed[0])
	require.Empty(t, reason)
	assert.Equal(t, "example.com", flushedTask.Target)
	assert.Equal(t, now, flushedTask.ScheduledAt)
	store.releaseRun(flushed[0], accountedBytes)

	flushed, expired = store.flush(now.Add(30*time.Second), 1, sourceDynamic)
	assert.Empty(t, flushed)
	assert.Equal(t, 0, expired)
	flushed, expired = store.flush(now.Add(2*time.Minute), 1, sourceDynamic)
	assert.Len(t, flushed, 1)
	assert.Equal(t, 0, expired)
	flushedTask, accountedBytes, reason = store.acquireRun(flushed[0])
	require.Empty(t, reason)
	store.releaseRun(flushed[0], accountedBytes)
	flushed, expired = store.flush(now.Add(3*time.Minute), 1, sourceDynamic)
	assert.Empty(t, flushed)
	assert.Equal(t, 1, expired)
	assert.Equal(t, 0, store.len())

	ok, reason = store.add(now, task{
		Source: sourceDynamic, ScheduleKey: "expires-without-budget", TTL: time.Second, Interval: time.Minute,
	})
	require.True(t, ok)
	assert.Empty(t, reason)
	flushed, expired = store.flush(now.Add(2*time.Second), 0, sourceDynamic)
	assert.Empty(t, flushed)
	assert.Equal(t, 1, expired)
}

func TestTaskStoreKeepsLocalTasksOutsideDynamicLimits(t *testing.T) {
	store := newTaskStore(1)
	now := time.Unix(100, 0)
	newTask := func(source, key string) task {
		return task{Source: source, ScheduleKey: key, Interval: time.Minute, TTL: time.Minute}
	}

	ok, reason := store.add(now, newTask(sourceDynamic, "dynamic-1"))
	require.True(t, ok)
	assert.Empty(t, reason)
	ok, reason = store.add(now, newTask(sourceDynamic, "dynamic-2"))
	assert.False(t, ok)
	assert.Equal(t, "contexts_limit", reason)

	for _, key := range []string{"local-1", "local-2"} {
		localTask := newTask(sourceLocal, key)
		localTask.TTL = 0
		ok, reason = store.add(now, localTask)
		require.True(t, ok)
		assert.Empty(t, reason)
	}
	assert.Equal(t, 3, store.len())

	localTasks, expired := store.flush(now, 0, sourceLocal)
	assert.Len(t, localTasks, 2)
	assert.Zero(t, expired)
	dynamicTasks, expired := store.flush(now, 0, sourceDynamic)
	assert.Empty(t, dynamicTasks)
	assert.Zero(t, expired)
	dynamicTasks, expired = store.flush(now, 1, sourceDynamic)
	assert.Len(t, dynamicTasks, 1)
	assert.Zero(t, expired)
}

func TestScheduleLocalTargetsBypassesDynamicInputQueue(t *testing.T) {
	ipt := defaultInput()
	ipt.Targets = []TargetConfig{
		{Target: "192.0.2.1", Protocol: protocolICMP},
		{Target: "192.0.2.2", Protocol: protocolICMP},
	}
	ipt.Dynamic.ContextsLimit = 1
	ipt.Dynamic.InputQueue = 1
	ipt.scheduler = newScheduler(ipt, ipt.Dynamic)

	ipt.scheduleLocalTargets()

	assert.Equal(t, 2, ipt.scheduler.store.len())
	assert.Equal(t, 0, len(ipt.scheduler.inputCh))
}

func TestSchedulerDispatchUsesSeparateWorkerQueues(t *testing.T) {
	ipt := defaultInput()
	scheduler := newScheduler(ipt, ipt.Dynamic)
	dynamicTask := scheduledTask{source: sourceDynamic}
	localTask := scheduledTask{source: sourceLocal}

	scheduler.dispatchDynamic([]scheduledTask{dynamicTask})
	require.True(t, scheduler.dispatchLocal([]scheduledTask{localTask}))

	require.Len(t, scheduler.processCh, 1)
	require.Len(t, scheduler.localCh, 1)
	assert.Equal(t, sourceDynamic, (<-scheduler.processCh).source)
	assert.Equal(t, sourceLocal, (<-scheduler.localCh).source)
}

func TestIntervalWithJitter(t *testing.T) {
	interval := time.Minute
	got := intervalWithJitter(interval, "branch")
	assert.GreaterOrEqual(t, got, interval)
	assert.Less(t, got, interval+interval/20)
	assert.Equal(t, got, intervalWithJitter(interval, "branch"))
}

func TestMinuteRateLimiterKeepsFractionalTokens(t *testing.T) {
	now := time.Unix(100, 0)
	limiter := newMinuteRateLimiter(1, now)
	total := 0
	for second := 1; second <= 120; second++ {
		total += limiter.take(now.Add(time.Duration(second) * time.Second))
		if second < 60 {
			assert.Equal(t, 0, total)
		}
		if second == 60 {
			assert.Equal(t, 1, total)
		}
	}
	assert.Equal(t, 2, total)

	fast := newMinuteRateLimiter(120, now)
	assert.Equal(t, 2, fast.take(now.Add(time.Second)))
}

const candidateTestToken = "secret"

func TestNetpathInputOwnerLifecycle(t *testing.T) {
	activeCandidateInput.Lock()
	previousOwner := activeCandidateInput.owner
	previousCandidate := activeCandidateInput.candidate
	previousStarted := activeCandidateInput.started
	activeCandidateInput.owner = nil
	activeCandidateInput.candidate = nil
	activeCandidateInput.started = false
	activeCandidateInput.Unlock()
	t.Cleanup(func() {
		activeCandidateInput.Lock()
		activeCandidateInput.owner = previousOwner
		activeCandidateInput.candidate = previousCandidate
		activeCandidateInput.started = previousStarted
		activeCandidateInput.Unlock()
	})

	first := defaultInput()
	second := defaultInput()
	require.True(t, claimNetpathInput(first))
	assert.True(t, isNetpathInputOwner(first))
	assert.False(t, claimNetpathInput(first), "the same instance must not register twice")
	assert.False(t, claimNetpathInput(second))

	deactivateCandidateInput(second)
	assert.True(t, isNetpathInputOwner(first), "a stale input must not release the current owner")
	deactivateCandidateInput(first)
	assert.False(t, isNetpathInputOwner(first))
	require.True(t, claimNetpathInput(second))
	assert.True(t, isNetpathInputOwner(second))
	activeCandidateInput.Lock()
	activeCandidateInput.started = true
	activeCandidateInput.Unlock()
	deactivateCandidateInput(second)
	activeCandidateInput.RLock()
	assert.False(t, activeCandidateInput.started)
	activeCandidateInput.RUnlock()
}

func TestStaleNetpathInputDoesNotStart(t *testing.T) {
	ipt := defaultInput()
	ipt.Run()
	assert.Nil(t, ipt.scheduler)
}

func authorizeCandidateTestRequest(req *http.Request) {
	req.Header.Set(tokenHeader, candidateTestToken)
}

func TestHandleCandidates(t *testing.T) {
	ipt := defaultInput()
	ipt.Dynamic.Enabled = true
	ipt.Dynamic.Token = candidateTestToken

	body := bytes.NewBufferString(`{"source":"ebpf_netflow","host":"node-a","tests":[{"hostname":"example.com","port":443,"protocol":"tcp"}]}`)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, apiPath, bytes.NewReader(body.Bytes()))
	authorizeCandidateTestRequest(req)
	ipt.handleCandidates(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp candidateResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 1, resp.Accepted)
	assert.Equal(t, 0, resp.Dropped)
	assert.Equal(t, 1, resp.QueueSize)
	assert.Equal(t, 1, ipt.scheduler.store.dynamicLen())
	assert.Equal(t, 0, len(ipt.scheduler.inputCh))
}

func TestHandleCandidatesKeepsInitializedReverseDNSCache(t *testing.T) {
	ipt := defaultInput()
	ipt.Dynamic.Enabled = true
	ipt.Dynamic.Token = candidateTestToken
	ipt.initialize()
	rdns := ipt.rdns

	body := bytes.NewBufferString(`{"source":"ebpf_netflow","host":"node-a","tests":[{"hostname":"example.com","port":443,"protocol":"tcp"}]}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, apiPath, bytes.NewReader(body.Bytes()))
	req.RemoteAddr = "127.0.0.1:1234"
	authorizeCandidateTestRequest(req)
	ipt.handleCandidates(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Same(t, rdns, ipt.rdns)
}

func TestHandleCandidatesConcurrent(t *testing.T) {
	ipt := defaultInput()
	ipt.Dynamic.Enabled = true
	ipt.Dynamic.Token = candidateTestToken

	const requests = 16
	var wg sync.WaitGroup
	errs := make(chan error, requests)
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, apiPath,
				bytes.NewBufferString(`{"tests":[{"hostname":"example.com","port":443,"protocol":"tcp"}]}`))
			req.RemoteAddr = "127.0.0.1:1234"
			authorizeCandidateTestRequest(req)
			ipt.handleCandidates(w, req)
			if w.Code != http.StatusOK {
				errs <- fmt.Errorf("unexpected status %d", w.Code)
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

func TestHandleCandidatesRejectsInvalidToken(t *testing.T) {
	for _, tc := range []struct {
		name  string
		token string
	}{
		{name: "missing header"},
		{name: "wrong token", token: "wrong-secret"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ipt := defaultInput()
			ipt.Dynamic.Enabled = true
			ipt.Dynamic.Token = candidateTestToken

			body := bytes.NewBufferString(`{"tests":[{"hostname":"example.com","port":443,"protocol":"tcp"}]}`)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, apiPath, body)
			if tc.token != "" {
				req.Header.Set(tokenHeader, tc.token)
			}

			ipt.handleCandidates(w, req)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Equal(t, 0, ipt.scheduler.store.dynamicLen())
		})
	}
}

func TestHandleCandidatesAllowsLoopbackWithoutConfiguredToken(t *testing.T) {
	for _, tc := range []struct {
		name       string
		token      string
		remoteAddr string
	}{
		{name: "IPv4", remoteAddr: "127.0.0.1:1234"},
		{name: "IPv6", token: "   ", remoteAddr: "[::1]:1234"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ipt := defaultInput()
			ipt.Dynamic.Enabled = true
			ipt.Dynamic.Token = tc.token

			body := bytes.NewBufferString(`{"tests":[{"hostname":"example.com","port":443,"protocol":"tcp"}]}`)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, apiPath, body)
			req.RemoteAddr = tc.remoteAddr
			localIP := net.ParseIP("127.0.0.1")
			if strings.Contains(tc.remoteAddr, "::1") {
				localIP = net.ParseIP("::1")
			}
			req = req.WithContext(context.WithValue(req.Context(), http.LocalAddrContextKey,
				&net.TCPAddr{IP: localIP, Port: 9529}))

			ipt.handleCandidates(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, 1, ipt.scheduler.store.dynamicLen())
		})
	}
}

func TestHandleCandidatesRejectsLoopbackProxyOnNonLoopbackListener(t *testing.T) {
	ipt := defaultInput()
	ipt.Dynamic.Enabled = true
	body := bytes.NewBufferString(`{"tests":[{"hostname":"example.com","port":443,"protocol":"tcp"}]}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, apiPath, body)
	req.RemoteAddr = "127.0.0.1:1234"
	req = req.WithContext(context.WithValue(req.Context(), http.LocalAddrContextKey,
		&net.TCPAddr{IP: net.ParseIP("192.0.2.10"), Port: 9529}))

	ipt.handleCandidates(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, ipt.scheduler.store.dynamicLen())
}

func TestHandleCandidatesRejectsExternalWithoutConfiguredToken(t *testing.T) {
	for _, remoteAddr := range []string{
		"192.0.2.1:1234",
		"[2001:db8::1]:1234",
		"invalid",
		"",
	} {
		ipt := defaultInput()
		ipt.Dynamic.Enabled = true

		body := bytes.NewBufferString(`{"tests":[{"hostname":"example.com","port":443,"protocol":"tcp"}]}`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, apiPath, body)
		req.RemoteAddr = remoteAddr

		ipt.handleCandidates(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, 0, ipt.scheduler.store.dynamicLen())
	}
}

func TestHandleCandidatesDropsFilteredCandidates(t *testing.T) {
	ipt := defaultInput()
	ipt.Dynamic.Enabled = true
	ipt.Dynamic.Token = candidateTestToken
	ipt.Dynamic.MonitorIPWithoutDomain = true
	ipt.Dynamic.Filters = []FilterConfig{{
		Name:       "skip-system",
		Namespaces: []string{"kube-system"},
	}}
	ipt.scheduler = newScheduler(ipt, ipt.Dynamic)

	body := bytes.NewBufferString(`{"source":"ebpf_netflow","host":"node-a","tests":[{"hostname":"example.com","port":443,"protocol":"tcp","namespace":"kube-system"}]}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, apiPath, bytes.NewReader(body.Bytes()))
	req.RemoteAddr = "127.0.0.1:1234"
	authorizeCandidateTestRequest(req)

	ipt.handleCandidates(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp candidateResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 0, resp.Accepted)
	assert.Equal(t, 1, resp.Dropped)
	assert.Equal(t, 1, resp.DropReasons["filtered:skip-system"])
	assert.Equal(t, 0, len(ipt.scheduler.inputCh))
}

func TestHandleCandidatesLimitsRequest(t *testing.T) {
	t.Run("too many tests", func(t *testing.T) {
		ipt := defaultInput()
		ipt.Dynamic.Enabled = true
		ipt.Dynamic.Token = candidateTestToken
		ipt.Dynamic.MaxTestsPerRequest = 1
		ipt.scheduler = newScheduler(ipt, ipt.Dynamic)

		body := bytes.NewBufferString(`{"tests":[{"hostname":"a.example.com","port":443,"protocol":"tcp"},{"hostname":"b.example.com","port":443,"protocol":"tcp"}]}`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, apiPath, bytes.NewReader(body.Bytes()))
		req.RemoteAddr = "127.0.0.1:1234"
		authorizeCandidateTestRequest(req)

		ipt.handleCandidates(w, req)
		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
		assert.Equal(t, 0, len(ipt.scheduler.inputCh))
	})

	t.Run("body too large", func(t *testing.T) {
		ipt := defaultInput()
		ipt.Dynamic.Enabled = true
		ipt.Dynamic.Token = candidateTestToken
		ipt.Dynamic.MaxBodyBytes = 16

		body := bytes.NewBufferString(`{"tests":[{"hostname":"example.com","port":443,"protocol":"tcp"}]}`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, apiPath, bytes.NewReader(body.Bytes()))
		req.RemoteAddr = "127.0.0.1:1234"
		authorizeCandidateTestRequest(req)

		ipt.handleCandidates(w, req)
		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	})

	t.Run("trailing json", func(t *testing.T) {
		ipt := defaultInput()
		ipt.Dynamic.Enabled = true
		ipt.Dynamic.Token = candidateTestToken
		body := bytes.NewBufferString(`{"tests":[{"hostname":"example.com","port":443,"protocol":"tcp"}]} {}`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, apiPath, body)
		req.RemoteAddr = "127.0.0.1:1234"
		authorizeCandidateTestRequest(req)

		ipt.handleCandidates(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, 0, len(ipt.scheduler.inputCh))
	})

	t.Run("tag value too long", func(t *testing.T) {
		ipt := defaultInput()
		ipt.Dynamic.Enabled = true
		ipt.Dynamic.Token = candidateTestToken
		body, err := json.Marshal(candidateRequest{Tests: []candidateSpec{{
			Hostname: "example.com", Port: 443, Protocol: protocolTCP,
			Tags: map[string]string{"oversized": strings.Repeat("x", maxCandidateTagValueBytes+1)},
		}}})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, apiPath, bytes.NewReader(body))
		req.RemoteAddr = "127.0.0.1:1234"
		authorizeCandidateTestRequest(req)

		ipt.handleCandidates(w, req)
		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
		assert.Equal(t, 0, len(ipt.scheduler.inputCh))
	})

	t.Run("identity field too long", func(t *testing.T) {
		ipt := defaultInput()
		ipt.Dynamic.Enabled = true
		ipt.Dynamic.Token = candidateTestToken
		body, err := json.Marshal(candidateRequest{Tests: []candidateSpec{{
			Hostname: strings.Repeat("x", maxCandidateIdentityFieldBytes+1),
			Port:     443,
			Protocol: protocolTCP,
		}}})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, apiPath, bytes.NewReader(body))
		req.RemoteAddr = "127.0.0.1:1234"
		authorizeCandidateTestRequest(req)

		ipt.handleCandidates(w, req)
		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
		assert.Equal(t, 0, len(ipt.scheduler.inputCh))
	})

	t.Run("empty tag key", func(t *testing.T) {
		ipt := defaultInput()
		ipt.Dynamic.Enabled = true
		ipt.Dynamic.Token = candidateTestToken
		body := bytes.NewBufferString(`{"tests":[{"hostname":"example.com","port":443,"protocol":"tcp","tags":{" ":"value"}}]}`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, apiPath, body)
		req.RemoteAddr = "127.0.0.1:1234"
		authorizeCandidateTestRequest(req)

		ipt.handleCandidates(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, 0, len(ipt.scheduler.inputCh))
	})
}

func TestValidateCandidateTagsLimitsCombinedCountAndKeyLength(t *testing.T) {
	requestTags := make(map[string]string, maxCandidateTags)
	for i := 0; i < maxCandidateTags; i++ {
		requestTags[fmt.Sprintf("tag-%d", i)] = "value"
	}
	err := validateCandidateTags(candidateRequest{
		Tags:  requestTags,
		Tests: []candidateSpec{{Tags: map[string]string{"one-more": "value"}}},
	})
	require.ErrorContains(t, err, "combined request and test tags")

	err = validateCandidateTags(candidateRequest{
		Tests: []candidateSpec{{Tags: map[string]string{strings.Repeat("k", maxCandidateTagKeyBytes+1): "value"}}},
	})
	require.ErrorContains(t, err, "tag key exceeds")
}

func TestValidateCandidateIdentityLimitsFieldsAndTotal(t *testing.T) {
	err := validateCandidateIdentity(candidateRequest{Tests: []candidateSpec{{
		Hostname: "example.com",
		Source: candidateSource{
			ServiceName: strings.Repeat("x", maxCandidateIdentityFieldBytes+1),
		},
	}}})
	require.ErrorContains(t, err, "source.service_name")

	value := strings.Repeat("x", maxCandidateIdentityFieldBytes)
	err = validateCandidateIdentity(candidateRequest{
		Source: value,
		Host:   value,
		Tests: []candidateSpec{{
			Hostname:  value,
			Origin:    value,
			Namespace: "x",
		}},
	})
	require.ErrorContains(t, err, "identity fields exceed")

	err = validateCandidateIdentity(candidateRequest{
		Source: value,
		Host:   value,
		Tests:  []candidateSpec{{Hostname: value, Origin: value}},
	})
	require.NoError(t, err)
}

func TestNormalizeProductionBounds(t *testing.T) {
	shortFlush := datakit.Duration{Duration: time.Millisecond}
	ipt := &Input{
		Protocol:          "udp",
		MaxTTL:            999,
		TracerouteQueries: 999,
		E2EQueries:        999,
		Dynamic: &DynamicConfig{
			Protocol:           "udp",
			ContextsLimit:      999999,
			FlushInterval:      &shortFlush,
			MaxPerMinute:       999999,
			Workers:            999,
			MaxTTL:             999,
			Queries:            999,
			E2EQueries:         999,
			InputQueue:         999999,
			ProcessQueue:       999999,
			MaxTestsPerRequest: 999999,
			MaxBodyBytes:       999999999,
		},
	}
	ipt.normalize()

	assert.Equal(t, protocolUDP, ipt.Protocol)
	assert.Equal(t, maxTTL, ipt.MaxTTL)
	assert.Equal(t, maxTracerouteQueries, ipt.TracerouteQueries)
	assert.Equal(t, maxE2EQueries, ipt.E2EQueries)
	assert.Equal(t, protocolUDP, ipt.Dynamic.Protocol)
	assert.Equal(t, maxDynamicContextsLimit, ipt.Dynamic.ContextsLimit)
	assert.Equal(t, minDynamicFlushInterval, ipt.Dynamic.FlushInterval.Duration)
	assert.Equal(t, maxDynamicMaxPerMinute, ipt.Dynamic.MaxPerMinute)
	assert.Equal(t, maxDynamicWorkers, ipt.Dynamic.Workers)
	assert.Equal(t, maxTTL, ipt.Dynamic.MaxTTL)
	assert.Equal(t, maxTracerouteQueries, ipt.Dynamic.Queries)
	assert.Equal(t, maxE2EQueries, ipt.Dynamic.E2EQueries)
	assert.Equal(t, maxQueueSize, ipt.Dynamic.InputQueue)
	assert.Equal(t, maxQueueSize, ipt.Dynamic.ProcessQueue)
	assert.Equal(t, maxTestsPerRequest, ipt.Dynamic.MaxTestsPerRequest)
	assert.Equal(t, maxBodyBytes, ipt.Dynamic.MaxBodyBytes)
}

func TestPointData(t *testing.T) {
	res := probeResult{
		task: task{
			ID:         "task-1",
			Name:       "api",
			Source:     sourceLocal,
			Origin:     "config",
			RunType:    runTypeScheduled,
			Target:     "example.com",
			Port:       443,
			Protocol:   protocolTCP,
			Hostname:   "api.example.com",
			SourceIP:   "192.0.2.10",
			SourcePort: 50000,
			DstIP:      "1.1.1.1",
			DstPort:    443,
			SourcePID:  1234,
			Tags:       map[string]string{"service": "api"},
		},
		tags: map[string]string{
			"traceroute_status": "reached",
			"probe_dest_ip":     "1.1.1.1",
		},
		fields: map[string]interface{}{
			"hops":                 2,
			"traceroute":           `[{"ttl":1}]`,
			"e2e_dest_ip":          "1.1.1.2",
			"e2e_packets_sent":     int64(10),
			"e2e_packets_received": int64(10),
			"e2e_rtt_avg":          float64(1200),
		},
		duration: time.Second,
	}

	tags, fields := pointData(res)
	moveTracerouteToMessage(fields)
	assert.Equal(t, "api", tags["task_name"])
	assert.Equal(t, sourceLocal, tags["task_source"])
	_, hasName := tags["name"]
	_, hasSource := tags["source"]
	assert.False(t, hasName)
	assert.False(t, hasSource)
	assert.Equal(t, "api", tags["service"])
	assert.Equal(t, "reached", tags["traceroute_status"])
	assert.Equal(t, protocolTCP, tags["traceroute_protocol"])
	assert.Equal(t, "192.0.2.10", tags["src_ip"])
	assert.Equal(t, "50000", tags["src_port"])
	assert.Equal(t, "1.1.1.1", tags["dst_ip"])
	assert.Equal(t, "443", tags["dst_port"])
	assert.Equal(t, "api.example.com", tags["dst_domain"])
	assert.NotContains(t, tags, "branch_key")
	assert.NotEmpty(t, fields["test_run_id"])
	assert.NotContains(t, fields, "result_id")
	assert.Equal(t, int64(2), fields["hop_count"])
	assert.Equal(t, normalizeTracerouteJSON(`[{"ttl":1}]`), fields["message"])
	_, hasTraceroute := fields["traceroute"]
	_, hasHopsJSON := fields["hops_json"]
	_, hasHops := fields["hops"]
	assert.False(t, hasTraceroute)
	assert.False(t, hasHopsJSON)
	assert.False(t, hasHops)
	assert.NotContains(t, fields, "success")
	assert.NotContains(t, fields, "path_success")
	assert.Equal(t, "1.1.1.2", fields["e2e_dest_ip"])
	assert.Equal(t, int64(10), fields["e2e_packets_sent"])
	assert.Equal(t, int64(10), fields["e2e_packets_received"])
	assert.Equal(t, float64(1200), fields["e2e_rtt_avg"])
	assert.Equal(t, int64(1234), fields["source_pid"])
	assert.Equal(t, int64(time.Second/time.Microsecond), fields["duration"])
	assert.NotZero(t, fields["scheduled_at"])
	assert.NotZero(t, fields["started_at"])
	assert.NotContains(t, fields, "finished_at")
}

func TestPointsForResultDropsCompatibilityTags(t *testing.T) {
	res := probeResult{
		task: task{
			ID:         "task-1",
			Name:       "api",
			Source:     sourceDynamic,
			Origin:     "ebpf_netflow",
			RunType:    runTypeDynamic,
			Target:     "198.51.100.20",
			Hostname:   "api.example.com",
			TargetIP:   "198.51.100.20",
			Port:       8443,
			Protocol:   protocolTCP,
			SourceIP:   "192.0.2.10",
			SourcePort: 50000,
			DstIP:      "203.0.113.10",
			DstPort:    443,
			Tags: map[string]string{
				"proto":               protocolTCP,
				"target":              "198.51.100.20",
				"target_ip":           "198.51.100.20",
				"target_port":         "8443",
				"transport":           protocolTCP,
				"src_ip":              "192.0.2.10",
				"dst_ip":              "203.0.113.10",
				"dst_port":            "443",
				"dst_domain":          "api.example.com",
				"dst_nat_ip":          "198.51.100.20",
				"dst_nat_port":        "8443",
				"observed_dest_ip":    "203.0.113.10",
				"observed_dest_port":  "443",
				"nat_dest_ip":         "198.51.100.20",
				"nat_dest_port":       "8443",
				"path_key":            "legacy-path",
				"branch_key":          "legacy-branch",
				"cloud_provider":      "aliyun",
				"dest_cloud_provider": "legacy-destination",
			},
		},
		tags: map[string]string{
			"status":              "OK",
			"fail_type":           "legacy",
			"path_status":         "reached",
			"e2e_protocol":        protocolTCP,
			"probe_dest_ip":       "198.51.100.20",
			"traceroute_status":   "reached",
			"traceroute_protocol": protocolICMP,
		},
		fields: map[string]interface{}{
			"task_id":                  "legacy-task",
			"result_id":                "legacy-result",
			"finished_at":              int64(1),
			"success":                  int64(1),
			"path_success":             int64(1),
			"path_destination_reached": int64(1),
			"e2e_success":              int64(1),
			"fail_reason":              "legacy",
			"response_time":            int64(1200),
			"response_time_with_dns":   int64(1500),
			"average_round_trip_time":  float64(1000),
			"min_round_trip_time":      float64(800),
			"max_round_trip_time":      float64(1200),
			"packet_loss_percent":      float64(0),
			"packets_sent":             int64(3),
			"packets_received":         int64(3),
			"path_latency":             float64(1000),
			"path_latency_with_dns":    int64(1500),
			"path_rtt_avg":             float64(1000),
			"path_rtt_min":             float64(800),
			"path_rtt_max":             float64(1200),
			"path_packet_loss_percent": float64(0),
			"path_packets_sent":        int64(3),
			"path_packets_received":    int64(3),
			"e2e_packets_sent":         int64(10),
			"e2e_packets_received":     int64(10),
			"e2e_rtt_avg":              float64(900),
		},
	}

	pts := (&Input{}).pointsForResult(res)
	require.Len(t, pts, 1)
	pt := pts[0]
	lineProto := pt.LineProto()
	for _, key := range []string{
		"proto", "target", "target_ip", "target_port",
		"transport", "dest_ip", "dest_port", "observed_dest_ip", "observed_dest_port", "nat_dest_ip", "nat_dest_port",
		"status", "fail_type", "path_status", "e2e_protocol", "path_key", "branch_key",
		"cloud_provider", "dest_cloud_provider",
	} {
		assert.Emptyf(t, pt.GetTag(key), "legacy tag %s reached the Dataway point", key)
		assert.NotContainsf(t, lineProto, ","+key+"=", "legacy tag %s reached Dataway line protocol", key)
	}
	for _, key := range []string{
		"task_id", "result_id", "finished_at", "success", "path_success", "path_destination_reached", "e2e_success", "fail_reason",
		"response_time", "response_time_with_dns",
		"average_round_trip_time", "min_round_trip_time", "max_round_trip_time",
		"packet_loss_percent", "packets_sent", "packets_received",
		"path_latency", "path_latency_with_dns",
		"path_rtt_avg", "path_rtt_min", "path_rtt_max",
		"path_packet_loss_percent", "path_packets_sent", "path_packets_received",
	} {
		assert.Nilf(t, pt.Get(key), "compatibility field %s reached the Dataway point", key)
		assert.NotContainsf(t, lineProto, " "+key+"=", "compatibility field %s reached Dataway line protocol", key)
		assert.NotContainsf(t, lineProto, ","+key+"=", "compatibility field %s reached Dataway line protocol", key)
	}
	assert.Equal(t, "192.0.2.10", pt.GetTag("src_ip"))
	assert.Equal(t, "50000", pt.GetTag("src_port"))
	assert.Equal(t, "203.0.113.10", pt.GetTag("dst_ip"))
	assert.Equal(t, "443", pt.GetTag("dst_port"))
	assert.Equal(t, "api.example.com", pt.GetTag("dst_domain"))
	assert.Empty(t, pt.GetTag("src_cloud_provider"))
	assert.Empty(t, pt.GetTag("dst_cloud_provider"))
	assert.Equal(t, "198.51.100.20", pt.GetTag("dst_nat_ip"))
	assert.Equal(t, "8443", pt.GetTag("dst_nat_port"))
	assert.Equal(t, int64(10), pt.Get("e2e_packets_sent"))
	assert.Equal(t, int64(10), pt.Get("e2e_packets_received"))
	assert.Equal(t, float64(900), pt.Get("e2e_rtt_avg"))
}

func TestDomainResolutionDoesNotCreateNATTags(t *testing.T) {
	res := probeResult{
		task: task{
			ID:       "task-domain",
			Name:     "api",
			Source:   sourceDynamic,
			Origin:   "ebpf_netflow",
			RunType:  runTypeDynamic,
			Target:   "api.example.com",
			Hostname: "api.example.com",
			TargetIP: "203.0.113.10",
			DstIP:    "203.0.113.10",
			Port:     443,
			DstPort:  443,
			Protocol: protocolTCP,
		},
		tags: map[string]string{
			"probe_dest_ip_multiple": "true",
			"traceroute_status":      "reached",
		},
		fields: map[string]interface{}{
			"traceroute": `{"runs":[{"run_id":"1","destination":{"ip_address":"203.0.113.11","port":443},"hops":[]},{"run_id":"2","destination":{"ip_address":"203.0.113.12","port":443},"hops":[]}],"hop_count":{"avg":0,"min":0,"max":0}}`,
		},
	}

	pts := (&Input{}).pointsForResult(res)
	require.Len(t, pts, 1)
	pt := pts[0]
	assert.Equal(t, "203.0.113.10", pt.GetTag("dst_ip"))
	assert.Empty(t, pt.GetTag("dst_nat_ip"))
	assert.Empty(t, pt.GetTag("dst_nat_port"))
	assert.Empty(t, pt.GetTag("probe_dest_ip_multiple"))
}

func TestCandidateTargetIPRemainsOriginalDestinationAcrossDNSRotation(t *testing.T) {
	cfg := normalizeDynamicConfig(&DynamicConfig{})
	task, err := taskFromCandidate(candidateRequest{}, candidateSpec{
		Hostname: "api.example.com", TargetIP: "203.0.113.10", Port: 443, Protocol: protocolTCP,
	}, cfg)
	require.NoError(t, err)
	assert.Equal(t, "203.0.113.10", task.DstIP)

	tags, _ := pointData(probeResult{
		task: task,
		tags: map[string]string{
			"probe_dest_ip_multiple": "true",
			"traceroute_status":      "reached",
		},
	})
	assert.Equal(t, "203.0.113.10", tags["dst_ip"])
	assert.NotContains(t, tags, "dst_nat_ip")
	assert.NotContains(t, tags, "dst_nat_port")
}

func TestCustomTagsCannotOverrideContractTags(t *testing.T) {
	tags, _ := pointData(probeResult{
		task: task{
			Name: "real-task", Source: sourceDynamic, Origin: "ebpf", RunType: runTypeDynamic,
			Target: "203.0.113.10", Protocol: protocolTCP, SourceIP: "192.0.2.10",
			ConfigTags: map[string]string{"task_name": "forged-config"},
			RequestTags: map[string]string{
				"src_ip": "198.51.100.1", "traceroute_fail_type": "permission", "e2e_dest_ip": "198.51.100.2",
			},
			Tags: map[string]string{"traceroute_status": "failed", "duration": "forged", "custom": "kept"},
		},
		tags: map[string]string{"traceroute_status": "reached"},
	})

	assert.Equal(t, "real-task", tags["task_name"])
	assert.Equal(t, "192.0.2.10", tags["src_ip"])
	assert.Equal(t, "reached", tags["traceroute_status"])
	assert.NotContains(t, tags, "traceroute_fail_type")
	assert.NotContains(t, tags, "duration")
	assert.NotContains(t, tags, "e2e_dest_ip")
	assert.Equal(t, "kept", tags["custom"])
}

func TestTracerouteReverseDNS(t *testing.T) {
	fields := normalizeRunnerFields(map[string]interface{}{
		"traceroute": `[{"total":2,"failed":0,"loss":0,"avg_cost":102.5,"min_cost":100,"max_cost":105,"std_cost":2.5,"items":[{"ip":"10.0.0.1","response_time":100},{"ip":"10.0.0.3","response_time":105}]},{"total":2,"failed":2,"loss":100,"items":[{"ip":"*","response_time":0},{"ip":"*","response_time":0}]}]`,
	})

	var normalized traceroutePayload
	require.NoError(t, json.Unmarshal([]byte(fields["traceroute"].(string)), &normalized))
	require.Len(t, normalized.Runs, 2)
	require.Len(t, normalized.Runs[0].Hops, 2)
	assert.Equal(t, "10.0.0.1", normalized.Runs[0].Hops[0].IPAddress)
	assert.Equal(t, "10.0.0.3", normalized.Runs[1].Hops[0].IPAddress)
	assert.Empty(t, normalized.Runs[0].Hops[0].ReverseDNS)

	res := probeResult{
		task: task{
			ID:          "task-1",
			Name:        "api",
			Source:      sourceDynamic,
			Origin:      "ebpf_netflow",
			RunType:     runTypeDynamic,
			Target:      "example.com",
			TargetIP:    "10.0.0.2",
			Port:        443,
			Protocol:    protocolTCP,
			ScheduleKey: "schedule",
		},
		tags: map[string]string{
			"traceroute_status": "reached",
			"probe_dest_ip":     "10.0.0.2",
		},
		fields:   fields,
		duration: time.Second,
	}

	cfg := normalizeReverseDNSConfig(&ReverseDNSConfig{Enabled: true})
	rdns := newReverseDNSEnricher(cfg)
	rdns.lookupFn = func(ctx context.Context, ip string) ([]string, error) {
		switch ip {
		case "10.0.0.1":
			return []string{"gateway.local."}, nil
		case "10.0.0.3":
			return []string{"gateway-alt.local."}, nil
		case "10.0.0.2":
			return []string{"dest.local."}, nil
		default:
			return nil, errors.New("not found")
		}
	}

	ipt := &Input{rdns: rdns}
	pts := ipt.pointsForResult(res)
	require.Len(t, pts, 1)
	assert.Equal(t, metricName, pts[0].Name())
	assert.Equal(t, "10.0.0.2", pts[0].GetTag("dst_ip"), pts[0].LineProto())
	assert.Equal(t, "dest.local", pts[0].Get("dst_reverse_dns"), pts[0].LineProto())

	var traceroute traceroutePayload
	require.NoError(t, json.Unmarshal([]byte(pts[0].Get("message").(string)), &traceroute))
	require.Len(t, traceroute.Runs, 2)
	require.Len(t, traceroute.Runs[0].Hops, 2)
	require.Len(t, traceroute.Runs[1].Hops, 2)

	assert.Equal(t, "1", traceroute.Runs[0].RunID)
	assert.Equal(t, 1, traceroute.Runs[0].Hops[0].TTL)
	assert.Equal(t, "10.0.0.1", traceroute.Runs[0].Hops[0].IPAddress)
	assert.Equal(t, []string{"gateway.local"}, traceroute.Runs[0].Hops[0].ReverseDNS)
	assert.Equal(t, 0.1, traceroute.Runs[0].Hops[0].RTT)
	assert.True(t, traceroute.Runs[0].Hops[0].Reachable)
	assert.Equal(t, 2, traceroute.Runs[0].Hops[1].TTL)
	assert.Empty(t, traceroute.Runs[0].Hops[1].IPAddress)
	assert.False(t, traceroute.Runs[0].Hops[1].Reachable)

	assert.Equal(t, "2", traceroute.Runs[1].RunID)
	assert.Equal(t, "10.0.0.3", traceroute.Runs[1].Hops[0].IPAddress)
	assert.Equal(t, []string{"gateway-alt.local"}, traceroute.Runs[1].Hops[0].ReverseDNS)
	assert.Equal(t, 0.105, traceroute.Runs[1].Hops[0].RTT)
	assert.Equal(t, float64(2), traceroute.HopCount.Avg)
	assert.Equal(t, 2, traceroute.HopCount.Min)
	assert.Equal(t, 2, traceroute.HopCount.Max)
}

func TestNormalizeTraceroutePreservesEnrichment(t *testing.T) {
	raw := `{"runs":[{"run_id":"1","hops":[{"ttl":1,"ip_address":"8.8.8.8","reachable":true,"asn":15169,"as_name":"GOOGLE","as_prefix":"8.8.8.0/24","cloud_provider":"gcp"}]}]}`

	normalized := normalizeTracerouteJSON(raw)
	var payload traceroutePayload
	require.NoError(t, json.Unmarshal([]byte(normalized), &payload))
	require.Len(t, payload.Runs, 1)
	require.Len(t, payload.Runs[0].Hops, 1)
	hop := payload.Runs[0].Hops[0]
	assert.Equal(t, uint64(15169), hop.ASN)
	assert.Equal(t, "GOOGLE", hop.ASName)
	assert.Equal(t, "8.8.8.0/24", hop.ASPrefix)
	assert.Equal(t, "gcp", hop.CloudProvider)
}

func TestPointDataPartialPath(t *testing.T) {
	res := probeResult{
		task: task{
			ID:       "task-1",
			Name:     "udp-service",
			Source:   sourceDynamic,
			Origin:   "manual",
			RunType:  runTypeDynamic,
			Target:   "203.0.113.10",
			Port:     53,
			Protocol: protocolUDP,
		},
		tags: map[string]string{
			"traceroute_status":   "partial",
			"traceroute_protocol": protocolUDP,
		},
		fields: map[string]interface{}{
			"traceroute_fail_reason": "traceroute run 2: context deadline exceeded",
		},
		err: context.DeadlineExceeded,
	}

	tags, fields := pointData(res)
	assert.Equal(t, "partial", tags["traceroute_status"])
	assert.Equal(t, protocolUDP, tags["traceroute_protocol"])
	assert.Equal(t, "timeout", tags["traceroute_fail_type"])
	assert.Equal(t, "traceroute run 2: context deadline exceeded", fields["traceroute_fail_reason"])
	assert.NotContains(t, fields, "path_success")
}

func TestPointDataDoesNotCreatePathQualityAliases(t *testing.T) {
	res := probeResult{
		task: task{
			ID:       "task-1",
			Name:     "host",
			Source:   sourceLocal,
			Origin:   "config",
			RunType:  runTypeScheduled,
			Target:   "example.com",
			Protocol: protocolICMP,
		},
		tags: map[string]string{"status": "OK"},
		fields: map[string]interface{}{
			"average_round_trip_time": float64(1000),
			"min_round_trip_time":     float64(800),
			"max_round_trip_time":     float64(1200),
			"packet_loss_percent":     float64(0),
			"packets_sent":            int64(3),
			"packets_received":        int64(3),
		},
	}

	_, fields := pointData(res)
	assert.NotContains(t, fields, "path_success")
	for _, key := range []string{
		"path_latency", "path_latency_with_dns",
		"path_rtt_avg", "path_rtt_min", "path_rtt_max",
		"path_packet_loss_percent", "path_packets_sent", "path_packets_received",
	} {
		assert.NotContains(t, fields, key)
	}
}

func TestPointDataFailureType(t *testing.T) {
	res := probeResult{
		task: task{
			ID:       "task-1",
			Name:     "host",
			Source:   sourceLocal,
			Origin:   "config",
			RunType:  runTypeScheduled,
			Target:   "example.com",
			Protocol: protocolTCP,
		},
		err: errors.New("lookup example.com: no such host"),
	}

	tags, fields := pointData(res)
	assert.Equal(t, "failed", tags["traceroute_status"])
	assert.Equal(t, "dns_error", tags["traceroute_fail_type"])
	assert.NotContains(t, fields, "path_success")
	assert.Equal(t, "lookup example.com: no such host", fields["traceroute_fail_reason"])

	assert.Equal(t, "timeout", classifyFailure("i/o timeout"))
	assert.Equal(t, "permission", classifyFailure("operation not permitted"))
	assert.Equal(t, "protocol_unsupported", classifyFailure("unsupported netpath protocol sctp"))
	assert.Equal(t, "invalid_target", classifyFailure("target resolved to non-unicast address"))
	assert.Equal(t, "target_unreachable", classifyFailure("no route to host"))
	assert.Equal(t, "connection_refused", classifyFailure("connection refused"))
}

func TestInputHelpers(t *testing.T) {
	ipt := defaultInput()
	assert.Contains(t, ipt.SampleConfig(), "[[inputs.netpath]]")
	assert.Equal(t, "network", ipt.Catalog())
	assert.NotEmpty(t, ipt.AvailableArchs())
	assert.NotContains(t, ipt.AvailableArchs(), datakit.OSLabelWindows)
	assert.True(t, netpathSupportedOS(datakit.OSLinux))
	assert.True(t, netpathSupportedOS(datakit.OSDarwin))
	assert.False(t, netpathSupportedOS(datakit.OSWindows))
	assert.False(t, netpathSupportedOS("unknown"))
	require.Len(t, ipt.SampleMeasurement(), 1)
	assert.Equal(t, metricName, ipt.SampleMeasurement()[0].Info().Name)

	ipt.Terminate()
	assert.NotPanics(t, func() { ipt.Terminate() })
}
