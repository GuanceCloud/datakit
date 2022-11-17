// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"fmt"
	"strings"
)

// EvaluatedSampleDependencies set of supported memory usage metrics.
var EvaluatedSampleDependencies = map[string]bool{
	"memory.usage": true,
	"memory.used":  true,
	"memory.total": true,
	"memory.free":  true,
}

func tryReportMemoryUsage(scalarSamples map[string]MetricSample, columnSamples map[string]map[string]MetricSample, outData *MetricDatas) error {
	if hasScalarMemoryUsage(scalarSamples) {
		return trySendScalarMemoryUsage(scalarSamples, outData)
	}

	return trySendColumnMemoryUsage(columnSamples, outData)
}

// hasScalarMemoryUsage is true if samples store contains at least one memory usage related metric.
func hasScalarMemoryUsage(scalarSamples map[string]MetricSample) bool {
	_, scalarMemoryUsageOk := scalarSamples["memory.usage"]
	_, scalarMemoryUsedOk := scalarSamples["memory.used"]
	_, scalarMemoryTotalOk := scalarSamples["memory.total"]
	_, scalarMemoryFreeOk := scalarSamples["memory.free"]

	return scalarMemoryUsageOk || scalarMemoryUsedOk || scalarMemoryTotalOk || scalarMemoryFreeOk
}

func trySendScalarMemoryUsage(scalarSamples map[string]MetricSample, outData *MetricDatas) error {
	_, scalarMemoryUsageOk := scalarSamples["memory.usage"]
	if scalarMemoryUsageOk {
		// memory usage is already sent through collected metrics
		return nil
	}

	scalarMemoryUsed, scalarMemoryUsedOk := scalarSamples["memory.used"]
	scalarMemoryTotal, scalarMemoryTotalOk := scalarSamples["memory.total"]
	scalarMemoryFree, scalarMemoryFreeOk := scalarSamples["memory.free"]

	if scalarMemoryUsedOk {
		floatMemoryUsed, err := scalarMemoryUsed.value.ToFloat64()
		if err != nil {
			return fmt.Errorf("metric `%s`: failed to convert to float64: %w", "memory.used", err)
		}

		if scalarMemoryTotalOk {
			// memory total and memory used
			floatMemoryTotal, err := scalarMemoryTotal.value.ToFloat64()
			if err != nil {
				return fmt.Errorf("metric `%s`: failed to convert to float64: %w", "memory.total", err)
			}

			memoryUsageValue, err := evaluateMemoryUsage(floatMemoryUsed, floatMemoryTotal)
			if err != nil {
				return err
			}

			memoryUsageSample := MetricSample{
				value:      ResultValue{Value: memoryUsageValue},
				tags:       scalarMemoryUsed.tags,
				symbol:     SymbolConfig{Name: "memory.usage"},
				options:    MetricsConfigOption{},
				forcedType: "",
			}

			sendMetric(memoryUsageSample, outData)
			return nil
		}

		if scalarMemoryFreeOk {
			// memory total and memory used
			floatMemoryFree, err := scalarMemoryFree.value.ToFloat64()
			if err != nil {
				l.Debugf("metric `%s`: failed to convert to float64: %v", "memory.free", err)
				return err
			}

			floatMemoryUsed, err := scalarMemoryUsed.value.ToFloat64()
			if err != nil {
				l.Debugf("metric `%s`: failed to convert to float64: %v", "memory.used", err)
				return err
			}

			memoryUsageValue, err := evaluateMemoryUsage(floatMemoryUsed, floatMemoryFree+floatMemoryUsed)
			if err != nil {
				return err
			}

			memoryUsageSample := MetricSample{
				value:      ResultValue{Value: memoryUsageValue},
				tags:       scalarMemoryUsed.tags,
				symbol:     SymbolConfig{Name: "memory.usage"},
				options:    MetricsConfigOption{},
				forcedType: "",
			}

			sendMetric(memoryUsageSample, outData)
			return nil
		}
	}

	if scalarMemoryFreeOk && scalarMemoryTotalOk {
		// memory total and memory used
		floatMemoryFree, err := scalarMemoryFree.value.ToFloat64()
		if err != nil {
			l.Debugf("metric `%s`: failed to convert to float64: %v", "memory.free", err)
			return err
		}

		floatMemoryTotal, err := scalarMemoryTotal.value.ToFloat64()
		if err != nil {
			l.Debugf("metric `%s`: failed to convert to float64: %v", "memory.total", err)
			return err
		}

		memoryUsageValue, err := evaluateMemoryUsage(floatMemoryTotal-floatMemoryFree, floatMemoryTotal)
		if err != nil {
			return err
		}

		memoryUsageSample := MetricSample{
			value:      ResultValue{Value: memoryUsageValue},
			tags:       scalarMemoryTotal.tags,
			symbol:     SymbolConfig{Name: "memory.usage"},
			options:    MetricsConfigOption{},
			forcedType: "",
		}

		sendMetric(memoryUsageSample, outData)
		return nil
	}

	// report missing dependency metrics
	missingMetrics := []string{}
	if !scalarMemoryUsedOk {
		missingMetrics = append(missingMetrics, "used")
	}
	if !scalarMemoryFreeOk {
		missingMetrics = append(missingMetrics, "free")
	}
	if !scalarMemoryTotalOk {
		missingMetrics = append(missingMetrics, "total")
	}

	return fmt.Errorf("missing %s memory metrics, skipping scalar memory usage", strings.Join(missingMetrics, ", "))
}

