// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package skywalking

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	skyimpl "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/v3/compile"
	"google.golang.org/grpc"
)

func registerServerV3(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorf("skywalking grpc server v3 listening on %s failed: %v", addr, err)

		return
	}
	log.Infof("skywalking grpc v3 listening on: %s", addr)

	skySvr = grpc.NewServer()
	skyimpl.RegisterTraceSegmentReportServiceServer(skySvr, &TraceReportServerV3{})
	skyimpl.RegisterEventServiceServer(skySvr, &EventServerV3{})
	skyimpl.RegisterJVMMetricReportServiceServer(skySvr, &JVMMetricReportServerV3{})
	skyimpl.RegisterManagementServiceServer(skySvr, &ManagementServerV3{})
	skyimpl.RegisterConfigurationDiscoveryServiceServer(skySvr, &DiscoveryServerV3{})

	if err = skySvr.Serve(listener); err != nil {
		log.Error(err.Error())
	}
	log.Info("skywalking v3 exits")
}

type TraceReportServerV3 struct {
	skyimpl.UnimplementedTraceSegmentReportServiceServer
}

func (trsvr *TraceReportServerV3) Collect(tsc skyimpl.TraceSegmentReportService_CollectServer) error {
	for {
		segobj, err := tsc.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return tsc.SendAndClose(&skyimpl.Commands{})
			}
			log.Error(err.Error())

			return err
		}

		log.Debug("v3 segment received")

		if dktrace := segobjToDkTrace(segobj); len(dktrace) == 0 {
			log.Warn("empty datakit trace")
		} else {
			afterGatherRun.Run(inputName, itrace.DatakitTraces{dktrace}, false)
		}
	}
}

func (*TraceReportServerV3) CollectInSync(ctx context.Context, seg *skyimpl.SegmentCollection) (*skyimpl.Commands, error) {
	log.Debugf("reveived collect insync: %s", seg.String())

	return &skyimpl.Commands{}, nil
}

func segobjToDkTrace(segment *skyimpl.SegmentObject) itrace.DatakitTrace {
	var dktrace itrace.DatakitTrace
	for _, span := range segment.Spans {
		if span == nil {
			continue
		}

		dkspan := &itrace.DatakitSpan{
			TraceID:    segment.TraceId,
			SpanID:     fmt.Sprintf("%s%d", segment.TraceSegmentId, span.SpanId),
			Service:    segment.Service,
			Resource:   span.OperationName,
			Operation:  span.OperationName,
			Source:     inputName,
			SourceType: itrace.SPAN_SOURCE_CUSTOMER,
			EndPoint:   span.Peer,
			Start:      span.StartTime * int64(time.Millisecond),
			Duration:   (span.EndTime - span.StartTime) * int64(time.Millisecond),
		}

		if span.ParentSpanId < 0 {
			if len(span.Refs) > 0 {
				dkspan.ParentID = fmt.Sprintf("%s%d", span.Refs[0].ParentTraceSegmentId, span.Refs[0].ParentSpanId)
			} else {
				dkspan.ParentID = "0"
			}
		} else {
			if len(span.Refs) > 0 {
				dkspan.ParentID = fmt.Sprintf("%s%d", span.Refs[0].ParentTraceSegmentId, span.Refs[0].ParentSpanId)
			} else {
				dkspan.ParentID = fmt.Sprintf("%s%d", segment.TraceSegmentId, span.ParentSpanId)
			}
		}

		dkspan.Status = itrace.STATUS_OK
		if span.IsError {
			dkspan.Status = itrace.STATUS_ERR
		}

		switch span.SpanType {
		case skyimpl.SpanType_Entry:
			dkspan.SpanType = itrace.SPAN_TYPE_ENTRY
		case skyimpl.SpanType_Local:
			dkspan.SpanType = itrace.SPAN_TYPE_LOCAL
		case skyimpl.SpanType_Exit:
			dkspan.SpanType = itrace.SPAN_TYPE_EXIT
		default:
			dkspan.SpanType = itrace.SPAN_TYPE_ENTRY
		}

		sourceTags := make(map[string]string)
		for _, tag := range span.Tags {
			sourceTags[tag.Key] = tag.Value
		}
		dkspan.Tags = itrace.MergeInToCustomerTags(customerKeys, tags, sourceTags)

		if buf, err := json.Marshal(span); err != nil {
			log.Warn(err.Error())
		} else {
			dkspan.Content = string(buf)
		}

		dktrace = append(dktrace, dkspan)
	}
	if len(dktrace) != 0 {
		dktrace[0].Metrics = make(map[string]interface{})
		dktrace[0].Metrics[itrace.FIELD_PRIORITY] = itrace.PRIORITY_AUTO_KEEP
	}

	return dktrace
}

type EventServerV3 struct {
	skyimpl.UnimplementedEventServiceServer
}

