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

	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (ipt *Input) feedResult(res probeResult) {
	ipt.feedResultContext(context.Background(), res)
}

func (ipt *Input) feedResultContext(ctx context.Context, res probeResult) {
	if ctx == nil {
		ctx = context.Background()
	}
	pts := ipt.pointsForResultContext(ctx, res)
	if len(pts) == 0 {
		return
	}
	if ctx.Err() != nil {
		return
	}

	if ipt.feeder == nil {
		ipt.reportFeedError(fmt.Errorf("netpath feeder is not initialized"))
		return
	}

	if err := ipt.feeder.Feed(point.Network, pts,
		dkio.WithSource(inputName), dkio.WithInput(inputName)); err != nil {
		ipt.reportFeedError(err)
	}
}

func (ipt *Input) reportFeedError(err error) {
	incFeedError()
	if ipt.feeder != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(point.Network),
		)
	}
	l.Errorf("feed netpath result: %s", err.Error())
}

func (ipt *Input) pointsForResult(res probeResult) []*point.Point {
	return ipt.pointsForResultContext(context.Background(), res)
}

func (ipt *Input) pointsForResultContext(ctx context.Context, res probeResult) []*point.Point {
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Err() != nil {
		return nil
	}
	ensureResultIdentity(&res)
	tags, fields := pointData(res)
	if ipt.rdns != nil {
		rdnsCtx, cancel := newReverseDNSEnrichmentContext(ctx)
		if name, ok := ipt.rdns.lookupContext(rdnsCtx,
			valueOrDefault(tags["dst_nat_ip"], tags["dst_ip"])); ok {
			fields["dst_reverse_dns"] = name
		}
		if ctx.Err() != nil {
			cancel()
			return nil
		}
		enrichTracerouteFieldsContext(rdnsCtx, fields, ipt.rdns)
		cancel()
	}
	if ctx.Err() != nil {
		return nil
	}
	dropLegacyTags(tags)
	dropLegacyFields(fields)
	moveTracerouteToMessage(fields)

	ts := res.finishedAt
	if ts.IsZero() {
		ts = time.Now()
	}
	ts = ntp.CalibrateTime(ts)
	return []*point.Point{
		point.NewPoint(metricName,
			append(point.NewTags(tags), point.NewKVs(fields)...),
			append(point.CommonLoggingOptions(), point.WithTime(ts))...),
	}
}

// dropLegacyTags keeps internal runner context and legacy aliases out of the
// Dataway point.
func dropLegacyTags(tags map[string]string) {
	for _, key := range []string{
		"proto",
		"target",
		"target_ip",
		"target_port",
		"transport",
		"dest_host",
		"dest_ip",
		"dest_port",
		"dest_domain",
		"observed_dest_ip",
		"observed_dest_port",
		"nat_dest_ip",
		"nat_dest_port",
		"source_ip",
		"status",
		"fail_type",
		"path_status",
		"e2e_protocol",
		"path_key",
		"branch_key",
		"cloud_provider",
		"source_cloud_provider",
		"dest_cloud_provider",
		"probe_dest_ip",
		"probe_dest_ip_multiple",
	} {
		delete(tags, key)
	}
}

func pointData(res probeResult) (map[string]string, map[string]interface{}) {
	t := res.task
	ensureResultIdentity(&res)
	globalHostTags := datakit.GlobalHostTags()
	tags := map[string]string{}
	mergeCustomTags(tags, t.ConfigTags)
	mergeCustomTags(tags, t.RequestTags)
	mergeCustomTags(tags, t.Tags)
	tags = inputs.MergeTags(globalHostTags, tags, "")
	for key := range reservedTagKeys {
		if key != "host" && key != "cloud_provider" {
			delete(tags, key)
		}
	}
	for k, v := range res.tags {
		tags[k] = v
	}
	tags["task_name"] = t.Name
	tags["task_source"] = t.Source
	tags["origin"] = valueOrDefault(t.Origin, t.Source)
	tags["run_type"] = t.RunType
	tags["protocol"] = t.Protocol
	if strings.TrimSpace(tags["traceroute_protocol"]) == "" {
		tags["traceroute_protocol"] = tracerouteProtocol(t.Protocol)
	}
	tags["namespace"] = t.Namespace
	tags["source_container_id"] = t.SourceContainerID
	tags["source_host"] = t.SourceHost
	tags["source_process"] = t.SourceProcess
	tags["source_service"] = t.SourceService
	tags["netns"] = t.NetNS
	normalizeEndpointTags(tags, t)

	fields := map[string]interface{}{
		"test_run_id":        res.testRunID,
		"duration":           res.duration.Microseconds(),
		"max_ttl":            int64(t.MaxTTL),
		"traceroute_queries": int64(t.TracerouteQueries),
		"hop_count":          int64(0),
		"traceroute":         "[]",
		"hops_json":          "[]",
		"scheduled_at":       calibratedUnixMicroOrZero(res.scheduledAt),
		"started_at":         calibratedUnixMicroOrZero(res.startedAt),
	}
	if t.SourcePID != 0 {
		fields["source_pid"] = int64(t.SourcePID)
	}
	for k, v := range res.fields {
		fields[k] = v
	}
	if hops, ok := asInt64(fields["hops"]); ok {
		fields["hop_count"] = hops
	}
	delete(fields, "hops")
	if res.err != nil {
		if _, ok := fields["traceroute_fail_reason"]; !ok {
			fields["traceroute_fail_reason"] = res.err.Error()
		}
		fields["message"] = res.err.Error()
		if tags["traceroute_status"] == "" {
			tags["traceroute_status"] = "failed"
		}
		tags["traceroute_fail_type"] = valueOrDefault(res.failType, classifyFailure(res.err.Error()))
	}
	dropEmptyTags(tags)
	return tags, fields
}

