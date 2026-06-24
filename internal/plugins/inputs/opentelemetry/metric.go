// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"time"

	"github.com/GuanceCloud/cliutils/otlp"
	"github.com/GuanceCloud/cliutils/point"
	metrics "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/metrics/v1"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

func (ipt *Input) parseResourceMetricsV2(resmcs []*metrics.ResourceMetrics) {
	start := time.Now()
	stringOpts := otlp.DefaultMetricStringMapOptions()
	pts := otlp.ParseResourceMetricsV2(resmcs, otlp.MetricsParserOptions{
		ResourceStringOptions: stringOpts,
		ScopeStringOptions:    stringOpts,
		PointStringOptions:    stringOpts,
	})

	for len(pts) > 0 {
		batch := pts
		if len(batch) > 1000 {
			batch = pts[:1000]
		}

		if err := ipt.feeder.Feed(point.Metric, batch,
			dkio.WithSource(inputName), dkio.WithInput(inputName),
			dkio.DisableGlobalTags(ipt.TracingMetricDisableGlobalHostTags),
			dkio.WithCollectCost(time.Since(start)),
		); err != nil {
			log.Errorf("feed err=%v", err)
		}

		pts = pts[len(batch):]
	}
}
