package promremote

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type Parser struct {
	MetricNameFilter  []string `toml:"metric_name_filter"`
	MeasurementPrefix string   `toml:"measurement_prefix"`
	MeasurementName   string   `toml:"measurement_name"`
}

// Parse parses given byte as protocol buffer. it performs necessary
// metric filtering and prefixing, and returns parsed measurements.
func (p *Parser) Parse(buf []byte) ([]inputs.Measurement, error) {
	var err error
	var metrics []inputs.Measurement
	var req prompb.WriteRequest
	if err := proto.Unmarshal(buf, &req); err != nil {
		return nil, fmt.Errorf("unable to unmarshal request body: %w", err)
	}
	now := time.Now()
	for _, ts := range req.Timeseries {
		tags := map[string]string{}

		for _, l := range ts.Labels {
			tags[l.Name] = l.Value
		}

		metricName := tags[model.MetricNameLabel]
		if metricName == "" {
			return nil, fmt.Errorf("metric name %q not found in tag-set or empty", model.MetricNameLabel)
		}
		delete(tags, model.MetricNameLabel)

		if !p.isValid(metricName) {
			continue
		}

		firstName, subName := p.getNames(metricName)

		for _, s := range ts.Samples {
			fields := make(map[string]interface{})
			if !math.IsNaN(s.Value) {
				fields[subName] = s.Value
			}
			// convert to measurement
			if len(fields) > 0 {
				t := now
				if s.Timestamp > 0 {
					t = time.Unix(0, s.Timestamp*1000000)
				}
				m := &Measurement{
					name:   firstName,
					tags:   tags,
					fields: fields,
					ts:     t,
				}
				metrics = append(metrics, m)
			}
		}
	}

	return metrics, err
}

func (p *Parser) getNames(metric string) (firstName, subName string) {
	if p.MeasurementName != "" {
		return p.MeasurementPrefix + p.MeasurementName, metric
	}
	// divide metric name by first '_' met
	index := strings.Index(metric, "_")
	firstName, subName = metric, metric
	if index != -1 {
		firstName = metric[:index]
		subName = metric[index+1:]
	}
	return p.MeasurementPrefix + firstName, subName
}