func trySendColumnMemoryUsage(columnSamples map[string]map[string]MetricSample, outData *MetricDatas) error {
	_, memoryUsageOk := columnSamples["memory.usage"]
	if memoryUsageOk {
		// memory usage is already sent through collected metrics
		return nil
	}

	memoryUsedRows, memoryUsedOk := columnSamples["memory.used"]
	memoryTotalRows, memoryTotalOk := columnSamples["memory.total"]
	memoryFreeRows, memoryFreeOk := columnSamples["memory.free"]

	if memoryUsedOk {
		if memoryTotalOk {
			for rowIndex, memoryUsedSample := range memoryUsedRows {
				memoryTotalSample, memoryTotalSampleOk := memoryTotalRows[rowIndex]
				if !memoryTotalSampleOk {
					return fmt.Errorf("missing memory total sample at row %s, skipping memory usage evaluation", rowIndex)
				}
				floatMemoryTotal, err := memoryTotalSample.value.ToFloat64()
				if err != nil {
					return fmt.Errorf("metric `%s[%s]`: failed to convert to float64: %w", "memory.total", rowIndex, err)
				}

				floatMemoryUsed, err := memoryUsedSample.value.ToFloat64()
				if err != nil {
					return fmt.Errorf("metric `%s[%s]`: failed to convert to float64: %w", "memory.used", rowIndex, err)
				}

				memoryUsageValue, err := evaluateMemoryUsage(floatMemoryUsed, floatMemoryTotal)
				if err != nil {
					return err
				}

				memoryUsageSample := MetricSample{
					value:      ResultValue{Value: memoryUsageValue},
					tags:       memoryUsedSample.tags,
					symbol:     SymbolConfig{Name: "memory.usage"},
					options:    MetricsConfigOption{},
					forcedType: "",
				}

				sendMetric(memoryUsageSample, outData)
			}
			return nil
		}

		if memoryFreeOk {
			for rowIndex, memoryUsedSample := range memoryUsedRows {
				memoryFreeSample, memoryFreeSampleOk := memoryFreeRows[rowIndex]
				if !memoryFreeSampleOk {
					return fmt.Errorf("missing memory free sample at row %s, skipping memory usage evaluation", rowIndex)
				}
				floatMemoryFree, err := memoryFreeSample.value.ToFloat64()
				if err != nil {
					return fmt.Errorf("metric `%s[%s]`: failed to convert to float64: %w", "memory.free", rowIndex, err)
				}

				floatMemoryUsed, err := memoryUsedSample.value.ToFloat64()
				if err != nil {
					return fmt.Errorf("metric `%s[%s]`: failed to convert to float64: %w", "memory.used", rowIndex, err)
				}

				memoryUsageValue, err := evaluateMemoryUsage(floatMemoryUsed, floatMemoryFree+floatMemoryUsed)
				if err != nil {
					return err
				}

				memoryUsageSample := MetricSample{
					value:      ResultValue{Value: memoryUsageValue},
					tags:       memoryUsedSample.tags,
					symbol:     SymbolConfig{Name: "memory.usage"},
					options:    MetricsConfigOption{},
					forcedType: "",
				}

				sendMetric(memoryUsageSample, outData)
			}
			return nil
		}
	}

	if memoryFreeOk && memoryTotalOk {
		for rowIndex, memoryTotalSample := range memoryTotalRows {
			memoryFreeSample, memoryFreeSampleOk := memoryFreeRows[rowIndex]
			if !memoryFreeSampleOk {
				return fmt.Errorf("missing memory free sample at row %s, skipping memory usage evaluation", rowIndex)
			}
			floatMemoryFree, err := memoryFreeSample.value.ToFloat64()
			if err != nil {
				return fmt.Errorf("metric `%s[%s]`: failed to convert to float64: %w", "memory.free", rowIndex, err)
			}

			floatMemoryTotal, err := memoryTotalSample.value.ToFloat64()
			if err != nil {
				return fmt.Errorf("metric `%s[%s]`: failed to convert to float64: %w", "memory.total", rowIndex, err)
			}

			memoryUsageValue, err := evaluateMemoryUsage(floatMemoryTotal-floatMemoryFree, floatMemoryTotal)
			if err != nil {
				return err
			}

			memoryUsageSample := MetricSample{
				value:      ResultValue{Value: memoryUsageValue},
				tags:       memoryTotalSample.tags,
				symbol:     SymbolConfig{Name: "memory.usage"},
				options:    MetricsConfigOption{},
				forcedType: "",
			}

			sendMetric(memoryUsageSample, outData)
		}
		return nil
	}

	// report missing dependency metrics
	missingMetrics := []string{}
	if !memoryUsedOk {
		missingMetrics = append(missingMetrics, "used")
	}
	if !memoryFreeOk {
		missingMetrics = append(missingMetrics, "free")
	}
	if !memoryTotalOk {
		missingMetrics = append(missingMetrics, "total")
	}

	return fmt.Errorf("missing %s memory metrics, skipping column memory usage", strings.Join(missingMetrics, ", "))
}

func evaluateMemoryUsage(memoryUsed float64, memoryTotal float64) (float64, error) {
	if memoryTotal == 0 {
		return 0, fmt.Errorf("cannot evaluate memory usage, total memory is 0")
	}
	if memoryUsed < 0 {
		return 0, fmt.Errorf("cannot evaluate memory usage, memory used is < 0")
	}
	return (memoryUsed / memoryTotal) * 100, nil
}