var reservedTagKeys = map[string]struct{}{ //nolint:gochecknoglobals
	"host": {}, "task_name": {}, "task_source": {}, "origin": {}, "run_type": {}, "protocol": {},
	"traceroute_protocol": {}, "traceroute_status": {}, "traceroute_fail_type": {}, "e2e_status": {},
	"src_ip": {}, "src_port": {}, "dst_ip": {}, "dst_port": {}, "dst_domain": {},
	"dst_nat_ip": {}, "dst_nat_port": {}, "namespace": {}, "source_host": {}, "source_process": {},
	"source_service": {}, "source_container_id": {}, "netns": {}, "src_cloud_provider": {},
	"dst_cloud_provider": {}, "cloud_provider": {}, "source_cloud_provider": {}, "dest_cloud_provider": {},
	"probe_source_ip": {}, "probe_dest_ip": {}, "probe_dest_ip_multiple": {}, "probe_gateway_ip": {},
	"probe_interface": {}, "probe_interface_mac": {}, "probe_netns": {},
	"proto": {}, "target": {}, "target_ip": {}, "target_port": {}, "transport": {},
	"dest_host": {}, "dest_ip": {}, "dest_port": {}, "dest_domain": {}, "observed_dest_ip": {},
	"observed_dest_port": {}, "nat_dest_ip": {}, "nat_dest_port": {}, "source_ip": {},
	"status": {}, "fail_type": {}, "path_status": {}, "e2e_protocol": {}, "path_key": {}, "branch_key": {},
	"test_run_id": {}, "scheduled_at": {}, "started_at": {}, "duration": {}, "source_pid": {},
	"max_ttl": {}, "traceroute_queries": {}, "hop_count": {}, "traceroute_fail_reason": {}, "message": {},
	"dst_reverse_dns": {}, "e2e_dest_ip": {}, "e2e_queries": {}, "e2e_packets_sent": {}, "e2e_packets_received": {},
	"e2e_unknown": {}, "e2e_probe_loss_percent": {}, "e2e_rtt_avg": {}, "e2e_rtt_min": {},
	"e2e_rtt_max": {}, "e2e_rtt_variation_samples": {}, "e2e_rtt_variation_avg": {},
	"e2e_rtt_variation_max": {}, "e2e_tcp_connection_refused": {}, "e2e_fail_reason": {},
}

func mergeCustomTags(dst, src map[string]string) {
	for key, value := range src {
		if _, reserved := reservedTagKeys[key]; reserved {
			continue
		}
		dst[key] = value
	}
}

// normalizeEndpointTags keeps the path endpoints searchable even when a probe
// fails before traceroute produces any hops. Explicit candidate and runner
// values win; host-level context is only used as a fallback.
func normalizeEndpointTags(tags map[string]string, t task) {
	if tags == nil {
		return
	}

	actualDestinationIP := strings.TrimSpace(tags["probe_dest_ip"])
	if actualDestinationIP == "" && net.ParseIP(strings.TrimSpace(t.Target)) != nil {
		actualDestinationIP = strings.TrimSpace(t.Target)
	}

	destinationIP := strings.TrimSpace(t.DstIP)
	if destinationIP == "" {
		destinationIP = strings.TrimSpace(tags["observed_dest_ip"])
	}
	if destinationIP == "" {
		destinationIP = actualDestinationIP
	}
	destinationPort := t.DstPort
	if destinationPort == 0 {
		destinationPort = parsePortTag(tags["observed_dest_port"])
	}
	destinationPort = firstPositivePort(destinationPort, t.Port)
	tags["dst_ip"] = destinationIP
	tags["dst_port"] = portTag(destinationPort)
	delete(tags, "dst_domain")
	if strings.TrimSpace(t.Hostname) != "" {
		tags["dst_domain"] = strings.TrimSpace(t.Hostname)
	}
	delete(tags, "dst_nat_ip")
	delete(tags, "dst_nat_port")
	translatedDestinationIP := strings.TrimSpace(t.TargetIP)
	if translatedDestinationIP == "" && net.ParseIP(strings.TrimSpace(t.Target)) != nil {
		translatedDestinationIP = strings.TrimSpace(t.Target)
	}
	if translatedDestinationIP != "" &&
		(translatedDestinationIP != destinationIP || t.Port != destinationPort) {
		tags["dst_nat_ip"] = translatedDestinationIP
		tags["dst_nat_port"] = portTag(t.Port)
	}

	if strings.TrimSpace(tags["source_host"]) == "" {
		tags["source_host"] = strings.TrimSpace(tags["host"])
	}
	tags["src_ip"] = valueOrDefault(t.SourceIP, tags["probe_source_ip"])
	tags["src_port"] = portTag(t.SourcePort)
	if strings.TrimSpace(tags["src_cloud_provider"]) == "" {
		tags["src_cloud_provider"] = strings.TrimSpace(tags["cloud_provider"])
	}
	if strings.TrimSpace(tags["dst_cloud_provider"]) == "" {
		tags["dst_cloud_provider"] = strings.TrimSpace(tags["dest_cloud_provider"])
	}
}

