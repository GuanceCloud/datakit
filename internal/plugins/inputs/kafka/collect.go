// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/jolokia"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

// Gather collects metrics from all clients concurrently.
func (ipt *Input) collect(ptTS int64) error {
	for _, client := range ipt.clients {
		if client == nil {
			continue
		}

		client := client
		ipt.g.Go(func(gCtx context.Context) error {
			// Collect customer object measurement (Kafka version and uptime)
			coCollectStart := time.Now()
			pts, err := ipt.collectCustomerObjectMeasurement(client)
			if err != nil {
				l.Errorf("Collect customer object failed for %s: %s, ignored", client.URL(), err.Error())
			}
			if len(pts) > 0 {
				if err := ipt.Feeder.Feed(point.CustomObject, pts,
					dkio.WithCollectCost(time.Since(coCollectStart)),
					dkio.WithElection(ipt.Election),
					dkio.WithSource(dkio.FeedSource(inputName, "CO"))); err != nil {
					l.Errorf("Feed customer object failed for %s: %s, ignored", client.URL(), err.Error())
				}
			}

			collectStart := time.Now()
			upState := 1
			// Auto collect mode
			if ipt.EnableAutoCollect {
				if err := ipt.autoCollect(client, ptTS); err != nil {
					upState = 0
					l.Errorf("Auto collect failed for %s: %v", client.URL(), err)
				} else {
					// Record total collect duration for auto mode
					collectDurationVec.WithLabelValues(client.URL(), "auto").Observe(time.Since(collectStart).Seconds())
				}
			} else {
				// Normal collect mode
				jolokiaMetrics := ipt.convertToCommonMetrics(ipt.Metrics)
				requests := client.BuildJolokiaRequests(jolokiaMetrics)

				jsonRequests, _ := json.Marshal(requests)
				l.Debugf("BuildJolokiaRequests for %s: %s", client.URL(), string(jsonRequests))

				responses, err := client.BatchExecute(requests)
				if err != nil {
					upState = 0
					l.Errorf("BatchExecute failed for %s: %v", client.URL(), err)
				} else {
					points := ipt.convertResponses(jolokiaMetrics, responses, client.URL(), ptTS)
					if len(points) > 0 {
						if err := ipt.Feeder.Feed(point.Metric,
							points,
							dkio.WithCollectCost(time.Since(collectStart)),
							dkio.WithElection(ipt.Election),
							dkio.WithSource(inputName)); err != nil {
							l.Errorf("Feed failed for %s: %s, ignored", client.URL(), err.Error())
						}
					}

					// Record total collect duration for normal mode
					collectDurationVec.WithLabelValues(client.URL(), "normal").Observe(time.Since(collectStart).Seconds())
				}
			}

			// Feed up metric
			ipt.feedUpMetric(client, upState)

			return nil
		})
	}

	if err := ipt.g.Wait(); err != nil {
		l.Errorf("collect failed: %s", err.Error())
		return err
	}

	return nil
}

// convertResponses converts jolokia.Response to kafka measurement points.
func (ipt *Input) convertResponses(
	jolokiaMetrics []jolokia.MetricConfig,
	responses []*jolokia.Response,
	clientURL string,
	ptTS int64,
) []*point.Point {
	converter := &jolokia.ConverterConfig{
		Types:      ipt.Types,
		GlobalTags: ipt.Tags,
		Election:   ipt.Election,
		Tagger:     ipt.Tagger,
		L:          l,
		ClientURL:  clientURL,
		PtTS:       ptTS,
		Metrics:    jolokiaMetrics,
		Responses:  responses,
	}
	return converter.ConvertResponses()
}

