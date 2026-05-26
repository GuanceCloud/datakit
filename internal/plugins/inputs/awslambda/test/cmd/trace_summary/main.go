// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	lambdatrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/trace"
)

type traceRecord struct {
	Timestamp string               `json:"timestamp"`
	Traces    [][]lambdatrace.Span `json:"traces"`
}

type pointSnapshotLine struct {
	Timestamp string                `json:"timestamp"`
	TestName  string                `json:"test"`
	Points    []pointSnapshotRecord `json:"points"`
}

type pointSnapshotRecord struct {
	TraceID    string `json:"trace_id"`
	SpanID     string `json:"span_id"`
	ParentID   string `json:"parent_id"`
	Operation  string `json:"operation"`
	Service    string `json:"service"`
	Resource   string `json:"resource"`
	Start      int64  `json:"start"`
	Duration   int64  `json:"duration"`
	Status     string `json:"status"`
	SpanType   string `json:"span_type"`
	Source     string `json:"source"`
	SourceType string `json:"source_type"`
	RequestID  string `json:"request_id"`
	Trigger    string `json:"trigger"`
	ColdStart  bool   `json:"cold_start"`
	InitType   string `json:"init_type"`
}

type traceSummary struct {
	TraceID      uint64
	FirstSeen    time.Time
	LastSeen     time.Time
	PayloadCount int
	Tests        []string
	RequestIDs   []string
	Triggers     []string
	SpanNames    []string
	ColdStart    bool
	Managed      bool
	SpansByID    map[uint64]lambdatrace.Span
}

func main() {
	defaultFile := filepath.Clean("./internal/plugins/inputs/awslambda/test/tmp/input-tracing-points.ndjson")
	filePath := flag.String("file", defaultFile, "path to received trace ndjson file")
	flag.Parse()

	files := existingFiles(
		filepath.Clean("./internal/plugins/inputs/awslambda/test/test.output/input-tracing-points.ndjson"),
		filepath.Clean("./internal/plugins/inputs/awslambda/test/test.output/input-ddtrace-points.ndjson"),
	)
	if flag.Lookup("file") != nil && *filePath != defaultFile {
		files = []string{*filePath}
	}
	summaries := map[uint64]*traceSummary{}
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "open trace file: no summary input file found\n")
		os.Exit(1)
	}

	for _, path := range files {
		// #nosec G304 // This is a test tool that processes known input files
		file, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "open trace file: %v\n", err)
			os.Exit(1)
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Bytes()
			if isPointSnapshotLine(line) {
				var record pointSnapshotLine
				if err := json.Unmarshal(line, &record); err != nil {
					_ = file.Close()
					fmt.Fprintf(os.Stderr, "decode input trace line: %v\n", err)
					os.Exit(1)
				}
				recordedAt := parseTimestamp(record.Timestamp)
				mergePointSummary(summaries, recordedAt, record.TestName, record.Points)
				continue
			}

			var record traceRecord
			if err := json.Unmarshal(line, &record); err != nil {
				_ = file.Close()
				fmt.Fprintf(os.Stderr, "decode line: %v\n", err)
				os.Exit(1)
			}
			recordedAt := parseTimestamp(record.Timestamp)
			for _, trace := range record.Traces {
				mergeTraceSummary(summaries, recordedAt, trace)
			}
		}
		if err := scanner.Err(); err != nil {
			_ = file.Close()
			fmt.Fprintf(os.Stderr, "scan trace file: %v\n", err)
			os.Exit(1)
		}
		_ = file.Close()
	}
	printSummaries(summaries)
}

func existingFiles(paths ...string) []string {
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			out = append(out, path)
		}
	}
	return out
}

func isPointSnapshotLine(line []byte) bool {
	return strings.Contains(string(line), `"points"`) && strings.Contains(string(line), `"trace_id"`)
}

func mergeTraceSummary(summaries map[uint64]*traceSummary, recordedAt time.Time, trace []lambdatrace.Span) {
	if len(trace) == 0 {
		return
	}
	traceID := trace[0].TraceID
	summary, ok := summaries[traceID]
	if !ok {
		summary = &traceSummary{
			TraceID:   traceID,
			FirstSeen: recordedAt,
			LastSeen:  recordedAt,
			SpansByID: map[uint64]lambdatrace.Span{},
		}
		summaries[traceID] = summary
	}
	if recordedAt.Before(summary.FirstSeen) {
		summary.FirstSeen = recordedAt
	}
	if recordedAt.After(summary.LastSeen) {
		summary.LastSeen = recordedAt
	}
	summary.PayloadCount++
	for _, span := range trace {
		appendUnique(&summary.RequestIDs, strings.TrimSpace(span.Meta["request_id"]))
		appendUnique(&summary.Triggers, strings.TrimSpace(span.Meta["trigger"]))
		appendUnique(&summary.SpanNames, strings.TrimSpace(span.Name))
		summary.SpansByID[span.SpanID] = span
		if span.Name == "aws.lambda.cold_start" || span.Meta["cold_start"] == "true" {
			summary.ColdStart = true
		}
		if span.Meta["init_type"] == "lambda-managed-instances" {
			summary.Managed = true
		}
	}
}

