// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"github.com/GuanceCloud/cliutils/pprofparser/domain/quantity"
	"github.com/GuanceCloud/cliutils/pprofparser/service/parsing"
	"github.com/google/pprof/profile"
)

func resolveJSONNumber(n json.Number) any {
	if !strings.Contains(n.String(), ".") {
		if x, err := n.Int64(); err == nil {
			return x
		}
	}
	if x, err := n.Float64(); err == nil {
		return x
	}
	return int64(0)
}

func rawMessage2String(message json.RawMessage) (string, error) {
	if message == nil {
		return "", nil
	}
	var s string
	if err := json.Unmarshal(message, &s); err != nil {
		return "", fmt.Errorf("illegal json string literal: %q", message)
	}
	return s, nil
}

func rawMessage2Number(message json.RawMessage) (json.Number, error) {
	var number json.Number
	if err := json.Unmarshal(message, &number); err != nil {
		return "", fmt.Errorf("illegal json number literal: %q", message)
	}
	return number, nil
}

func parseMetricsJSONFile(r io.Reader) (map[string]json.Number, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("unable to read metrics.json: %w", err)
	}

	var rawMetrics [][2]json.RawMessage

	if err = json.Unmarshal(body, &rawMetrics); err != nil {
		log.Errorf("unable to unmarshal metrics.json, %q: %v", string(body), err)
		return nil, fmt.Errorf("unable to unmarshal metrics.json: %w", err)
	}

	rm := make(map[string]json.Number)

	for _, numbers := range rawMetrics {
		name, err := rawMessage2String(numbers[0])
		if err != nil {
			return nil, fmt.Errorf("invalid metric key: %w", err)
		}
		num, err := rawMessage2Number(numbers[1])
		if err != nil {
			return nil, fmt.Errorf("invalid metric value: %w", err)
		}
		rm[name] = num
	}

	jsonMetrics := make(map[string]json.Number, len(rm))
	for name, jsonField := range goMetricsNameMapping {
		if v, ok := rm[jsonField]; ok {
			jsonMetrics[name] = v
		}
	}

	return jsonMetrics, nil
}

type pprofQuantity struct {
	Unit  string
	Value int64
}

func pprofCPUDuration(fileHeader *multipart.FileHeader) (durationNS int64, err error) {
	cpuMetrics, err := pprofSummaryHeader(fileHeader)
	if err != nil {
		return 0, fmt.Errorf("unable to resolve pprof file [%s] cpu metrics: %w", fileHeader.Filename, err)
	}

	if cpuDuration, ok := cpuMetrics["cpu"]; ok {
		unit, err := quantity.ParseUnit(quantity.Duration, cpuDuration.Unit)
		if err != nil {
			return 0, fmt.Errorf("unable to resolve cpu duration unit [%s]: %v", cpuDuration.Unit, cpuDuration.Value)
		}

		cpuNanos, err := unit.Quantity(cpuDuration.Value).IntValueIn(quantity.NanoSecond)
		if err != nil {
			return 0, fmt.Errorf("unable to convert cpu duration to nanoseconds: %w", err)
		}
		return cpuNanos, nil
	}
	return 0, fmt.Errorf("cpu profiling metrics not found")
}

func liveHeapSummary(fileHeader *multipart.FileHeader) (liveHeapObjects, liveHeapBytes int64, err error) {
	allocMetrics, err := pprofSummaryHeader(fileHeader)
	if err != nil {
		return 0, 0, fmt.Errorf("unable to resolve pprof file [%s] allocs metrics: %w", fileHeader.Filename, err)
	}

	if inuseObjects, ok := allocMetrics["inuse_objects"]; ok {
		liveHeapObjects = inuseObjects.Value
	}

	if inuseSpace, ok := allocMetrics["inuse_space"]; ok {
		unit, err := quantity.ParseUnit(quantity.Memory, inuseSpace.Unit)
		if err != nil {
			return liveHeapObjects, 0, fmt.Errorf("unable to resolve cpu duration unit [%s]: %v", inuseSpace.Unit, inuseSpace.Value)
		}

		liveHeapBytes, err = unit.Quantity(inuseSpace.Value).IntValueIn(quantity.Byte)
		if err != nil {
			return liveHeapObjects, liveHeapBytes, fmt.Errorf("unable to convert to inuse space value to Bytes: %w", err)
		}
	}
	return
}

func delayDurationNS(fileHeader *multipart.FileHeader) (durationNS int64, err error) {
	delayMetrics, err := pprofSummaryHeader(fileHeader)
	if err != nil {
		return 0, fmt.Errorf("unable to resolve go pprof block metrics: %w", err)
	}

	if delay, ok := delayMetrics["delay"]; ok {
		unit, err := quantity.ParseUnit(quantity.Duration, delay.Unit)
		if err != nil {
			return 0, fmt.Errorf("unable to resolve cpu duration unit [%s]: %v", delay.Unit, delay.Value)
		}
		durationNS, err = unit.Quantity(delay.Value).IntValueIn(quantity.NanoSecond)
		if err != nil {
			return 0, fmt.Errorf("unable to convert go pprof blocked duration to nanoseconds: %w", err)
		}
	}
	return
}

func goroutinesCount(fileHeader *multipart.FileHeader) (int64, error) {
	goroutineMetrics, err := pprofSummaryHeader(fileHeader)
	if err != nil {
		return 0, fmt.Errorf("unable to resolve goroutines count metrics: %w", err)
	}
	if goroutines, ok := goroutineMetrics["goroutines"]; ok {
		return goroutines.Value, nil
	}
	return 0, nil
}

func pprofSummaryHeader(mh *multipart.FileHeader) (map[string]*pprofQuantity, error) {
	if mh == nil {
		return nil, fmt.Errorf("nil FileHeader")
	}

	f, err := mh.Open()
	if err != nil {
		return nil, fmt.Errorf("unable to open file [%s]: %w", mh.Filename, err)
	}
	defer f.Close() //nolint:errcheck

	summaries, err := pprofSummary(f)
	if err != nil {
		return nil, fmt.Errorf("unable to parse pprof file [%s]: %w", mh.Filename, err)
	}
	return summaries, nil
}

func pprofSummary(r io.Reader) (map[string]*pprofQuantity, error) {
	prof, err := profile.Parse(parsing.NewDecompressor(r))
	if err != nil {
		return nil, fmt.Errorf("unable to parse pprof: %w", err)
	}

	summaries := make(map[string]*pprofQuantity, len(prof.SampleType))

	for _, valueType := range prof.SampleType {
		summaries[valueType.Type] = &pprofQuantity{
			Unit:  valueType.Unit,
			Value: 0,
		}
	}

	for _, sample := range prof.Sample {
		if len(sample.Value) != len(prof.SampleType) {
			return nil, fmt.Errorf("malformed pprof, SampleType count: %d, Value count: %d",
				len(prof.SampleType), len(sample.Value))
		}
		for idx, v := range sample.Value {
			summaries[prof.SampleType[idx].Type].Value += v
		}
	}

	return summaries, nil
}
