// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package jolokia provides conversion utilities for Jolokia responses to data points.
package jolokia

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type ConverterConfig struct {
	Types      map[string]string
	GlobalTags map[string]string
	Election   bool
	Tagger     datakit.GlobalTagger
	ClientURL  string
	Metrics    []MetricConfig
	Responses  []*Response
	PtTS       int64
	L          *logger.Logger
}

type MetricConfig struct {
	Name           string
	Mbean          string
	Paths          []string
	FieldName      *string
	FieldPrefix    *string
	FieldSeparator *string
	TagPrefix      *string
	TagKeys        []string
}

// BuildJolokiaRequests converts MetricConfig to jolokia.Request list.
func (c *Client) BuildJolokiaRequests(metrics []MetricConfig) []*Request {
	var requests []*Request

	for _, metric := range metrics {
		if len(metric.Paths) == 0 {
			// Read all attributes
			req := c.NewReadRequest(metric.Mbean, nil, "")
			requests = append(requests, req)
		} else {
			// Process paths
			attributes := make(map[string][]string)
			for _, path := range metric.Paths {
				segments := strings.Split(path, "/")
				attribute := segments[0]

				if _, ok := attributes[attribute]; !ok {
					attributes[attribute] = make([]string, 0)
				}

				if len(segments) > 1 {
					attributes[attribute] = append(attributes[attribute], strings.Join(segments[1:], "/"))
				}
			}

			// Build requests for attributes without paths
			rootAttributes := findRequestAttributesWithoutPaths(attributes)
			if len(rootAttributes) > 0 {
				req := c.NewReadRequest(metric.Mbean, rootAttributes, "")
				requests = append(requests, req)
			}

			// Build requests for attributes with paths
			for _, deepAttribute := range findRequestAttributesWithPaths(attributes) {
				for _, path := range attributes[deepAttribute] {
					req := c.NewReadRequest(metric.Mbean, []string{deepAttribute}, path)
					requests = append(requests, req)
				}
			}
		}
	}

	return requests
}

func (config *ConverterConfig) ConvertResponses() []*point.Point {
	series := make(map[string][]jpoint)

	// Group responses by metric name
	for _, metric := range config.Metrics {
		points := series[metric.Name]

		responsePoints, responseErrors := generatePoints(metric, config.Responses, config.ClientURL)
		points = append(points, responsePoints...)
		if len(responseErrors) != 0 {
			config.L.Warnf("Errors generating points for metric %s: %v", metric.Name, responseErrors)
		}

		series[metric.Name] = points
	}

	// Convert jpoint to point.Point
	result := make([]*point.Point, 0)
	for measurement, points := range series {
		for _, p := range compactPoints(points) {
			if len(p.Fields) == 0 {
				continue
			}

			// Convert fields, handling json.Number
			fields := make(map[string]interface{})
			for k, v := range p.Fields {
				if v == nil {
					continue
				}

				if jn, ok := v.(json.Number); ok {
					fields[k] = convertJSONNumber(k, jn, config.Types)
				} else {
					fields[k] = v
				}
			}

			if len(fields) == 0 {
				continue
			}

			// Merge tags (similar to old code's mergeTags)
			tags := make(map[string]string)
			for k, v := range p.Tags {
				tags[k] = v
			}

			hostURL := tags["jolokia_agent_url"]
			if hostURL == "" {
				hostURL = tags["jolokia_proxy_url"]
			}
			if hostURL != "" {
				if host, err := getHostFromURL(hostURL); err != nil {
					config.L.Warnf("failed to get host from URL %s: %v", hostURL, err)
				} else {
					tags["host"] = host
				}
			}

			// Merge with global tags
			if config.Election {
				tags = inputs.MergeTagsWrapper(tags, config.Tagger.ElectionTags(), config.GlobalTags, hostURL)
			} else {
				tags = inputs.MergeTagsWrapper(tags, config.Tagger.HostTags(), config.GlobalTags, hostURL)
			}

			// Create point
			opts := point.DefaultMetricOptions()
			opts = append(opts, point.WithTimestamp(config.PtTS))

			result = append(result, point.NewPoint(measurement,
				append(point.NewTags(tags), point.NewKVs(fields)...),
				opts...))
		}
	}

	return result
}

