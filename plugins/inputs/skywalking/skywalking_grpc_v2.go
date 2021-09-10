package skywalking

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"google.golang.org/grpc"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/v2/common"
	swV2 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/v2/language-agent-v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/v2/register"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

var (
	NetAddrIdGen   = GenGlobalId(10000)
	ServiceIdGen   = GenGlobalId(20000)
	InstanceIdGen  = GenGlobalId(30000)
	EndpointIdGen  = GenGlobalId(40000)
	SerialNumIdGen = GenGlobalId(50000)
)

var (
	RegService     = &sync.Map{} //key: id,           value: serviceName
	RegServiceRev  = &sync.Map{} //key: serviceName,  value: id
	RegInstance    = &sync.Map{} //key: id,           value: instanceUUID
	RegInstanceRev = &sync.Map{} //key: instanceUUID, value: id
	RegEndpoint    = &sync.Map{} //key: id,           value: endpointName
	RegEndpointRev = &sync.Map{} //key: endpointName, value: id
)

func GenGlobalId(startCnt int32) func() int32 {
	var id int32 = startCnt
	var mutex sync.Mutex

	return func() int32 {
		var rtnId int32
		mutex.Lock()
		id += 1
		rtnId = id
		mutex.Unlock()
		return rtnId
	}
}

func skyWalkingV2ServerRun(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorf("start skywalking V2 gRPC server %s failed: %v", addr, err)

		return
	}
	log.Infof("skywalking v2 listening on: %s", addr)

	srv := grpc.NewServer()
	swV2.RegisterTraceSegmentReportServiceServer(srv, &SkyWalkingServerV2{})
	swV2.RegisterJVMMetricReportServiceServer(srv, &SkywalkingJVMMetricServerV2{})
	register.RegisterRegisterServer(srv, &SkyWalkingRegisterServerV2{})
	register.RegisterServiceInstancePingServer(srv, &SkywalkingPingServerV2{})
	if err = srv.Serve(listener); err != nil {
		log.Error(err)
	}
	log.Info("skywalking v2 exits")
}

type SkyWalkingServerV2 struct {
	swV2.UnimplementedTraceSegmentReportServiceServer
}

func (*SkyWalkingServerV2) Collect(tsc swV2.TraceSegmentReportService_CollectServer) (err error) {
	defer func() {
		if err != nil {
			log.Error(err)
		}
	}()

	cmd := &common.Commands{}
	for {
		segment, err := tsc.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return tsc.SendAndClose(cmd)
			}

			return err
		}

		group, err := upstmSegmentToAdapters(segment)
		if err != nil {
			return err
		}

		if len(group) != 0 {
			trace.MkLineProto(group, inputName)
		} else {
			log.Debug("empty v2 segment")
		}
	}
}

func upstmSegmentToAdapters(segment *swV2.UpstreamSegment, filters ...upstmSegmentFilter) ([]*trace.TraceAdapter, error) {
	// run all filters
	for _, filter := range filters {
		if filter(segment) == nil {
			return nil, nil
		}
	}

	var (
		traceId = segment.GlobalTraceIds[0].IdParts[2]
		sgid    = segment.Segment.TraceSegmentId.IdParts[2]
		sid     = segment.Segment.ServiceId
	)

	sn, ok := RegService.Load(sid)
	if !ok {
		return nil, fmt.Errorf("Service Id %v not registered", sid)
	}

	var service string
	switch sn.(type) {
	case string:
		service = sn.(string)
	default:
		return nil, fmt.Errorf("Service Name wrong type")
	}

	var group []*trace.TraceAdapter
	for _, span := range segment.Segment.Spans {
		adapter := &trace.TraceAdapter{Source: inputName}
		adapter.Duration = (span.EndTime - span.StartTime) * 1000000
		adapter.Start = span.StartTime * 1000000
		js, err := json.Marshal(span)
		if err != nil {
			return nil, err
		}
		adapter.Content = string(js)
		adapter.ServiceName = service
		adapter.OperationName = span.OperationName
		if adapter.OperationName == "" {
			on, ok := RegEndpoint.Load(span.OperationNameId)
			if !ok {
				return nil, fmt.Errorf("operation name %s null", adapter.OperationName)
			}
			switch on.(type) {
			case string:
				adapter.OperationName = on.(string)
			default:
				return nil, fmt.Errorf("operation Name wrong type")
			}
		}

		if len(span.Refs) > 0 {
			adapter.ParentID = fmt.Sprintf("%v%v", span.Refs[0].ParentTraceSegmentId.IdParts[2], span.Refs[0].ParentSpanId)
		} else if span.ParentSpanId != -1 {
			adapter.ParentID = fmt.Sprintf("%v%v", sgid, span.ParentSpanId)
		}

		adapter.TraceID = fmt.Sprintf("%d", traceId)
		adapter.SpanID = fmt.Sprintf("%v%v", sgid, span.SpanId)
		adapter.Status = trace.STATUS_OK
		if span.IsError {
			adapter.Status = trace.STATUS_ERR
		}
		if span.SpanType == common.SpanType_Entry {
			adapter.SpanType = trace.SPAN_TYPE_ENTRY
		} else if span.SpanType == common.SpanType_Local {
			adapter.SpanType = trace.SPAN_TYPE_LOCAL
		} else {
			adapter.SpanType = trace.SPAN_TYPE_EXIT
		}
		adapter.EndPoint = span.Peer
		adapter.Tags = skywalkingV2Tags

		group = append(group, adapter)
	}

	return group, nil
}

