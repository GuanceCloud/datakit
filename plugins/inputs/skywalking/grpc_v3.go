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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	configv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/agent/configuration/v3"
	commonv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/common/v3"
	eventv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/event/v3"
	agentv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/language/agent/v3"
	profilev3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/language/profile/v3"
	mgmtv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/management/v3"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func registerServerV3(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorf("### skywalking grpc server v3 listening on %s failed: %v", addr, err)

		return
	}
	log.Debugf("### skywalking grpc v3 listening on: %s", addr)

	skySvr = grpc.NewServer()
	agentv3.RegisterTraceSegmentReportServiceServer(skySvr, &TraceReportServerV3{})
	eventv3.RegisterEventServiceServer(skySvr, &EventServerV3{})
	agentv3.RegisterJVMMetricReportServiceServer(skySvr, &JVMMetricReportServerV3{})
	// profilev3.RegisterProfileTaskServer(skySvr, &ProfileTaskServerV3{})
	mgmtv3.RegisterManagementServiceServer(skySvr, &ManagementServerV3{})
	configv3.RegisterConfigurationDiscoveryServiceServer(skySvr, &DiscoveryServerV3{})

	if err = skySvr.Serve(listener); err != nil {
		log.Error(err.Error())
	}
	log.Debug("### skywalking v3 exits")
}

type TraceReportServerV3 struct {
	agentv3.UnimplementedTraceSegmentReportServiceServer
}

func (*TraceReportServerV3) Collect(tsc agentv3.TraceSegmentReportService_CollectServer) error {
	for {
		segobj, err := tsc.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return tsc.SendAndClose(&commonv3.Commands{})
			}
			log.Error(err.Error())

			return err
		}
		log.Debug("### v3 segment received")

		if storage == nil {
			parseSegmentObject(segobj)
		} else {
			buf, err := proto.Marshal(segobj)
			if err != nil {
				log.Error(err.Error())

				return err
			}

			param := &itrace.TraceParameters{Meta: &itrace.TraceMeta{Buf: buf}}
			if err = storage.Send(param); err != nil {
				log.Error(err.Error())

				return err
			}
		}
	}
}

func (*TraceReportServerV3) CollectInSync(ctx context.Context, seg *agentv3.SegmentCollection) (*commonv3.Commands, error) {
	log.Debugf("### reveived collect sync: %s", seg.String())

	return &commonv3.Commands{}, nil
}

func getTagValue(tags []*commonv3.KeyStringValuePair, key string) (value string, ok bool) {
	for i := range tags {
		if key == tags[i].Key {
			if len(tags[i].Value) == 0 {
				return "", false
			} else {
				return tags[i].Value, true
			}
		}
	}

	return "", false
}

func mapToSpanSourceType(layer agentv3.SpanLayer) string {
	switch layer {
	case agentv3.SpanLayer_Database:
		return itrace.SPAN_SOURCE_DB
	case agentv3.SpanLayer_Cache:
		return itrace.SPAN_SOURCE_CACHE
	case agentv3.SpanLayer_RPCFramework:
		return itrace.SPAN_SOURCE_FRAMEWORK
	case agentv3.SpanLayer_Http:
		return itrace.SPAN_SOURCE_WEB
	case agentv3.SpanLayer_MQ:
		return itrace.SPAN_SOURCE_MSGQUE
	case agentv3.SpanLayer_FAAS:
		return itrace.SPAN_SOURCE_APP
	case agentv3.SpanLayer_Unknown:
		return itrace.SPAN_SOURCE_CUSTOMER
	default:
		return itrace.SPAN_SOURCE_CUSTOMER
	}
}

func parseSegmentObjectWrapper(param *itrace.TraceParameters) error {
	if param == nil || param.Meta == nil || len(param.Meta.Buf) == 0 {
		return errors.New("invalid parameters")
	}

	segobj := &agentv3.SegmentObject{}
	if err := proto.Unmarshal(param.Meta.Buf, segobj); err != nil {
		return err
	}

	parseSegmentObject(segobj)

	return nil
}