// generatePoints creates points for the supplied metric from the jolokia.Response objects.
func generatePoints(metric MetricConfig, responses []*Response, clientURL string) ([]jpoint, []error) {
	points := make([]jpoint, 0)
	errors := make([]error, 0)

	for _, response := range responses {
		if response == nil {
			continue
		}

		if response.Request == nil {
			errors = append(errors, fmt.Errorf("request is nil for %s: %v", clientURL, response))
			continue
		}

		// Filter responses that don't match this metric's MBean pattern early
		// This prevents processing and reporting errors for unrelated responses
		if !matchObjectName(response.Request.Mbean, metric.Mbean) {
			continue
		}

		switch response.Status {
		case 200:
			// Extract attributes and path from request
			respPath := response.Request.Path
			var respAttributes []string
			switch v := response.Request.Attribute.(type) {
			case []string:
				respAttributes = v
			case string:
				respAttributes = []string{v}
			case []interface{}:
				// JSON decoding may return []interface{} instead of []string
				respAttributes = make([]string, 0, len(v))
				for _, item := range v {
					if str, ok := item.(string); ok {
						respAttributes = append(respAttributes, str)
					}
				}
			}

			if !metricMatchesResponse(metric, respAttributes, respPath) {
				continue
			}

			pb := newPointBuilder(metric, respAttributes, respPath)
			for _, p := range pb.Build(metric.Mbean, response.Value) {
				if response.Request.Target != nil {
					p.Tags["jolokia_agent_url"] = response.Request.Target.URL
					p.Tags["jolokia_proxy_url"] = clientURL
				} else {
					p.Tags["jolokia_agent_url"] = clientURL
				}
				points = append(points, p)
			}
		default:
			// Report error for this response (already matched MBean pattern above)
			errors = append(errors, fmt.Errorf("unexpected status in response from %s (%q): %d",
				clientURL, response.Request.Mbean, response.Status))
			continue
		}
	}

	return points, errors
}

// findRequestAttributesWithoutPaths returns attributes that have no nested paths.
func findRequestAttributesWithoutPaths(attributes map[string][]string) []string {
	results := make([]string, 0)
	for attr, paths := range attributes {
		if len(paths) == 0 {
			results = append(results, attr)
		}
	}

	sort.Strings(results)
	return results
}

// findRequestAttributesWithPaths returns attributes that have nested paths.
func findRequestAttributesWithPaths(attributes map[string][]string) []string {
	results := make([]string, 0)
	for attr, paths := range attributes {
		if len(paths) != 0 {
			results = append(results, attr)
		}
	}

	sort.Strings(results)
	return results
}

// metricMatchesResponse returns true when the metric matches the response.
func metricMatchesResponse(metric MetricConfig, respAttributes []string, respPath string) bool {
	if len(metric.Paths) == 0 {
		return len(respAttributes) == 0
	}

	// Check if any path matches
	for _, attribute := range respAttributes {
		if matchAttributeAndPath(attribute, respPath, metric.Paths) {
			return true
		}
	}

	return false
}

func matchObjectName(respMbean, metricMbean string) bool {
	if respMbean == metricMbean {
		return true
	}

	respDomain, respProps := parseMbeanObjectName(respMbean)
	metricDomain, metricProps := parseMbeanObjectName(metricMbean)

	if respDomain != metricDomain {
		return false
	}

	if len(respProps) != len(metricProps) {
		return false
	}

	// Build property map for fast lookup
	propSet := make(map[string]struct{}, len(respProps))
	for _, p := range respProps {
		propSet[p] = struct{}{}
	}

	// Check if metricProps all exist in the set
	for _, p := range metricProps {
		if _, ok := propSet[p]; !ok {
			return false
		}
	}

	return true
}