func (*EventServerV3) Collect(esrv skyimpl.EventService_CollectServer) error {
	for {
		event, err := esrv.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return esrv.SendAndClose(&skyimpl.Commands{})
			}
			log.Debug(err.Error())

			return err
		}

		log.Debugf("reveived service event: %s", event.String())
	}
}

type jvmMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *jvmMeasurement) LineProto() (*dkio.Point, error) {
	return dkio.NewPoint(m.name, m.tags, m.fields, inputs.OptMetric)
}

func (m *jvmMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{}
}

func extraceJVMCpuMetric(service string, start time.Time, cpu *skyimpl.CPU) inputs.Measurement {
	return &jvmMeasurement{
		name:   jvmMetricName,
		tags:   map[string]string{"service": service},
		fields: map[string]interface{}{"cpu_usage_percent": cpu.UsagePercent},
		ts:     start,
	}
}

func extractJVMMemoryMetrics(service string, start time.Time, memory []*skyimpl.Memory) []inputs.Measurement {
	isHeap := func(isHeap bool) string {
		if isHeap {
			return "heap"
		} else {
			return "stack"
		}
	}

	var m []inputs.Measurement
	for _, v := range memory {
		if v == nil {
			continue
		}
		m = append(m, &jvmMeasurement{
			name: jvmMetricName,
			tags: map[string]string{"service": service},
			fields: map[string]interface{}{
				isHeap(v.IsHeap) + "_init":      v.Init,
				isHeap(v.IsHeap) + "_max":       v.Max,
				isHeap(v.IsHeap) + "_used":      v.Used,
				isHeap(v.IsHeap) + "_committed": v.Committed,
			},
			ts: start,
		})
	}

	return m
}

func extractJVMMemoryPoolMetrics(service string, start time.Time, memoryPool []*skyimpl.MemoryPool) []inputs.Measurement {
	var m []inputs.Measurement
	for _, v := range memoryPool {
		if v == nil {
			continue
		}
		m = append(m, &jvmMeasurement{
			name: jvmMetricName,
			tags: map[string]string{"service": service},
			fields: map[string]interface{}{
				"pool_" + strings.ToLower(skyimpl.PoolType_name[int32(v.Type)]) + "_init":      v.Init,
				"pool_" + strings.ToLower(skyimpl.PoolType_name[int32(v.Type)]) + "_max":       v.Max,
				"pool_" + strings.ToLower(skyimpl.PoolType_name[int32(v.Type)]) + "_used":      v.Used,
				"pool_" + strings.ToLower(skyimpl.PoolType_name[int32(v.Type)]) + "_committed": v.Committed,
			},
			ts: start,
		})
	}

	return m
}

func extractJVMGCMetrics(service string, start time.Time, gc []*skyimpl.GC) []inputs.Measurement {
	var m []inputs.Measurement
	for _, v := range gc {
		if v == nil {
			continue
		}
		m = append(m, &jvmMeasurement{
			name: jvmMetricName,
			tags: map[string]string{"service": service},
			fields: map[string]interface{}{
				"gc_" + strings.ToLower(skyimpl.GCPhrase_name[int32(v.Phrase)]) + "_count": v.Count,
			},
			ts: start,
		})
	}

	return m
}

func extractJVMThread(service string, start time.Time, thread *skyimpl.Thread) inputs.Measurement {
	return &jvmMeasurement{
		name: jvmMetricName,
		tags: map[string]string{"service": service},
		fields: map[string]interface{}{
			"thread_live_count":               thread.LiveCount,
			"thread_daemon_count":             thread.DaemonCount,
			"thread_peak_count":               thread.PeakCount,
			"thread_runnable_state_count":     thread.RunnableStateThreadCount,
			"thread_blocked_state_count":      thread.BlockedStateThreadCount,
			"thread_waiting_state_count":      thread.WaitingStateThreadCount,
			"thread_time_waiting_state_count": thread.TimedWaitingStateThreadCount,
		},
		ts: start,
	}
}

func extractJVMClass(service string, start time.Time, class *skyimpl.Class) inputs.Measurement {
	return &jvmMeasurement{
		name: jvmMetricName,
		tags: map[string]string{"service": service},
		fields: map[string]interface{}{
			"class_loaded_count":         class.LoadedClassCount,
			"class_total_unloaded_count": class.TotalUnloadedClassCount,
			"class_total_loaded_count":   class.TotalLoadedClassCount,
		},
		ts: start,
	}
}

type JVMMetricReportServerV3 struct {
	skyimpl.UnimplementedJVMMetricReportServiceServer
}