type SkyWalkingRegisterServerV2 struct {
	register.UnimplementedRegisterServer
}

func (*SkyWalkingRegisterServerV2) DoServiceRegister(ctx context.Context, r *register.Services) (*register.ServiceRegisterMapping, error) {
	var (
		sid       interface{}
		serviceID int32
		ok        bool
		serMap    = register.ServiceRegisterMapping{}
	)
	for _, s := range r.Services {
		service := s.ServiceName
		if sid, ok = RegServiceRev.Load(service); !ok {
			sid = ServiceIdGen()
			RegService.Store(sid, service)
			RegServiceRev.Store(service, sid)
		}
		switch sid.(type) {
		case int32:
			serviceID = sid.(int32)
		default:
			log.Errorf("serviceID wrong type")
		}
		log.Infof("DoServiceRegister service: %v serviceID: %d\n", service, serviceID)
		kp := &common.KeyIntValuePair{Key: service, Value: serviceID}
		serMap.Services = append(serMap.Services, kp)
	}

	return &serMap, nil
}

func (*SkyWalkingRegisterServerV2) DoServiceInstanceRegister(ctx context.Context, r *register.ServiceInstances) (*register.ServiceInstanceRegisterMapping, error) {
	var (
		ok            bool
		serInstanceID int32
		regMap        = register.ServiceInstanceRegisterMapping{}
	)
	for _, sin := range r.Instances {
		uuid := sin.InstanceUUID
		sid := sin.ServiceId
		if _, ok = RegInstanceRev.Load(uuid); !ok {
			serInstanceID = InstanceIdGen()
			RegInstance.Store(serInstanceID, uuid)
			RegInstanceRev.Store(uuid, serInstanceID)
		}
		kp := &common.KeyIntValuePair{Key: uuid, Value: serInstanceID}
		regMap.ServiceInstances = append(regMap.ServiceInstances, kp)
		log.Infof("DoServiceInstanceRegister serviceID: %v uuid: %v instanceID: %v\n", sid, uuid, serInstanceID)
	}

	return &regMap, nil
}

func (*SkyWalkingRegisterServerV2) DoEndpointRegister(ctx context.Context, r *register.Endpoints) (*register.EndpointMapping, error) {
	var (
		epid       interface{}
		ok         bool
		endpointID int32
		reg        = register.EndpointMapping{}
	)
	for _, v := range r.Endpoints {
		r := register.EndpointMappingElement{}
		sid := v.ServiceId
		eName := v.EndpointName
		from := v.From

		if epid, ok = RegEndpointRev.Load(eName); !ok {
			epid = EndpointIdGen()
			RegEndpoint.Store(epid, eName)
			RegEndpointRev.Store(eName, epid)
		}
		switch epid.(type) {
		case int32:
			endpointID = epid.(int32)
		default:
			log.Errorf("endpointId wrong type")
		}

		r.EndpointName = eName
		r.ServiceId = sid
		r.From = from
		r.EndpointId = endpointID
		log.Infof("DoEndpointRegister serviceID=%v endpontName=%v from=%v endpointID=%v\n", r.ServiceId, r.EndpointName, r.From, r.EndpointId)
		reg.Elements = append(reg.Elements, &r)
	}

	return &reg, nil
}

func (*SkyWalkingRegisterServerV2) DoNetworkAddressRegister(ctx context.Context, r *register.NetAddresses) (*register.NetAddressMapping, error) {
	reg := register.NetAddressMapping{}
	for _, addr := range r.Addresses {
		kp := common.KeyIntValuePair{}
		kp.Key = addr
		kp.Value = NetAddrIdGen()
		reg.AddressIds = append(reg.AddressIds, &kp)
		log.Infof("DoNetworkAddressRegister addr: %v id: %v", addr, kp.Value)
	}

	return &reg, nil
}

func (*SkyWalkingRegisterServerV2) DoServiceAndNetworkAddressMappingRegister(ctx context.Context, r *register.ServiceAndNetworkAddressMappings) (*common.Commands, error) {
	return new(common.Commands), nil
}

type SkywalkingJVMMetricServerV2 struct {
	swV2.UnimplementedJVMMetricReportServiceServer
}

func (*SkywalkingJVMMetricServerV2) Collect(ctx context.Context, jvm *swV2.JVMMetricCollection) (*common.Commands, error) {
	cmds := &common.Commands{}
	log.Debugf("JVMMetricReportService %v", jvm.ServiceInstanceId)

	return cmds, nil
}

type SkywalkingPingServerV2 struct {
	register.UnimplementedServiceInstancePingServer
}

func (*SkywalkingPingServerV2) DoPing(ctx context.Context, r *register.ServiceInstancePingPkg) (*common.Commands, error) {
	cmds := &common.Commands{}
	if _, ok := RegInstanceRev.Load(r.ServiceInstanceUUID); !ok {
		cmd := &common.Command{Command: "ServiceMetadataReset"}
		kv := &common.KeyStringValuePair{
			Key:   "SerialNumber",
			Value: fmt.Sprintf("%v", SerialNumIdGen()),
		}
		cmd.Args = append(cmd.Args, kv)
		cmds.Commands = append(cmds.Commands, cmd)
		log.Errorf("Ping %v, %v, %v", r.ServiceInstanceId, r.ServiceInstanceUUID, r.Time)
	} else {
		log.Debugf("Ping %v, %v, %v", r.ServiceInstanceId, r.ServiceInstanceUUID, r.Time)
	}

	return cmds, nil
}