// parseMbeanObjectName parses an MBean object name into domain and properties.
func parseMbeanObjectName(name string) (string, []string) {
	index := strings.Index(name, ":")
	if index == -1 {
		return name, []string{}
	}

	domain := name[:index]

	if index+1 > len(name) {
		return domain, []string{}
	}

	return domain, strings.Split(name[index+1:], ",")
}

// matchAttributeAndPath checks if attribute and path match any of the metric paths.
func matchAttributeAndPath(attribute, innerPath string, metricPaths []string) bool {
	path := attribute
	if innerPath != "" {
		path = path + "/" + innerPath
	}

	for i := range metricPaths {
		if path == metricPaths[i] {
			return true
		}
	}

	return false
}

// convertJSONNumber converts json.Number to appropriate type.
func convertJSONNumber(fieldName string, jn json.Number, fieldTyp map[string]string) interface{} {
	if fieldTyp != nil {
		return convertSpecifyType(fieldName, jn, fieldTyp)
	}

	if intVal, err := jn.Int64(); err == nil {
		return intVal
	}

	if floatVal, err := jn.Float64(); err == nil {
		return floatVal
	}

	return jn.String()
}

// convertSpecifyType converts json.Number to specified type.
func convertSpecifyType(fieldName string, jn json.Number, fieldTyp map[string]string) interface{} {
	val, _ := jn.Float64()
	if typ, ok := fieldTyp[fieldName]; ok && typ == "int" {
		return int64(val)
	}
	return val
}

// getHostFromURL extracts host from URL string.
func getHostFromURL(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	return u.Host, nil
}

// jpoint represents an intermediate point structure.
type jpoint struct {
	Tags   map[string]string
	Fields map[string]interface{}
}

// pointBuilder builds points from metric configuration and response values.
type pointBuilder struct {
	metric           MetricConfig
	objectAttributes []string
	objectPath       string
	substitutions    []string
}

// newPointBuilder creates a new pointBuilder.
func newPointBuilder(metric MetricConfig, attributes []string, path string) *pointBuilder {
	return &pointBuilder{
		metric:           metric,
		objectAttributes: attributes,
		objectPath:       path,
		substitutions:    makeSubstitutionList(metric.Mbean),
	}
}

// Build generates points for a given mbean name/pattern and value object.
func (pb *pointBuilder) Build(mbean string, value interface{}) []jpoint {
	hasPattern := strings.Contains(mbean, "*")
	if !hasPattern {
		value = map[string]interface{}{mbean: value}
	}

	valueMap, ok := value.(map[string]interface{})
	if !ok {
		return nil
	}

	points := make([]jpoint, 0)
	for mbean, value := range valueMap {
		points = append(points, jpoint{
			Tags:   pb.extractTags(mbean),
			Fields: pb.extractFields(mbean, value),
		})
	}

	return compactPoints(points)
}

// extractTags generates the map of tags for a given mbean name/pattern.
func (pb *pointBuilder) extractTags(mbean string) map[string]string {
	propertyMap := makePropertyMap(mbean, pb.metric.Mbean)
	tagMap := make(map[string]string)

	for key, value := range propertyMap {
		if pb.includeTag(key) {
			tagName := pb.formatTagName(key)
			tagName = strings.ReplaceAll(tagName, "-", "_")
			tagMap[tagName] = value
		}
	}

	return tagMap
}

func (pb *pointBuilder) includeTag(tagName string) bool {
	for _, t := range pb.metric.TagKeys {
		if tagName == t {
			return true
		}
	}
	return false
}

func (pb *pointBuilder) formatTagName(tagName string) string {
	if tagName == "" {
		return ""
	}

	if pb.metric.TagPrefix != nil && *pb.metric.TagPrefix != "" {
		return *pb.metric.TagPrefix + tagName
	}

	return tagName
}