func (ipt *Input) convertToCommonMetrics(metrics []MetricConfig) []jolokia.MetricConfig {
	commonMetrics := make([]jolokia.MetricConfig, 0, len(metrics))
	for _, metric := range metrics {
		tagPrefix := metric.TagPrefix
		if tagPrefix == nil && ipt.DefaultTagPrefix != "" {
			tagPrefix = &ipt.DefaultTagPrefix
		}

		fieldPrefix := metric.FieldPrefix
		if fieldPrefix == nil && ipt.DefaultFieldPrefix != "" {
			fieldPrefix = &ipt.DefaultFieldPrefix
		}

		fieldSeparator := metric.FieldSeparator
		if fieldSeparator == nil && ipt.DefaultFieldSeparator != "" {
			fieldSeparator = &ipt.DefaultFieldSeparator
		}

		commonMetrics = append(commonMetrics, jolokia.MetricConfig{
			Name:           metric.Name,
			Mbean:          metric.Mbean,
			Paths:          metric.Paths,
			FieldName:      metric.FieldName,
			FieldPrefix:    fieldPrefix,
			FieldSeparator: fieldSeparator,
			TagPrefix:      tagPrefix,
			TagKeys:        metric.TagKeys,
		})
	}
	return commonMetrics
}

// parseClientURL parses client URL and returns the host/instance.
// If parsing fails, returns the original URL as instance.
func parseClientURL(clientURL string) string {
	uu, err := url.Parse(clientURL)
	if err != nil {
		l.Warnf("failed to parse client URL %s: %v", clientURL, err)
		return clientURL
	}
	return uu.Host
}

// autoCollect automatically collects all Kafka MBeans using search interface.
func (ipt *Input) autoCollect(client *jolokia.Client, ptTS int64) error {
	// Search for all MBeans
	searchStart := time.Now()
	searchReq := client.NewSearchRequest(defaultMBeanPattern)
	searchResp, err := client.Execute(searchReq)
	if err != nil {
		return err
	}

	// Record search MBean duration
	searchMBeanDurationVec.WithLabelValues(client.URL()).Observe(time.Since(searchStart).Seconds())

	if searchResp.Status != 200 {
		return fmt.Errorf("search request failed with status %d: %s", searchResp.Status, searchResp.Error)
	}

	// Parse search response to get MBean names
	var mbeanNames []string
	switch v := searchResp.Value.(type) {
	case []string:
		mbeanNames = v
	case []interface{}:
		mbeanNames = make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				mbeanNames = append(mbeanNames, str)
			} else {
				return fmt.Errorf("unexpected item type in search response: %T, expected string", item)
			}
		}
	default:
		return fmt.Errorf("unexpected search response type: %T, expected []string or []interface{}", searchResp.Value)
	}

	// Filter MBeans by blacklist
	originalCount := len(mbeanNames)
	mbeanNames = ipt.filterMBeansByBlacklist(mbeanNames)
	if len(mbeanNames) == 0 {
		l.Debugf("No MBeans found for %s (after filtering)", client.URL())
		return nil
	}

	if len(mbeanNames) < originalCount {
		l.Debugf("Filtered %d MBeans by blacklist, %d remaining for %s", originalCount-len(mbeanNames), len(mbeanNames), client.URL())
	} else {
		l.Debugf("Found %d MBeans for %s", len(mbeanNames), client.URL())
	}

	clientURL := client.URL()
	instance := parseClientURL(clientURL)
	return ipt.processInBatches(mbeanNames, defaultBatchSize, func(batch []string, startIdx int) error {
		batchCollectStart := time.Now()
		// Build read requests for this batch
		readRequests := make([]*jolokia.Request, 0, len(batch))
		for _, mbeanName := range batch {
			readReq := client.NewReadRequest(mbeanName, nil, "")
			readRequests = append(readRequests, readReq)
		}

		// Execute read requests for this batch
		responses, err := client.BatchExecute(readRequests)
		if err != nil {
			return fmt.Errorf("batch execute read requests failed: %w", err)
		}

		// Convert responses to points
		if len(responses) != len(batch) {
			l.Warnf("Response count %d != batch size %d, skipping", len(responses), len(batch))
			return nil
		}
		batchPoints := make([]*point.Point, 0, len(responses))
		for j, resp := range responses {
			mbeanName := batch[j]

			if resp == nil {
				l.Warnf("Skip response with nil for MBean %s", mbeanName)
				continue
			}
			if resp.Status != 200 {
				l.Warnf("Skip response with status %d for MBean %s", resp.Status, mbeanName)
				continue
			}

			pt := ipt.convertMBeanToPoint(mbeanName, resp.Value, clientURL, instance, ptTS)
			if pt != nil {
				batchPoints = append(batchPoints, pt)
			}
		}

		if len(batchPoints) > 0 {
			if err := ipt.Feeder.Feed(point.Metric,
				batchPoints,
				dkio.WithCollectCost(time.Since(batchCollectStart)),
				dkio.WithElection(ipt.Election),
				dkio.WithSource(inputName)); err != nil {
				l.Errorf("feed batch %d-%d failed: %v", startIdx+1, startIdx+len(batch), err)
			}
			l.Debugf("Fed batch %d-%d (%d points) successfully", startIdx+1, startIdx+len(batch), len(batchPoints))
		}

		return nil
	})
}

