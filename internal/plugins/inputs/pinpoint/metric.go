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
			if err := metricFeeder.FeedV2(point.Metric, pts,
				dkio.WithInputName(inputName),
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
			statKV = statKV.Add("SystemCpuLoad", cpuLoad.GetSystemCpuLoad(), false, false).
				Add("JvmCpuLoad", cpuLoad.GetJvmCpuLoad(), false, false).
				AddTag("agent_id", agentID)
		}

		gc := stat.GetGc()
		if gc != nil {
			statKV = statKV.AddTag("agent_id", agentID).
				Add("JvmMemoryHeapUsed", gc.GetJvmMemoryHeapUsed(), false, false).
				Add("JvmMemoryHeapMax", gc.GetJvmMemoryHeapMax(), false, false).
				Add("JvmMemoryNonHeapUsed", gc.GetJvmMemoryNonHeapUsed(), false, false).
				Add("JvmMemoryNonHeapMax", gc.GetJvmMemoryNonHeapMax(), false, false).
				Add("JvmGcOldCount", gc.GetJvmGcOldCount(), false, false).
				Add("JvmGcOldTime", gc.GetJvmGcOldTime(), false, false)

			if gc.GetJvmGcDetailed() != nil {
				detailed := gc.GetJvmGcDetailed()
				statKV = statKV.Add("GcNewCount", detailed.GetJvmGcNewCount(), false, false).
					Add("GcNewTime", detailed.GetJvmGcNewTime(), false, false).
					Add("PoolCodeCacheUsed", detailed.GetJvmPoolCodeCacheUsed(), false, false).
					Add("PoolNewGenUsed", detailed.GetJvmPoolNewGenUsed(), false, false).
					Add("PoolOldGenUsed", detailed.GetJvmPoolOldGenUsed(), false, false).
					Add("PoolSurvivorSpaceUsed", detailed.GetJvmPoolSurvivorSpaceUsed(), false, false).
					Add("PoolPermGenUsed", detailed.GetJvmPoolPermGenUsed(), false, false).
					Add("PoolMetaspaceUsed", detailed.GetJvmPoolMetaspaceUsed(), false, false)
			}
		}
		for k, v := range infoTags {
			statKV = statKV.AddTag(k, v)
		}
		pt := point.NewPointV2("pp-agentStats", statKV, opts...)
		pts = append(pts, pt)
	}

	return pts
}