func mergePointSummary(summaries map[uint64]*traceSummary, recordedAt time.Time, testName string, points []pointSnapshotRecord) {
	if len(points) == 0 {
		return
	}

	traceID, _ := strconv.ParseUint(points[0].TraceID, 10, 64)
	summary, ok := summaries[traceID]
	if !ok {
		summary = &traceSummary{
			TraceID:   traceID,
			FirstSeen: recordedAt,
			LastSeen:  recordedAt,
			SpansByID: map[uint64]lambdatrace.Span{},
		}
		summaries[traceID] = summary
	}
	if recordedAt.Before(summary.FirstSeen) {
		summary.FirstSeen = recordedAt
	}
	if recordedAt.After(summary.LastSeen) {
		summary.LastSeen = recordedAt
	}
	summary.PayloadCount++
	appendUnique(&summary.Tests, strings.TrimSpace(testName))
	for _, record := range points {
		appendUnique(&summary.RequestIDs, strings.TrimSpace(record.RequestID))
		appendUnique(&summary.Triggers, strings.TrimSpace(record.Trigger))
		appendUnique(&summary.SpanNames, strings.TrimSpace(record.Operation))
		if record.ColdStart {
			summary.ColdStart = true
		}
		if record.InitType == "lambda-managed-instances" {
			summary.Managed = true
		}
		spanID, _ := strconv.ParseUint(record.SpanID, 10, 64)
		parentID, _ := strconv.ParseUint(record.ParentID, 10, 64)
		summary.SpansByID[spanID] = lambdatrace.Span{
			TraceID:  traceID,
			SpanID:   spanID,
			ParentID: parentID,
			Name:     record.Operation,
			Resource: record.Resource,
			Service:  record.Service,
			Start:    record.Start * int64(time.Microsecond),
			Duration: record.Duration * int64(time.Microsecond),
			Meta: map[string]string{
				"request_id": record.RequestID,
				"trigger":    record.Trigger,
				"init_type":  record.InitType,
			},
		}
	}
}

func printSummaries(summaries map[uint64]*traceSummary) {
	ordered := make([]*traceSummary, 0, len(summaries))
	for _, summary := range summaries {
		sort.Strings(summary.Tests)
		sort.Strings(summary.RequestIDs)
		sort.Strings(summary.Triggers)
		sort.Strings(summary.SpanNames)
		ordered = append(ordered, summary)
	}
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].FirstSeen.Before(ordered[j].FirstSeen)
	})

	for idx, summary := range ordered {
		fmt.Fprintf(os.Stdout, "trace #%d\n", idx+1)
		fmt.Fprintf(os.Stdout, "  trace_id: %d\n", summary.TraceID)
		fmt.Fprintf(os.Stdout, "  mode: %s\n", ternary(summary.Managed, "managed-instance", "on-demand"))
		fmt.Fprintf(os.Stdout, "  cold_start: %t\n", summary.ColdStart)
		fmt.Fprintf(os.Stdout, "  first_seen: %s\n", summary.FirstSeen.Format(time.RFC3339Nano))
		fmt.Fprintf(os.Stdout, "  last_seen: %s\n", summary.LastSeen.Format(time.RFC3339Nano))
		fmt.Fprintf(os.Stdout, "  payloads: %d\n", summary.PayloadCount)
		fmt.Fprintf(os.Stdout, "  tests: %s\n", joinOrDash(summary.Tests))
		fmt.Fprintf(os.Stdout, "  request_ids: %s\n", joinOrDash(summary.RequestIDs))
		fmt.Fprintf(os.Stdout, "  triggers: %s\n", joinOrDash(summary.Triggers))
		fmt.Fprintf(os.Stdout, "  spans: %s\n", joinOrDash(summary.SpanNames))
		fmt.Fprintf(os.Stdout, "  call_chain:\n")
		printCallChain(summary)
	}
}

func printCallChain(summary *traceSummary) {
	children := map[uint64][]lambdatrace.Span{}
	roots := make([]lambdatrace.Span, 0, len(summary.SpansByID))
	for _, span := range summary.SpansByID {
		if span.ParentID == 0 {
			roots = append(roots, span)
			continue
		}
		if _, ok := summary.SpansByID[span.ParentID]; !ok {
			roots = append(roots, span)
			continue
		}
		children[span.ParentID] = append(children[span.ParentID], span)
	}
	sortSpans(roots)
	for _, root := range roots {
		printSpanNode(root, children, 2)
	}
}

func printSpanNode(span lambdatrace.Span, children map[uint64][]lambdatrace.Span, depth int) {
	indent := strings.Repeat("  ", depth)
	fmt.Fprintf(os.Stdout, "%s- %s span_id=%d parent_id=%d service=%s duration=%s\n",
		indent, span.Name, span.SpanID, span.ParentID, emptyFallback(span.Service, "-"), formatDuration(span.Duration))
	kids := children[span.SpanID]
	sortSpans(kids)
	for _, child := range kids {
		printSpanNode(child, children, depth+1)
	}
}

func sortSpans(spans []lambdatrace.Span) {
	sort.Slice(spans, func(i, j int) bool {
		if spans[i].Start == spans[j].Start {
			if spans[i].Name == spans[j].Name {
				return spans[i].SpanID < spans[j].SpanID
			}
			return spans[i].Name < spans[j].Name
		}
		return spans[i].Start < spans[j].Start
	})
}

func appendUnique(values *[]string, value string) {
	if value == "" || contains(*values, value) {
		return
	}
	*values = append(*values, value)
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func joinOrDash(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ", ")
}

func parseTimestamp(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func formatDuration(ns int64) string {
	if ns <= 0 {
		return "0s"
	}
	d := time.Duration(ns)
	if d >= time.Millisecond {
		return strconv.FormatFloat(float64(d)/float64(time.Millisecond), 'f', 3, 64) + "ms"
	}
	if d >= time.Microsecond {
		return strconv.FormatFloat(float64(d)/float64(time.Microsecond), 'f', 3, 64) + "us"
	}
	return d.String()
}

func emptyFallback(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func ternary(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}