// convertMBeanToPoint converts a single MBean response to a point.
func (ipt *Input) convertMBeanToPoint(mbeanName string, value interface{}, clientURL, instance string, ptTS int64) *point.Point {
	// Parse MBean name to extract domain and properties
	domain, props := parseMbeanName(mbeanName)

	// Build tags from properties, excluding blacklisted ones
	tags := make(map[string]string)
	for key, val := range props {
		tagKey := strings.ReplaceAll(key, "-", "_")
		// Skip blacklisted tags
		if _, ok := ipt.TagBlackMap[tagKey]; ok {
			continue
		}
		tags[tagKey] = val
	}

	// Add jolokia_agent_url tag
	tags["jolokia_agent_url"] = clientURL
	if instance != "" {
		tags["host"] = instance
	}

	if ipt.Election {
		tags = inputs.MergeTagsWrapper(tags, ipt.Tagger.ElectionTags(), ipt.Tags, clientURL)
	} else {
		tags = inputs.MergeTagsWrapper(tags, ipt.Tagger.HostTags(), ipt.Tags, clientURL)
	}

	// Extract fields from response value with domain.type.name.attr format
	fields := extractFieldsFromValue(value, domain, props)
	if len(fields) == 0 {
		l.Debugf("No fields extracted from MBean %s", mbeanName)
		return nil
	}

	// Create point
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(ptTS))

	return point.NewPoint(inputName,
		append(point.NewTags(tags), point.NewKVs(fields)...),
		opts...)
}

// parseMbeanName parses MBean object name and returns domain and properties.
func parseMbeanName(mbeanName string) (string, map[string]string) {
	props := make(map[string]string)
	index := strings.Index(mbeanName, ":")
	if index == -1 {
		// No colon, the whole string is domain
		return mbeanName, props
	}

	domain := mbeanName[:index]
	if index+1 >= len(mbeanName) {
		return domain, props
	}

	propertyStr := mbeanName[index+1:]
	for _, keyProperty := range strings.Split(propertyStr, ",") {
		pair := strings.SplitN(keyProperty, "=", 2)
		if len(pair) == 2 {
			key := strings.TrimSpace(pair[0])
			value := strings.TrimSpace(pair[1])
			// Normalize tag value: replace special characters that might cause issues
			value = normalizeValue(value)
			if key != "" {
				props[key] = value
			}
		}
	}

	return domain, props
}

func normalizeValue(value string) string {
	return strings.ReplaceAll(value, "-", "_")
}

