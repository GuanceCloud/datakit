// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package awslambda

import (
	"time"

	"go.uber.org/zap"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/model"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/telemetry"
)

func (ipt *Input) toLogPoint(e *telemetry.LogEvent) *point.Point {
	if ipt.UseNowTimeInstead {
		e.Time = time.Now()
	}
	opts := point.DefaultLoggingOptions()
	opts = append(opts, point.WithTime(e.Time))
	kvs := append(point.NewTags(ipt.tags), point.NewKV(AWSLogFrom, e.Record.GetType(), point.WithKVTagSet(true)))
	kvs = append(kvs, point.NewKVs(e.Record.GetFields())...)
	return point.NewPointV2(inputName, kvs, opts...)
}

func (ipt *Input) toLogPointArr(es []*telemetry.LogEvent) (pts []*point.Point) {
	for _, v := range es {
		logPoint := ipt.toLogPoint(v)
		if l.Level() <= zap.DebugLevel {
			l.Debugf("logPoint: %s", logPoint.Pretty())
		}
		pts = append(pts, logPoint)
	}
	return
}

func (ipt *Input) toMetricPoint(e *telemetry.Event) (metric *point.Point) {
	if ipt.UseNowTimeInstead {
		e.Time = time.Now()
	}
	optsMetric := point.DefaultMetricOptions()
	optsMetric = append(optsMetric, point.WithTime(e.Time))
	var kvsMetric point.KVs
	switch e.Record.GetType() {
	case telemetry.TypePlatformInitStart:
		r := e.Record.(*telemetry.PlatformInitStart)
		if r.RuntimeVersion != "" {
			ipt.tags["aws_runtime_version"] = r.RuntimeVersion
		}
		if r.RuntimeVersionArn != "" {
			ipt.tags["aws_runtime_version_arn"] = r.RuntimeVersionArn
		}
	case telemetry.TypePlatformInitRuntimeDone:
	case telemetry.TypePlatformInitReport:
		r := e.Record.(*telemetry.PlatformInitReport)
		kvsMetric = append(kvsMetric, point.NewKV(metrics.InitDurationMetric, r.Metrics.DurationMs))
	case telemetry.TypePlatformRestoreStart:
		r := e.Record.(*telemetry.PlatformRestoreStart)
		if r.RuntimeVersion != "" {
			ipt.tags["aws_runtime_version"] = r.RuntimeVersion
		}
		if r.RuntimeVersionArn != "" {
			ipt.tags["aws_runtime_version_arn"] = r.RuntimeVersionArn
		}
	case telemetry.TypePlatformRestoreRuntimeDone:
	case telemetry.TypePlatformRestoreReport:
	case telemetry.TypePlatformStart:
		r := e.Record.(*telemetry.PlatformStart)
		if v, ok := ipt.lambdaCtxCache.Get(r.RequestID); ok {
			v.flag |= flagStart
			// kvsMetric = ipt.addPostRuntimeDurationMetric(v, kvsMetric, r.RequestID)
		} else {
			ipt.lambdaCtxCache.Set(r.RequestID, &lambdaCtx{
				flag: flagStart,
			})
		}
	case telemetry.TypePlatformRuntimeDone:
		if r, ok := e.Record.(*telemetry.PlatformRuntimeDone); ok {
			if r.Metrics != nil {
				if v, ok := ipt.lambdaCtxCache.Get(r.RequestID); ok {
					v.funDurationMs -= r.Metrics.DurationMs
					v.flag |= flagRuntimeDone
					kvsMetric = ipt.addPostRuntimeDurationMetric(v, kvsMetric, r.RequestID)
				} else {
					ipt.lambdaCtxCache.Set(r.RequestID, &lambdaCtx{
						funDurationMs: -r.Metrics.DurationMs,
						flag:          flagRuntimeDone,
					})
				}
				kvsMetric = append(kvsMetric, point.NewKV(metrics.RuntimeDurationMetric, r.Metrics.DurationMs))
				if r.Status != model.StatusSuccess {
					kvsMetric = append(kvsMetric, point.NewKV(metrics.ErrorsMetric, 1))
					if r.Status == model.StatusTimeout {
						kvsMetric = append(kvsMetric, point.NewKV(metrics.TimeoutsMetric, 1))
					}
				}
			}
		}

	case telemetry.TypePlatformReport:
		r := e.Record.(*telemetry.PlatformReport)
		if v, ok := ipt.lambdaCtxCache.Get(r.RequestID); ok {
			v.funDurationMs += r.Metrics.DurationMs
			v.flag |= flagReport
			kvsMetric = ipt.addPostRuntimeDurationMetric(v, kvsMetric, r.RequestID)
		} else {
			ipt.lambdaCtxCache.Set(r.RequestID, &lambdaCtx{
				funDurationMs: r.Metrics.DurationMs,
				flag:          flagReport,
			})
		}
		kvsMetric = append(kvsMetric,
			point.NewKV(metrics.DurationMetric, r.Metrics.DurationMs),
			point.NewKV(metrics.BilledDurationMetric, r.Metrics.BilledDurationMs),
			point.NewKV(metrics.MaxMemoryUsedMetric, r.Metrics.MaxMemoryUsedMB),
			point.NewKV(metrics.MemorySizeMetric, r.Metrics.MemorySizeMB),
		)
	case telemetry.TypePlatformExtension:
	case telemetry.TypePlatformTelemetrySubscription:
	case telemetry.TypePlatformLogsDropped:
	default:
	}
	kvsMetric = append(kvsMetric, point.NewTags(ipt.tags)...)
	return point.NewPointV2(inputName, kvsMetric, optsMetric...)
}

func (ipt *Input) addPostRuntimeDurationMetric(v *lambdaCtx, kvsMetric point.KVs, id string) point.KVs {
	if (v.flag | flagStart) == flagAll {
		kvsMetric = append(kvsMetric, point.NewKV(metrics.PostRuntimeDurationMetric, v.funDurationMs))
		ipt.lambdaCtxCache.Del(id)
	}
	return kvsMetric
}

func (ipt *Input) toMetricPointArr(es []*telemetry.Event) (metrics []*point.Point) {
	l.Debugf("convert %d event to point arr", len(es))

	for _, v := range es {
		l.Debugf("event: %+#v", v)

		metricPoint := ipt.toMetricPoint(v)
		if l.Level() <= zap.DebugLevel {
			l.Debugf("metricPoint: %s", metricPoint.Pretty())
		}

		metrics = append(metrics, metricPoint)
	}
	return
}