// extractFields generates the map of fields for a given mbean name and value object.
func (pb *pointBuilder) extractFields(mbean string, value interface{}) map[string]interface{} {
	fieldMap := make(map[string]interface{})
	valueMap, ok := value.(map[string]interface{})

	if ok {
		// complex value
		switch {
		case len(pb.objectAttributes) == 0:
			// if there were no attributes requested, then the keys are attributes
			pb.fillFields("", valueMap, fieldMap)
		case len(pb.objectAttributes) == 1:
			// if there was a single attribute requested, then the keys are the attribute's properties
			fieldName := pb.formatFieldName(pb.objectAttributes[0], pb.objectPath)
			pb.fillFields(fieldName, valueMap, fieldMap)
		default:
			// if there were multiple attributes requested, then the keys are the attribute names
			for _, attribute := range pb.objectAttributes {
				fieldName := pb.formatFieldName(attribute, pb.objectPath)
				if attrValue, ok := valueMap[attribute]; ok {
					pb.fillFields(fieldName, attrValue, fieldMap)
				}
			}
		}
	} else {
		// scalar value
		var fieldName string
		if len(pb.objectAttributes) == 0 {
			fieldName = pb.formatFieldName("value", pb.objectPath)
		} else {
			fieldName = pb.formatFieldName(pb.objectAttributes[0], pb.objectPath)
		}

		pb.fillFields(fieldName, value, fieldMap)
	}

	if len(pb.substitutions) > 1 {
		pb.applySubstitutions(mbean, fieldMap)
	}

	return fieldMap
}

// formatFieldName generates a field name from the supplied attribute and path.
func (pb *pointBuilder) formatFieldName(attribute, path string) string {
	fieldName := attribute
	fieldPrefix := ""
	fieldSeparator := ""

	if pb.metric.FieldPrefix != nil {
		fieldPrefix = *pb.metric.FieldPrefix
	}
	if pb.metric.FieldSeparator != nil {
		fieldSeparator = *pb.metric.FieldSeparator
	}

	if fieldPrefix != "" {
		fieldName = fieldPrefix + fieldName
	}

	if path != "" {
		fieldName = fieldName + fieldSeparator + strings.ReplaceAll(path, "/", fieldSeparator)
	}

	return fieldName
}

// fillFields recurses into the supplied value object, generating a named field for every value it discovers.
func (pb *pointBuilder) fillFields(name string, value interface{}, fieldMap map[string]interface{}) {
	fieldPrefix := ""
	if pb.metric.FieldPrefix != nil {
		fieldPrefix = *pb.metric.FieldPrefix
	}
	fieldSeparator := ""
	if pb.metric.FieldSeparator != nil {
		fieldSeparator = *pb.metric.FieldSeparator
	}

	if valueMap, ok := value.(map[string]interface{}); ok {
		// keep going until we get to something that is not a map
		for key, innerValue := range valueMap {
			if _, ok := innerValue.([]interface{}); ok {
				continue
			}

			var innerName string
			if name == "" {
				innerName = fieldPrefix + key
			} else {
				innerName = name + fieldSeparator + key
			}

			pb.fillFields(innerName, innerValue, fieldMap)
		}

		return
	}

	if _, ok := value.([]interface{}); ok {
		return
	}

	if pb.metric.FieldName != nil && *pb.metric.FieldName != "" {
		name = *pb.metric.FieldName
		if fieldPrefix != "" {
			name = fieldPrefix + name
		}
	}

	if name == "" {
		name = "value"
	}
	name = strings.ReplaceAll(name, "-", "_")
	fieldMap[name] = value
}

// applySubstitutions updates all the keys in the supplied map of fields to account for $1-style substitution instructions.
func (pb *pointBuilder) applySubstitutions(mbean string, fieldMap map[string]interface{}) {
	properties := makePropertyMap(mbean, pb.metric.Mbean)

	for i, subKey := range pb.substitutions[1:] {
		symbol := fmt.Sprintf("$%d", i+1)
		substitution := properties[subKey]

		for fieldName, fieldValue := range fieldMap {
			newFieldName := strings.ReplaceAll(fieldName, symbol, substitution)
			if fieldName != newFieldName {
				fieldMap[newFieldName] = fieldValue
				delete(fieldMap, fieldName)
			}
		}
	}
}