func parseSegmentObject(segment *agentv3.SegmentObject) {
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
				if span.Refs[0].RefType == agentv3.RefType_CrossProcess && strings.Contains(span.Refs[0].ParentService, "_rum_") {
					dktrace = append(dktrace, &itrace.DatakitSpan{
						TraceID:    segment.TraceId,
						ParentID:   "0",
						SpanID:     dkspan.ParentID,
						Service:    span.Refs[0].ParentService,
						Resource:   span.Refs[0].ParentEndpoint,
						Operation:  span.Refs[0].ParentEndpoint,
						Source:     inputName,
						SpanType:   itrace.SPAN_TYPE_ENTRY,
						SourceType: itrace.SPAN_SOURCE_WEB,
						EndPoint:   span.Refs[0].NetworkAddressUsedAtPeer,
						Start:      dkspan.Start - int64(time.Millisecond),
						Duration:   int64(time.Millisecond),
						Status:     itrace.STATUS_OK,
					})
				}
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
		case agentv3.SpanType_Entry:
			dkspan.SpanType = itrace.SPAN_TYPE_ENTRY
		case agentv3.SpanType_Local:
			dkspan.SpanType = itrace.SPAN_TYPE_LOCAL
		case agentv3.SpanType_Exit:
			dkspan.SpanType = itrace.SPAN_TYPE_EXIT
		default:
			dkspan.SpanType = itrace.SPAN_TYPE_ENTRY
		}

		for i := range plugins {
			if value, ok := getTagValue(span.Tags, plugins[i]); ok {
				dkspan.Service = value
				dkspan.SpanType = itrace.SPAN_TYPE_ENTRY
				dkspan.SourceType = mapToSpanSourceType(span.SpanLayer)
				switch span.SpanLayer { // nolint: exhaustive
				case agentv3.SpanLayer_Database, agentv3.SpanLayer_Cache:
					if res, ok := getTagValue(span.Tags, "db.statement"); ok {
						dkspan.Resource = res
					}
				case agentv3.SpanLayer_MQ:
				case agentv3.SpanLayer_Http:
				case agentv3.SpanLayer_RPCFramework:
				case agentv3.SpanLayer_FAAS:
				case agentv3.SpanLayer_Unknown:
				}
			}
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

	if len(dktrace) != 0 && afterGatherRun != nil {
		afterGatherRun.Run(inputName, itrace.DatakitTraces{dktrace}, false)
	}
}

type EventServerV3 struct {
	eventv3.UnimplementedEventServiceServer
}

func (*EventServerV3) Collect(esrv eventv3.EventService_CollectServer) error {
	for {
		event, err := esrv.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return esrv.SendAndClose(&commonv3.Commands{})
			}
			log.Debug(err.Error())

			return err
		}

		log.Debugf("### reveived service event: %s", event.String())
	}
}

type jvmMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *jvmMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOpt())
}

func (*jvmMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{}
}

func extractJVMCpuMetric(service string, start time.Time, cpu *commonv3.CPU) inputs.Measurement {
	return &jvmMeasurement{
		name:   jvmMetricName,
		tags:   map[string]string{"service": service},
		fields: map[string]interface{}{"cpu_usage_percent": cpu.UsagePercent},
		ts:     start,
	}
}