// extractFieldsFromValue extracts fields from MBean attribute values.
// Field keys are formatted as: domain.type.name.attr.
func extractFieldsFromValue(value interface{}, domain string, props map[string]string) map[string]interface{} {
	fields := make(map[string]interface{})

	valueMap, ok := value.(map[string]interface{})
	if !ok {
		// If not a map, wrap it as "value" with domain.type.name format
		if converted := convertValue(value); converted != nil {
			fieldKey := buildFieldKey(domain, props, "value")
			fields[fieldKey] = converted
		} else {
			l.Debugf("field value is nil for MBean %v", value)
		}
		return fields
	}

	for attr, val := range valueMap {
		// Skip arrays
		if _, isArray := val.([]interface{}); isArray {
			l.Debugf("Skip array value for attr %s", attr)
			continue
		}

		// Build field key: domain.type.name.attr
		fieldKey := buildFieldKey(domain, props, attr)

		if converted := convertValue(val); converted != nil {
			fields[fieldKey] = converted
		} else {
			l.Debugf("field value is nil for attr %s: %v", attr, val)
		}
	}

	return fields
}

// buildFieldKey builds field key in format: domain.type.name.attr.
func buildFieldKey(domain string, props map[string]string, attr string) string {
	parts := make([]string, 0, 4)

	// Add domain
	if domain != "" {
		parts = append(parts, domain)
	}

	// Add type (if exists)
	if typeVal, ok := props["type"]; ok && typeVal != "" {
		parts = append(parts, typeVal)
	}

	// Add name (if exists)
	if nameVal, ok := props["name"]; ok && nameVal != "" {
		parts = append(parts, nameVal)
	}

	// Add attribute name
	if attr != "" {
		normalizedAttr := normalizeValue(attr)
		parts = append(parts, normalizedAttr)
	}

	return strings.Join(parts, ".")
}

// convertValue converts value to appropriate type.
func convertValue(value interface{}) interface{} {
	jn, ok := value.(json.Number)
	if !ok {
		// Not a json.Number, skip it (string, bool, map, array, etc.)
		return nil
	}
	if intVal, err := jn.Int64(); err == nil {
		return intVal
	}
	if floatVal, err := jn.Float64(); err == nil {
		return floatVal
	}
	// json.Number that can't be converted to int or float is invalid
	return nil
}

// processInBatches processes items in batches using a generic batch processing function.
func (ipt *Input) processInBatches(items []string, batchSize int, processor func([]string, int) error) error {
	start := 0
	for start < len(items) {
		select {
		case <-datakit.Exit.Wait():
			return fmt.Errorf("datakit exiting")

		case <-ipt.SemStop.Wait():
			return fmt.Errorf("kafka input return")
		default:
		}

		end := start + batchSize
		if end > len(items) {
			end = len(items)
		}

		batch := items[start:end]
		l.Debugf("Processing batch %d-%d (total: %d)", start+1, end, len(items))

		if err := processor(batch, start); err != nil {
			return fmt.Errorf("process batch %d-%d failed: %w", start+1, end, err)
		}
		start += batchSize
	}

	return nil
}

// filterMBeansByBlacklist filters MBean names based on blacklist patterns.
func (ipt *Input) filterMBeansByBlacklist(mbeanNames []string) []string {
	if len(ipt.MBeanBlacklist) == 0 {
		return mbeanNames
	}

	filtered := make([]string, 0, len(mbeanNames))
	for _, mbeanName := range mbeanNames {
		if !ipt.isMBeanBlacklisted(mbeanName) {
			filtered = append(filtered, mbeanName)
		}
	}

	return filtered
}

// isMBeanBlacklisted checks if an MBean name matches any pattern in the blacklist.
func (ipt *Input) isMBeanBlacklisted(mbeanName string) bool {
	for _, pattern := range ipt.MBeanBlacklist {
		matched, err := filepath.Match(pattern, mbeanName)
		if err != nil {
			// Skip invalid patterns silently
			continue
		}
		if matched {
			return true
		}
	}
	return false
}