// makePropertyMap returns the mbean property-key list as a dictionary.
func makePropertyMap(mbean, metricMbean string) map[string]string {
	props := make(map[string]string)
	object := strings.SplitN(mbean, ":", 2)
	domain := object[0]

	// If metricMbean contains wildcards, use regex matching for tag values
	metricDomain, metricRegexMap := makeTagValueRegexMap(metricMbean)
	if metricDomain != domain { // domain should equal
		metricRegexMap = make(map[string]*regexp.Regexp)
	}

	if domain != "" && len(object) == 2 {
		list := object[1]

		for _, keyProperty := range strings.Split(list, ",") {
			pair := strings.SplitN(keyProperty, "=", 2)

			if len(pair) != 2 {
				continue
			}

			if key := pair[0]; key != "" {
				props[key] = pair[1]

				// Extract first submatch if regex pattern exists
				if v, ok := metricRegexMap[key]; ok && v != nil {
					match := v.FindAllStringSubmatch(pair[1], -1)
					if len(match) >= 1 && len(match[0]) > 1 {
						props[key] = match[0][1]
					}
				}
			}
		}
	}

	return props
}

// makeTagValueRegexMap replaces * in mbean with (.*) and compiles regex patterns.
func makeTagValueRegexMap(mbean string) (string, map[string]*regexp.Regexp) {
	subs := make(map[string]*regexp.Regexp)
	object := strings.SplitN(mbean, ":", 2)
	domain := object[0]
	if domain != "" && len(object) == 2 {
		list := object[1]
		for _, keyProperty := range strings.Split(list, ",") {
			pair := strings.SplitN(keyProperty, "=", 2)
			if len(pair) == 2 && pair[0] != "" {
				// default nil
				subs[pair[0]] = nil
				property := pair[1]
				if strings.Contains(property, "*") {
					property = strings.ReplaceAll(property, "*", "(.*)")
					if r, err := regexp.Compile(property); err == nil {
						// if successful
						subs[pair[0]] = r
					}
				}
			}
		}
	}
	return domain, subs
}

func makeSubstitutionList(mbean string) []string {
	subs := make([]string, 0)

	object := strings.SplitN(mbean, ":", 2)
	domain := object[0]

	if domain != "" && len(object) == 2 {
		subs = append(subs, domain)
		list := object[1]

		for _, keyProperty := range strings.Split(list, ",") {
			pair := strings.SplitN(keyProperty, "=", 2)

			if len(pair) != 2 {
				continue
			}

			key := pair[0]
			if key == "" {
				continue
			}

			property := pair[1]
			if !strings.Contains(property, "*") {
				continue
			}

			subs = append(subs, key)
		}
	}

	return subs
}

// compactPoints attempts to remove points by compacting points with matching tag sets.
func compactPoints(points []jpoint) []jpoint {
	if len(points) == 0 {
		return points
	}

	// Use map to quickly find points with matching tag sets
	tagKeyToPoint := make(map[string]*jpoint, len(points))

	for i := range points {
		sourcePoint := &points[i]
		tagKey := makeTagKey(sourcePoint.Tags)

		if existingPoint, exists := tagKeyToPoint[tagKey]; exists {
			// Merge fields from sourcePoint into existingPoint
			for key, val := range sourcePoint.Fields {
				existingPoint.Fields[key] = val
			}
		} else {
			// First point with this tag set, store it
			tagKeyToPoint[tagKey] = sourcePoint
		}
	}

	// Convert map values back to slice
	compactedPoints := make([]jpoint, 0, len(tagKeyToPoint))
	for _, p := range tagKeyToPoint {
		compactedPoints = append(compactedPoints, *p)
	}

	return compactedPoints
}

// makeTagKey generates a unique string key for a tag map by sorting keys and concatenating.
func makeTagKey(tags map[string]string) string {
	if len(tags) == 0 {
		return ""
	}

	// Sort keys for consistent key generation
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.Grow(len(keys) * 16) // Pre-allocate reasonable capacity
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('\x00')
		}
		b.WriteString(k)
		b.WriteByte('\x00')
		b.WriteString(tags[k])
	}

	return b.String()
}
