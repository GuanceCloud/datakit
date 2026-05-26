// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kafka

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/jolokia"
)

type customObject struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

func (m *customObject) Point() *point.Point {
	opts := point.DefaultObjectOptions()

	if m.election {
		opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	}

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

// collectCustomerObjectMeasurement collects Kafka version and uptime as custom object.
func (ipt *Input) collectCustomerObjectMeasurement(client *jolokia.Client) ([]*point.Point, error) {
	clientURL := client.URL()
	uu, err := url.Parse(clientURL)
	if err != nil {
		l.Errorf("failed to parse client URL %s: %v", clientURL, err)
		return []*point.Point{}, nil
	}

	host, _, err := net.SplitHostPort(uu.Host)
	if err != nil {
		host = uu.Host
		l.Warnf("failed to split host and port: %s", err)
	}

	version, uptime, err := ipt.getKafkaVersionAndUptime(client)
	if err != nil {
		l.Errorf("failed to get kafka version and uptime: %s", err)
		return []*point.Point{}, nil
	}

	l.Debugf("kafka version:%s,uptime:%d", version, uptime)

	fields := map[string]interface{}{
		"display_name": host,
		"version":      version,
		"uptime":       fmt.Sprintf("%d", int(uptime)),
	}
	tags := map[string]string{
		"name":          fmt.Sprintf("%s-%s", inputName, host),
		"host":          host,
		"ip":            host,
		"col_co_status": "OK",
	}

	co := &customObject{
		name:     "mq",
		tags:     tags,
		fields:   fields,
		election: ipt.Election,
	}
	pt := co.Point()

	// Merge with input tags
	for k, v := range ipt.Tags {
		pt.AddTag(k, v)
	}
	return []*point.Point{pt}, nil
}

// getKafkaVersionAndUptime gets Kafka version and uptime from Jolokia.
func (ipt *Input) getKafkaVersionAndUptime(client *jolokia.Client) (version string, uptimeSeconds int64, err error) {
	// Create a read request for kafka.server:type=app-info
	req := client.NewReadRequest(kafkaAppInfoMBean, nil, "")
	response, err := client.Execute(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to execute request: %w", err)
	}

	if response.Status != 200 {
		return "", 0, fmt.Errorf("response has status code %d, expected 200", response.Status)
	}

	// Parse response value
	valueMap, ok := response.Value.(map[string]interface{})
	if !ok {
		return "", 0, fmt.Errorf("unexpected response value type")
	}

	versionVal, ok := valueMap["version"]
	if !ok {
		return "", 0, fmt.Errorf("version not found in response")
	}
	version, ok = versionVal.(string)
	if !ok {
		return "", 0, fmt.Errorf("version is not a string")
	}

	startTimeVal, ok := valueMap["start-time-ms"]
	if !ok {
		return "", 0, fmt.Errorf("start-time-ms not found in response")
	}

	var startTime int64
	switch v := startTimeVal.(type) {
	case json.Number:
		val, err := v.Int64()
		if err != nil {
			return "", 0, fmt.Errorf("failed to parse start-time-ms as int64: %w", err)
		}
		startTime = val
	case float64:
		startTime = int64(v)
	case int64:
		startTime = v
	case int:
		startTime = int64(v)
	default:
		return "", 0, fmt.Errorf("start-time-ms has unexpected type: %T", v)
	}

	uptimeMillis := time.Now().UnixMilli() - startTime
	uptimeSeconds = uptimeMillis / 1000

	return version, uptimeSeconds, nil
}
