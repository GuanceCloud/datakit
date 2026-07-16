// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

import (
	"context"
	"encoding/json"
	"net"
	"strconv"
	"strings"
)

type routeItem struct {
	IP           string  `json:"ip"`
	ResponseTime float64 `json:"response_time"`
}

type routeHop struct {
	TTL     int         `json:"ttl,omitempty"`
	Total   int         `json:"total"`
	Failed  int         `json:"failed"`
	Loss    float64     `json:"loss"`
	AvgCost float64     `json:"avg_cost"`
	MinCost float64     `json:"min_cost"`
	MaxCost float64     `json:"max_cost"`
	StdCost float64     `json:"std_cost"`
	Items   []routeItem `json:"items"`
}

type traceroutePayload struct {
	Runs     []tracerouteRun    `json:"runs"`
	HopCount tracerouteHopCount `json:"hop_count"`
}

type tracerouteRun struct {
	RunID       string                 `json:"run_id"`
	Destination *tracerouteDestination `json:"destination,omitempty"`
	Hops        []tracerouteHop        `json:"hops"`
}

type tracerouteDestination struct {
	IPAddress  string   `json:"ip_address"`
	Port       uint16   `json:"port,omitempty"`
	ReverseDNS []string `json:"reverse_dns,omitempty"`
}

type tracerouteHop struct {
	TTL           int      `json:"ttl"`
	IPAddress     string   `json:"ip_address,omitempty"`
	ReverseDNS    []string `json:"reverse_dns,omitempty"`
	RTT           float64  `json:"rtt,omitempty"` // milliseconds, matching the Datadog network path payload
	Reachable     bool     `json:"reachable"`
	ASN           uint64   `json:"asn,omitempty"`
	ASName        string   `json:"as_name,omitempty"`
	ASPrefix      string   `json:"as_prefix,omitempty"`
	CloudProvider string   `json:"cloud_provider,omitempty"`
}

type tracerouteHopCount struct {
	Avg float64 `json:"avg"`
	Min int     `json:"min"`
	Max int     `json:"max"`
}

func newTracerouteDestination(host, ip string, port uint16) *tracerouteDestination {
	destination := &tracerouteDestination{IPAddress: ip, Port: port}
	if strings.TrimSpace(host) != "" && net.ParseIP(host) == nil {
		destination.ReverseDNS = []string{host}
	}
	return destination
}

func normalizeTracerouteFields(fields map[string]interface{}) {
	raw, _ := fields["traceroute"].(string)
	if strings.TrimSpace(raw) == "" || raw == "[]" {
		raw, _ = fields["hops_json"].(string)
	}
	normalized := normalizeTracerouteJSON(raw)
	if normalized == "" {
		return
	}
	fields["traceroute"] = normalized
	fields["hops_json"] = normalized
}

func normalizeTracerouteJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return ""
	}

	var payload traceroutePayload
	if err := json.Unmarshal([]byte(raw), &payload); err == nil && len(payload.Runs) > 0 {
		cleanTraceroutePayload(&payload)
		if b, err := json.Marshal(payload); err == nil {
			return string(b)
		}
		return raw
	}

	var routes []routeHop
	if err := json.Unmarshal([]byte(raw), &routes); err == nil {
		payload := routeHopsToPayload(routes)
		if len(payload.Runs) == 0 {
			return ""
		}
		if b, err := json.Marshal(payload); err == nil {
			return string(b)
		}
		return raw
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return ""
	}
	for _, key := range []string{"hops", "routes", "traceroute"} {
		if v, ok := obj[key]; ok {
			return normalizeTracerouteJSON(string(v))
		}
	}
	return raw
}

