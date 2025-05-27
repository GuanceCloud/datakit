// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jolokia

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type gatherer struct {
	metrics  []Metric
	requests []readRequest
}

func newGatherer(metrics []Metric) *gatherer {
	return &gatherer{
		metrics:  metrics,
		requests: makeReadRequests(metrics),
	}
}

// gather adds points to an accumulator from responses returned
// by a Jolokia agent.
func (g *gatherer) gather(client *jclient, j *JolokiaAgent, ptTS int64) error {
	tags := make(map[string]string)
	if client.config.proxyCfg != nil {
		tags["jolokia_proxy_url"] = client.url
	} else {
		tags["jolokia_agent_url"] = client.url
	}

	requests := makeReadRequests(g.metrics)
	responses, err := client.read(requests)
	if err != nil {
		return err
	}

	return g.gatherResponses(responses, mergeTags(tags, j.Tags), j, ptTS)
}

// gatherResponses adds points to an accumulator from the ReadResponse objects
// returned by a Jolokia agent.
func (g *gatherer) gatherResponses(responses []readResponse, tags map[string]string, j *JolokiaAgent, ptTS int64) error {
	series := make(map[string][]jpoint)

	for _, metric := range g.metrics {
		points, ok := series[metric.Name]
		if !ok {
			points = make([]jpoint, 0)
		}

		responsePoints, responseErrors := g.generatePoints(metric, responses)
		points = append(points, responsePoints...)
		if len(responseErrors) != 0 {
			return responseErrors[0]
		}

		series[metric.Name] = points
	}

	for measurement, points := range series {
		for _, point := range compactPoints(points) {
			if len(point.Fields) == 0 {
				continue
			}

			field := make(map[string]interface{})
			for k, v := range point.Fields {
				if v == nil {
					continue
				}

				if jn, ok := v.(json.Number); ok {
					field[k] = convertJSONNumber(k, jn, j.Types)
				} else {
					field[k] = v
				}
			}

			if len(field) == 0 {
				continue
			}

			tag := mergeTags(point.Tags, tags)

			var hostURL string
			if len(tag["jolokia_agent_url"]) > 0 {
				hostURL = tag["jolokia_agent_url"]
			} else if len(tag["jolokia_proxy_url"]) > 0 {
				hostURL = tag["jolokia_proxy_url"]
			}
			if len(hostURL) > 0 {
				if host, err := getHostFromURL(hostURL); err != nil {
					return err
				} else {
					tag["host"] = host
				}
			}

			if j.Election {
				tag = inputs.MergeTagsWrapper(tag, j.Tagger.ElectionTags(), j.Tags, hostURL)
			} else {
				tag = inputs.MergeTagsWrapper(tag, j.Tagger.HostTags(), j.Tags, hostURL)
			}

			metric := &JolokiaMeasurement{
				name:   measurement,
				tags:   tag,
				fields: field,
				ts:     ptTS,
			}
			j.collectCache = append(j.collectCache, metric.Point())
		}
	}
	return nil
}

func getHostFromURL(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	return u.Host, nil
}

// generatePoints creates points for the supplied metric from the ReadResponse
// objects returned by the Jolokia client.
func (g *gatherer) generatePoints(metric Metric, responses []readResponse) ([]jpoint, []error) {
	points := make([]jpoint, 0)
	errors := make([]error, 0)

	for _, response := range responses {
		switch response.status {
		case 200:
			if !metricMatchesResponse(metric, response) {
				continue
			}

			pb := newPointBuilder(metric, response.requestAttributes, response.requestPath)
			for _, point := range pb.Build(metric.Mbean, response.value) {
				if response.requestTarget != "" {
					point.Tags["jolokia_agent_url"] = response.requestTarget
				}

				points = append(points, point)
			}

		case 404:
			continue
		default:
			errors = append(errors, fmt.Errorf("unexpected status in response from target %s (%q): %d",
				response.requestTarget, response.requestMbean, response.status))
			continue
		}
	}

	return points, errors
}

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

func convertSpecifyType(fieldName string, jn json.Number, fieldTyp map[string]string) interface{} {
	val, _ := jn.Float64()
	if typ, ok := fieldTyp[fieldName]; ok && typ == "int" {
		return int64(val)
	}
	return val
}

// makeReadRequests creates ReadRequest objects from metrics definitions.
func makeReadRequests(metrics []Metric) []readRequest {
	var requests []readRequest
	for _, metric := range metrics {
		if len(metric.Paths) == 0 {
			requests = append(requests, readRequest{
				mbean:      metric.Mbean,
				attributes: []string{},
			})
		} else {
			attributes := make(map[string][]string)

			for _, path := range metric.Paths {
				segments := strings.Split(path, "/")
				attribute := segments[0]

				if _, ok := attributes[attribute]; !ok {
					attributes[attribute] = make([]string, 0)
				}

				if len(segments) > 1 {
					paths := attributes[attribute]
					attributes[attribute] = append(paths, strings.Join(segments[1:], "/"))
				}
			}
			rootAttributes := findRequestAttributesWithoutPaths(attributes)
			if len(rootAttributes) > 0 {
				requests = append(requests, readRequest{
					mbean:      metric.Mbean,
					attributes: rootAttributes,
				})
			}

			for _, deepAttribute := range findRequestAttributesWithPaths(attributes) {
				for _, path := range attributes[deepAttribute] {
					requests = append(requests, readRequest{
						mbean:      metric.Mbean,
						attributes: []string{deepAttribute},
						path:       path,
					})
				}
			}
		}
	}
	return requests
}

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

// metricMatchesResponse returns true when the name, attributes, and path
// of a Metric match the corresponding elements in a ReadResponse object
// returned by a Jolokia agent.
func metricMatchesResponse(metric Metric, response readResponse) bool {
	if !metric.MatchObjectName(response.requestMbean) {
		return false
	}

	if len(metric.Paths) == 0 {
		return len(response.requestAttributes) == 0
	}

	for _, attribute := range response.requestAttributes {
		if metric.MatchAttributeAndPath(attribute, response.requestPath) {
			return true
		}
	}

	return false
}

// mergeTags combines two tag sets into a single tag set.
func mergeTags(metricTags, outerTags map[string]string) map[string]string {
	tags := make(map[string]string)
	for k, v := range outerTags {
		tags[k] = v
	}
	for k, v := range metricTags {
		tags[k] = v
	}

	return tags
}