func extractJVMMemoryMetrics(service string, start time.Time, memory []*agentv3.Memory) []inputs.Measurement {
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

func extractJVMMemoryPoolMetrics(service string, start time.Time, memoryPool []*agentv3.MemoryPool) []inputs.Measurement {
	var m []inputs.Measurement
	for _, v := range memoryPool {
		if v == nil {
			continue
		}
		m = append(m, &jvmMeasurement{
			name: jvmMetricName,
			tags: map[string]string{"service": service},
			fields: map[string]interface{}{
				"pool_" + strings.ToLower(agentv3.PoolType_name[int32(v.Type)]) + "_init":      v.Init,
				"pool_" + strings.ToLower(agentv3.PoolType_name[int32(v.Type)]) + "_max":       v.Max,
				"pool_" + strings.ToLower(agentv3.PoolType_name[int32(v.Type)]) + "_used":      v.Used,
				"pool_" + strings.ToLower(agentv3.PoolType_name[int32(v.Type)]) + "_committed": v.Committed,
			},
			ts: start,
		})
	}

	return m
}

func extractJVMGCMetrics(service string, start time.Time, gc []*agentv3.GC) []inputs.Measurement {
	var m []inputs.Measurement
	for _, v := range gc {
		if v == nil {
			continue
		}
		m = append(m, &jvmMeasurement{
			name: jvmMetricName,
			tags: map[string]string{"service": service},
			fields: map[string]interface{}{
				"gc_" + strings.ToLower(agentv3.GCPhase_name[int32(v.Phase)]) + "_count": v.Count,
			},
			ts: start,
		})
	}

	return m
}

func extractJVMThread(service string, start time.Time, thread *agentv3.Thread) inputs.Measurement {
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

func extractJVMClass(service string, start time.Time, class *agentv3.Class) inputs.Measurement {
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
	agentv3.UnimplementedJVMMetricReportServiceServer
}

func (*JVMMetricReportServerV3) Collect(ctx context.Context, jvm *agentv3.JVMMetricCollection) (*commonv3.Commands, error) {
	log.Debugf("### JVMMetricReportService service:%v instance:%v", jvm.Service, jvm.ServiceInstance)

	var (
		m     []inputs.Measurement
		start = time.Now()
	)
	for _, jm := range jvm.Metrics {
		if jm.Cpu != nil {
			m = append(m, extractJVMCpuMetric(jvm.Service, start, jm.Cpu))
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

	return &commonv3.Commands{}, nil
}

type ProfileTaskServerV3 struct {
	profilev3.UnimplementedProfileTaskServer
}

func (*ProfileTaskServerV3) GetProfileTaskCommands(ctx context.Context, task *profilev3.ProfileTaskCommandQuery) (*commonv3.Commands, error) {
	return nil, nil
}

func (*ProfileTaskServerV3) CollectSnapshot(psrv profilev3.ProfileTask_CollectSnapshotServer) error {
	return nil
}

func (*ProfileTaskServerV3) ReportTaskFinish(ctx context.Context, reporter *profilev3.ProfileTaskFinishReport) (*commonv3.Commands, error) {
	return nil, nil
}

type ManagementServerV3 struct {
	mgmtv3.UnimplementedManagementServiceServer
}

func (*ManagementServerV3) ReportInstanceProperties(ctx context.Context, mgmt *mgmtv3.InstanceProperties) (*commonv3.Commands, error) {
	var kvpStr string
	for _, kvp := range mgmt.Properties {
		kvpStr += fmt.Sprintf("[%v:%v]", kvp.Key, kvp.Value)
	}
	log.Debugf("### ReportInstanceProperties service:%v instance:%v properties:%v", mgmt.Service, mgmt.ServiceInstance, kvpStr)

	return &commonv3.Commands{}, nil
}

func (*ManagementServerV3) KeepAlive(ctx context.Context, ping *mgmtv3.InstancePingPkg) (*commonv3.Commands, error) {
	log.Debugf("### KeepAlive service:%v instance:%v", ping.Service, ping.ServiceInstance)

	return &commonv3.Commands{}, nil
}

type DiscoveryServerV3 struct {
	configv3.UnimplementedConfigurationDiscoveryServiceServer
}

func (*DiscoveryServerV3) FetchConfigurations(ctx context.Context, cfgReq *configv3.ConfigurationSyncRequest) (*commonv3.Commands, error) {
	log.Debugf("### DiscoveryServerV3 service: %s", cfgReq.String())

	return &commonv3.Commands{}, nil
}

var _ inputs.Measurement = &skywalkingMetricMeasurement{}

type skywalkingMetricMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *skywalkingMetricMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOpt())
}

func (*skywalkingMetricMeasurement) Info() *inputs.MeasurementInfo {
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