func (*JVMMetricReportServerV3) Collect(ctx context.Context, jvm *skyimpl.JVMMetricCollection) (*skyimpl.Commands, error) {
	log.Debugf("JVMMetricReportService service:%v instance:%v", jvm.Service, jvm.ServiceInstance)

	var (
		m     []inputs.Measurement
		start = time.Now()
	)
	for _, jm := range jvm.Metrics {
		if jm.Cpu != nil {
			m = append(m, extraceJVMCpuMetric(jvm.Service, start, jm.Cpu))
		}
		if len(jm.Memory) != 0 {
			m = append(m, extractJVMMemoryMetrics(jvm.Service, start, jm.Memory)...)
		}
		if len(jm.MemoryPool) != 0 {
			m = append(m, extractJVMMemoryPoolMetrics(jvm.Service, start, jm.MemoryPool)...)
		}
		if len(jm.Gc) != 0 {
			m = append(m, extractJVMGCMetrics(jvm.Service, start, jm.Gc)...)
		}
		if jm.Thread != nil {
			m = append(m, extractJVMThread(jvm.Service, start, jm.Thread))
		}
		if jm.Clazz != nil {
			m = append(m, extractJVMClass(jvm.Service, start, jm.Clazz))
		}
	}

	if len(m) != 0 {
		if err := inputs.FeedMeasurement(jvmMetricName, datakit.Metric, m, &dkio.Option{CollectCost: time.Since(start)}); err != nil {
			dkio.FeedLastError(jvmMetricName, err.Error())
		}
	}

	return &skyimpl.Commands{}, nil
}

type ManagementServerV3 struct {
	skyimpl.UnimplementedManagementServiceServer
}

func (*ManagementServerV3) ReportInstanceProperties(ctx context.Context, mng *skyimpl.InstanceProperties) (*skyimpl.Commands, error) {
	var kvpStr string
	for _, kvp := range mng.Properties {
		kvpStr += fmt.Sprintf("[%v:%v]", kvp.Key, kvp.Value)
	}
	log.Debugf("ReportInstanceProperties service:%v instance:%v properties:%v", mng.Service, mng.ServiceInstance, kvpStr)

	return &skyimpl.Commands{}, nil
}

func (*ManagementServerV3) KeepAlive(ctx context.Context, ping *skyimpl.InstancePingPkg) (*skyimpl.Commands, error) {
	log.Debugf("KeepAlive service:%v instance:%v", ping.Service, ping.ServiceInstance)

	return &skyimpl.Commands{}, nil
}

type DiscoveryServerV3 struct {
	skyimpl.UnimplementedConfigurationDiscoveryServiceServer
}

func (*DiscoveryServerV3) FetchConfigurations(ctx context.Context, cfgReq *skyimpl.ConfigurationSyncRequest) (*skyimpl.Commands, error) {
	log.Debugf("DiscoveryServerV3 service: %s", cfgReq.String())

	return &skyimpl.Commands{}, nil
}

var _ inputs.Measurement = &skywalkingMetricMeasurement{}

type skywalkingMetricMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *skywalkingMetricMeasurement) LineProto() (*dkio.Point, error) {
	return dkio.NewPoint(m.name, m.tags, m.fields, inputs.OptMetric)
}

func (m *skywalkingMetricMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: jvmMetricName,
		Desc: "jvm metrics collected by skywalking language agent.",
		Type: "metric",
		Tags: map[string]interface{}{"service": &inputs.TagInfo{Desc: "service name"}},
		Fields: map[string]interface{}{
			"cpu_usage_percent": &inputs.FieldInfo{
				Type:     inputs.Rate,
				DataType: inputs.Float,
				Unit:     inputs.Percent,
				Desc:     "cpu usage percentile",
			},
			"heap/stack_init": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "heap or stack initialized amount of memory.",
			},
			"heap/stack_max": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "heap or stack max amount of memory.",
			},
			"heap/stack_used": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "heap or stack used amount of memory.",
			},
			"heap/stack_committed": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "heap or stack committed amount of memory.",
			},
			"pool_*_init": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "initialized amount of memory in variety of pool(code_cache_usage,newgen_usage,oldgen_usage,survivor_usage,permgen_usage,metaspace_usage).", // nolint:lll
			},
			"pool_*_max": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "max amount of memory in variety of pool(code_cache_usage,newgen_usage,oldgen_usage,survivor_usage,permgen_usage,metaspace_usage).", // nolint:lll
			},
			"pool_*_used": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "used amount of memory in variety of pool(code_cache_usage,newgen_usage,oldgen_usage,survivor_usage,permgen_usage,metaspace_usage).", // nolint:lll
			},
			"pool_*_committed": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "committed amount of memory in variety of pool(code_cache_usage,newgen_usage,oldgen_usage,survivor_usage,permgen_usage,metaspace_usage).", // nolint:lll
			},
			"gc_phrase_old/new_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "gc old or new count.",
			},
			"thread_live_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "thread live count.",
			},
			"thread_daemon_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "thread daemon count.",
			},
			"thread_peak_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "thread peak count.",
			},
			"thread_runnable_state_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "runnable state thread count.",
			},
			"thread_blocked_state_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "blocked state thread count",
			},
			"thread_waiting_state_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "waiting state thread count.",
			},
			"thread_time_waiting_state_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "time waiting state thread count.",
			},
			"class_loaded_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "loaded class count.",
			},
			"class_total_unloaded_class_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "total unloaded class count.",
			},
			"class_total_loaded_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "total loaded class count.",
			},
		},
	}
}
