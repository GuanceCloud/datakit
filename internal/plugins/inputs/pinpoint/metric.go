// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pinpoint metrics.
package pinpoint

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	ppv1 "github.com/GuanceCloud/tracing-protos/pinpoint-gen-go/v1"
	"google.golang.org/grpc/metadata"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

func ParsePPAgentStatMessage(md metadata.MD, msg *ppv1.PStatMessage) {
	var pts []*point.Point
	statBatch := msg.GetAgentStatBatch()
	if statBatch != nil && statBatch.GetAgentStat() != nil {
		pts = statBatchToPoints(md, statBatch)
		if len(pts) > 0 && metricFeeder != nil {
			if err := metricFeeder.Feed(point.Metric, pts,
				dkio.WithSource(inputName),
			); err != nil {
				log.Errorf("feed metric to io err=%v", err)
			}
		}
	}

	uriStat := msg.GetAgentUriStat()
	if uriStat != nil {
		log.Debugf("uri stat=%+v", uriStat)
	}
}

func statBatchToPoints(md metadata.MD, batch *ppv1.PAgentStatBatch) (pts []*point.Point) {
	pts = make([]*point.Point, 0)
	agentID := "unknown"
	if vals := md.Get("agentid"); len(vals) > 0 {
		agentID = vals[0]
	}
	infoTags := map[string]string{}
	if agentCache != nil {
		info := agentCache.GetAgentInfo(agentID)
		if info != nil {
			infoTags = fromAgentTag(info)
		}
	}
	for _, stat := range batch.GetAgentStat() {
		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(time.UnixMilli(stat.GetTimestamp())))
		var statKV point.KVs
		cpuLoad := stat.GetCpuLoad()
		if cpuLoad != nil {
			statKV = statKV.Add("SystemCpuLoad", cpuLoad.GetSystemCpuLoad()).
				Add("JvmCpuLoad", cpuLoad.GetJvmCpuLoad()).
				AddTag("agent_id", agentID)
		}

		gc := stat.GetGc()
		if gc != nil {
			statKV = statKV.AddTag("agent_id", agentID).
				Add("JvmMemoryHeapUsed", gc.GetJvmMemoryHeapUsed()).
				Add("JvmMemoryHeapMax", gc.GetJvmMemoryHeapMax()).
				Add("JvmMemoryNonHeapUsed", gc.GetJvmMemoryNonHeapUsed()).
				Add("JvmMemoryNonHeapMax", gc.GetJvmMemoryNonHeapMax()).
				Add("JvmGcOldCount", gc.GetJvmGcOldCount()).
				Add("JvmGcOldTime", gc.GetJvmGcOldTime())

			if gc.GetJvmGcDetailed() != nil {
				detailed := gc.GetJvmGcDetailed()
				statKV = statKV.Add("GcNewCount", detailed.GetJvmGcNewCount()).
					Add("GcNewTime", detailed.GetJvmGcNewTime()).
					Add("PoolCodeCacheUsed", detailed.GetJvmPoolCodeCacheUsed()).
					Add("PoolNewGenUsed", detailed.GetJvmPoolNewGenUsed()).
					Add("PoolOldGenUsed", detailed.GetJvmPoolOldGenUsed()).
					Add("PoolSurvivorSpaceUsed", detailed.GetJvmPoolSurvivorSpaceUsed()).
					Add("PoolPermGenUsed", detailed.GetJvmPoolPermGenUsed()).
					Add("PoolMetaspaceUsed", detailed.GetJvmPoolMetaspaceUsed())
			}
		}
		for k, v := range infoTags {
			statKV = statKV.AddTag(k, v)
		}
		pt := point.NewPoint("pp-agentStats", statKV, opts...)
		pts = append(pts, pt)
	}

	return pts
}