func portTag(port uint16) string {
	if port == 0 {
		return "*"
	}
	return fmt.Sprintf("%d", port)
}

func parsePortTag(value string) uint16 {
	port, err := strconv.ParseUint(strings.TrimSpace(value), 10, 16)
	if err != nil {
		return 0
	}
	return uint16(port)
}

func isLegacyField(name string) bool {
	switch name {
	case "response_time", "response_time_with_dns",
		"average_round_trip_time", "min_round_trip_time", "max_round_trip_time",
		"packet_loss_percent", "packets_sent", "packets_received",
		"path_latency", "path_latency_with_dns",
		"path_rtt_avg", "path_rtt_min", "path_rtt_max",
		"path_packet_loss_percent", "path_packets_sent", "path_packets_received",
		"task_id", "result_id", "finished_at",
		"success", "path_success", "path_destination_reached", "e2e_success",
		"traceroute_protocol", "fail_reason":
		return true
	default:
		return false
	}
}

func ensureResultIdentity(res *probeResult) {
	if res.startedAt.IsZero() {
		res.startedAt = time.Now()
	}
	if res.scheduledAt.IsZero() {
		res.scheduledAt = res.startedAt
	}
	if res.finishedAt.IsZero() {
		res.finishedAt = res.startedAt.Add(res.duration)
	}
	if res.testRunID == "" {
		res.testRunID = makeRunID(res.task, res.startedAt)
	}
}

func calibratedUnixMicroOrZero(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return ntp.CalibrateTime(t).UnixMicro()
}

func dropLegacyFields(fields map[string]interface{}) {
	for key := range fields {
		if isLegacyField(key) {
			delete(fields, key)
		}
	}
}

func dropEmptyTags(tags map[string]string) {
	for key, value := range tags {
		if strings.TrimSpace(value) == "" {
			delete(tags, key)
		}
	}
}

// moveTracerouteToMessage keeps the full path only once, in the JSON-indexed
// message field. The runner retains traceroute temporarily for path completion
// detection before the result is converted into a point.
func moveTracerouteToMessage(fields map[string]interface{}) {
	raw, _ := fields["traceroute"].(string)
	if strings.TrimSpace(raw) == "" || raw == "[]" {
		raw, _ = fields["hops_json"].(string)
	}
	if normalized := normalizeTracerouteJSON(raw); normalized != "" {
		fields["message"] = normalized
	}
	delete(fields, "traceroute")
	delete(fields, "hops_json")
}

func classifyFailure(message string) string {
	msg := strings.ToLower(message)
	switch {
	case strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded"):
		return "timeout"
	case strings.Contains(msg, "lookup") || strings.Contains(msg, "no such host") || strings.Contains(msg, "dns"):
		return "dns_error"
	case strings.Contains(msg, "permission") ||
		strings.Contains(msg, "operation not permitted") ||
		strings.Contains(msg, "access denied"):
		return "permission"
	case strings.Contains(msg, "unsupported") || strings.Contains(msg, "protocol"):
		return "protocol_unsupported"
	case strings.Contains(msg, "missing target") ||
		strings.Contains(msg, "missing port") ||
		strings.Contains(msg, "invalid target") ||
		strings.Contains(msg, "non-unicast") ||
		strings.Contains(msg, "unicast ipv4"):
		return "invalid_target"
	case strings.Contains(msg, "connection refused"):
		return "connection_refused"
	case strings.Contains(msg, "no route") ||
		strings.Contains(msg, "network unreachable") ||
		strings.Contains(msg, "host unreachable"):
		return "target_unreachable"
	case msg == "":
		return "unknown"
	default:
		return "runner_error"
	}
}

func asInt64(v interface{}) (int64, bool) {
	switch x := v.(type) {
	case int:
		return int64(x), true
	case int64:
		return x, true
	case uint64:
		return int64(x), true
	case float64:
		return int64(x), true
	default:
		return 0, false
	}
}
