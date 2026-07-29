// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"strings"
	"time"

	dt "github.com/GuanceCloud/cliutils/dialtesting"
)

const emptyDialTraceroute = `{"runs":[],"hop_count":{"avg":0,"min":0,"max":0}}`

type DialProbeResult struct {
	Tags        map[string]string
	Fields      map[string]interface{}
	Duration    time.Duration
	ScheduledAt time.Time
	StartedAt   time.Time
	FinishedAt  time.Time
	TestRunID   string
	FailType    string
	Err         error
}

func RunDialProbe(
	ctx context.Context,
	cfg dt.NetPathProbeConfig,
	validateResolvedIP func(net.IP) error,
) DialProbeResult {
	if err := CheckDialPlatform(cfg.Protocol); err != nil {
		now := time.Now()
		scheduledAt := cfg.ScheduledAt
		if scheduledAt.IsZero() {
			scheduledAt = now
		}
		return DialProbeResult{
			Tags:        map[string]string{},
			Fields:      map[string]interface{}{},
			ScheduledAt: scheduledAt,
			StartedAt:   now,
			FinishedAt:  now,
			FailType:    "unsupported_platform",
			Err:         err,
		}
	}

	host := strings.TrimSpace(cfg.Host)
	hostname := host
	if net.ParseIP(host) != nil {
		hostname = ""
	}

	res := runProbeContext(ctx, task{
		ID:                  cfg.ID,
		Name:                cfg.Name,
		Target:              host,
		Hostname:            hostname,
		Port:                cfg.Port,
		DstPort:             cfg.Port,
		Protocol:            strings.ToLower(strings.TrimSpace(cfg.Protocol)),
		ScheduleKey:         cfg.ID,
		Timeout:             cfg.Timeout,
		MaxTTL:              cfg.MaxTTL,
		TracerouteQueries:   cfg.TracerouteQueries,
		E2EQueries:          cfg.E2EQueries,
		ScheduledAt:         cfg.ScheduledAt,
		resolvedIPValidator: validateResolvedIP,
	})
	enrichProbeGatewayWithLookup(&res, lookupCurrentProbeGateway)

	tags := cloneTags(res.tags)
	fields := make(map[string]interface{}, len(res.fields)+8)
	for key, value := range res.fields {
		fields[key] = value
	}
	if hops, ok := asInt64(fields["hops"]); ok {
		fields["hop_count"] = hops
	}
	delete(fields, "hops")
	delete(fields, "hops_json")
	fields["max_ttl"] = int64(res.task.MaxTTL)
	fields["traceroute_queries"] = int64(res.task.TracerouteQueries)
	fields["e2e_queries"] = int64(res.task.E2EQueries)
	if _, ok := fields["hop_count"]; !ok {
		fields["hop_count"] = int64(0)
	}
	normalizeDialTraceroute(fields)

	return DialProbeResult{
		Tags:        tags,
		Fields:      fields,
		Duration:    res.duration,
		ScheduledAt: res.scheduledAt,
		StartedAt:   res.startedAt,
		FinishedAt:  res.finishedAt,
		TestRunID:   res.testRunID,
		FailType:    res.failType,
		Err:         res.err,
	}
}

func normalizeDialTraceroute(fields map[string]interface{}) {
	raw, ok := fields["traceroute"].(string)
	if !ok || strings.TrimSpace(raw) == "" || strings.TrimSpace(raw) == "[]" {
		fields["traceroute"] = emptyDialTraceroute
	}
}

// CheckDialPlatform mirrors the checks performed by the netpath input so the
// central dialtesting path can reject unsupported OS/protocol combinations
// both at task dispatch time and before launching a probe.
func CheckDialPlatform(protocol string) error {
	return validateDialPlatformFor(runtime.GOOS, protocol)
}

// validateDialPlatformFor is the testable core: it accepts the GOOS explicitly
// so callers can assert the OS rejection path on any platform.
func validateDialPlatformFor(goos, protocol string) error {
	if !netpathSupportedOS(goos) {
		return fmt.Errorf("netpath dial testing is unsupported on %s", goos)
	}
	protocol = strings.ToLower(strings.TrimSpace(protocol))
	if goos == "darwin" && (protocol == protocolTCP || protocol == protocolUDP) {
		return fmt.Errorf("unsupported netpath protocol %q on darwin", protocol)
	}
	return nil
}
