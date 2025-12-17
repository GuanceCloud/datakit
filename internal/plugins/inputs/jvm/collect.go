// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jvm

import (
	"context"
	"encoding/json"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/jolokia"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

// collect collects metrics from all clients concurrently.
func (ipt *Input) collect(ptTS int64) error {
	for _, client := range ipt.clients {
		if client == nil {
			continue
		}

		client := client
		ipt.g.Go(func(gCtx context.Context) error {
			collectStart := time.Now()
			upState := 1

			// Normal collect mode
			jolokiaMetrics := ipt.convertToCommonMetrics(ipt.Metrics)
			requests := client.BuildJolokiaRequests(jolokiaMetrics)

			jsonRequests, _ := json.Marshal(requests)
			l.Debugf("BuildJolokiaRequests for %s: %s", client.URL(), string(jsonRequests))

			responses, err := client.BatchExecute(requests)
			if err != nil {
				l.Errorf("BatchExecute failed for %s: %v", client.URL(), err)
				upState = 0
			} else {
				points := ipt.convertResponses(jolokiaMetrics, responses, client.URL(), ptTS)
				if len(points) > 0 {
					if err := ipt.Feeder.Feed(point.Metric, points,
						dkio.WithCollectCost(time.Since(collectStart)),
						dkio.WithElection(ipt.Election),
						dkio.WithSource(inputName),
						dkio.WithMeasurement(inputs.GetOverrideMeasurement(ipt.MeasurementVersion, measurementJVM))); err != nil {
						l.Errorf("Feed failed for %s: %s, ignored", client.URL(), err.Error())
					}
				}

				// Record total collect duration for normal mode
				collectDurationVec.WithLabelValues(client.URL()).Observe(time.Since(collectStart).Seconds())
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

// convertResponses converts jolokia.Response to jvm measurement points.
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