func routeHopsToPayload(routes []routeHop) traceroutePayload {
	maxRuns := 0
	for _, route := range routes {
		if len(route.Items) > maxRuns {
			maxRuns = len(route.Items)
		}
	}
	if maxRuns == 0 && len(routes) > 0 {
		maxRuns = 1
	}

	runs := make([]tracerouteRun, 0, maxRuns)
	for i := 0; i < maxRuns; i++ {
		runs = append(runs, tracerouteRun{RunID: strconv.Itoa(i + 1)})
	}

	for routeIdx, route := range routes {
		ttl := route.TTL
		if ttl <= 0 {
			ttl = routeIdx + 1
		}
		for runIdx := range runs {
			item := routeItem{IP: "*"}
			if runIdx < len(route.Items) {
				item = route.Items[runIdx]
			}
			runs[runIdx].Hops = append(runs[runIdx].Hops, routeItemToHop(ttl, item))
		}
	}

	return traceroutePayload{
		Runs:     runs,
		HopCount: buildHopCount(runs),
	}
}

func routeItemToHop(ttl int, item routeItem) tracerouteHop {
	ip := strings.TrimSpace(item.IP)
	hop := tracerouteHop{
		TTL:       ttl,
		Reachable: ip != "" && ip != "*",
	}
	if !hop.Reachable {
		return hop
	}

	hop.IPAddress = ip
	hop.RTT = item.ResponseTime / 1000
	return hop
}

func cleanTraceroutePayload(payload *traceroutePayload) {
	if payload == nil {
		return
	}
	for i := range payload.Runs {
		for j := range payload.Runs[i].Hops {
			hop := &payload.Runs[i].Hops[j]
			hop.Reachable = hop.Reachable && strings.TrimSpace(hop.IPAddress) != "" && hop.IPAddress != "*"
			if !hop.Reachable {
				hop.IPAddress = ""
				hop.ReverseDNS = nil
				hop.RTT = 0
				hop.ASN = 0
				hop.ASName = ""
				hop.ASPrefix = ""
				hop.CloudProvider = ""
				continue
			}
		}
	}
	if payload.HopCount == (tracerouteHopCount{}) {
		payload.HopCount = buildHopCount(payload.Runs)
	}
}

func enrichTracerouteFieldsContext(ctx context.Context, fields map[string]interface{}, rdns *reverseDNSEnricher) {
	if rdns == nil {
		return
	}
	raw, _ := fields["traceroute"].(string)
	if strings.TrimSpace(raw) == "" || raw == "[]" {
		raw, _ = fields["hops_json"].(string)
	}
	enriched := enrichTracerouteJSONContext(ctx, raw, rdns)
	if enriched == "" {
		return
	}
	fields["traceroute"] = enriched
	fields["hops_json"] = enriched
}

func enrichTracerouteJSONContext(ctx context.Context, raw string, rdns *reverseDNSEnricher) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return ""
	}

	var payload traceroutePayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil || len(payload.Runs) == 0 {
		return raw
	}

	lookupCtx, cancel := newReverseDNSEnrichmentContext(ctx)
	defer cancel()
	ips := make([]string, 0)
	for i := range payload.Runs {
		for j := range payload.Runs[i].Hops {
			hop := &payload.Runs[i].Hops[j]
			if !hop.Reachable || strings.TrimSpace(hop.IPAddress) == "" || len(hop.ReverseDNS) > 0 {
				continue
			}
			ips = append(ips, hop.IPAddress)
		}
	}
	names := rdns.lookupManyContext(lookupCtx, ips)
	for i := range payload.Runs {
		for j := range payload.Runs[i].Hops {
			hop := &payload.Runs[i].Hops[j]
			if len(hop.ReverseDNS) > 0 {
				continue
			}
			if name, ok := names[strings.TrimSpace(hop.IPAddress)]; ok {
				hop.ReverseDNS = []string{name}
			}
		}
	}
	if b, err := json.Marshal(payload); err == nil {
		return string(b)
	}
	return raw
}

func buildHopCount(runs []tracerouteRun) tracerouteHopCount {
	if len(runs) == 0 {
		return tracerouteHopCount{}
	}
	minCount := len(runs[0].Hops)
	maxCount := len(runs[0].Hops)
	total := 0
	for _, run := range runs {
		count := len(run.Hops)
		total += count
		if count < minCount {
			minCount = count
		}
		if count > maxCount {
			maxCount = count
		}
	}
	return tracerouteHopCount{
		Avg: float64(total) / float64(len(runs)),
		Min: minCount,
		Max: maxCount,
	}
}
