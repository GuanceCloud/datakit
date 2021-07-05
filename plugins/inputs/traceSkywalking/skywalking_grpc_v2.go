package traceSkywalking

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"

	"google.golang.org/grpc"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/traceSkywalking/v2/common"
	swV2 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/traceSkywalking/v2/language-agent-v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/traceSkywalking/v2/register"
)

type SkywalkingServerV2 struct{}
type SkywalkingRegisterServerV2 struct{}
type SkywalkingPingServerV2 struct{}
type SkywalkingJVMMetricServerV2 struct{}

var (
	NetAddrIdGen   = GenGlobalId(10000)
	ServiceIdGen   = GenGlobalId(20000)
	InstanceIdGen  = GenGlobalId(30000)
	EndpointIdGen  = GenGlobalId(40000)
	SerialNumIdGen = GenGlobalId(50000)

	RegService     = &sync.Map{} //key: id,           value: serviceName
	RegServiceRev  = &sync.Map{} //key: serviceName,  value: id
	RegInstance    = &sync.Map{} //key: id,           value: instanceUUID
	RegInstanceRev = &sync.Map{} //key: instanceUUID, value: id
	RegEndpoint    = &sync.Map{} //key: id,           value: endpointName
	RegEndpointRev = &sync.Map{} //key: endpointName, value: id
)

func SkyWalkingServerRunV2(addr string) {
	log.Infof("skywalking V2 gRPC starting...")

	rpcListener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorf("start skywalking V2 gRPC server %s failed: %v", addr, err)
		return
	}

	log.Infof("start skywalking V2 gRPC server on %s ok", addr)

	rpcServer := grpc.NewServer()
	swV2.RegisterTraceSegmentReportServiceServer(rpcServer, &SkywalkingServerV2{})
	swV2.RegisterJVMMetricReportServiceServer(rpcServer, &SkywalkingJVMMetricServerV2{})
	register.RegisterRegisterServer(rpcServer, &SkywalkingRegisterServerV2{})
	register.RegisterServiceInstancePingServer(rpcServer, &SkywalkingPingServerV2{})

	go func() {
		if err := rpcServer.Serve(rpcListener); err != nil {
			log.Error(err)
		}

		log.Info("skywalking V2 gRPC server exit")
	}()

	<-datakit.Exit.Wait()
	log.Info("skywalking V2 gRPC server stopping...")
	rpcServer.Stop()
	return
}

func (s *SkywalkingServerV2) Collect(tsc swV2.TraceSegmentReportService_CollectServer) error {
	cmd := new(common.Commands)
	for {
		sgo, err := tsc.Recv()
		if err == io.EOF {
			return tsc.SendAndClose(cmd)
		}

		if err != nil {
			return err
		}

		b, _ := json.Marshal(sgo)
		log.Debugf("%#v\n", string(b))

		if err := skywalkGrpcV2ToLineProto(sgo); err != nil {
			dkio.FeedLastError(inputName, err.Error())
			log.Error(err)
		}
	}
}

func skywalkGrpcV2ToLineProto(sg *swV2.UpstreamSegment) error {
	var service string

	traceId := sg.GlobalTraceIds[0].IdParts[2]
	sgid := sg.Segment.TraceSegmentId.IdParts[2]
	sid := sg.Segment.ServiceId

	sn, ok := RegService.Load(sid)
	if !ok {
		return fmt.Errorf("Service Id %v not registered", sid)
	}
	switch sn.(type) {
	case string:
		service = sn.(string)
	default:
		return fmt.Errorf("Service Name wrong type")
	}

	adapterGroup := []*trace.TraceAdapter{}
	for _, span := range sg.Segment.Spans {
		tAdapter := &trace.TraceAdapter{}

		tAdapter.Source = "skywalking"
		tAdapter.Duration = (span.EndTime - span.StartTime) * 1000000
		tAdapter.Start = span.StartTime * 1000000
		js, err := json.Marshal(span)
		if err != nil {
			return err
		}
		tAdapter.Content = string(js)
		tAdapter.ServiceName = service
		tAdapter.OperationName = span.OperationName
		if tAdapter.OperationName == "" {
			on, ok := RegEndpoint.Load(span.OperationNameId)
			if !ok {
				return fmt.Errorf("operation name %s null", tAdapter.OperationName)
			}
			switch on.(type) {
			case string:
				tAdapter.OperationName = on.(string)
			default:
				return fmt.Errorf("operation Name wrong type")
			}
		}

		if len(span.Refs) > 0 {
			tAdapter.ParentID = fmt.Sprintf("%v%v", span.Refs[0].ParentTraceSegmentId.IdParts[2], span.Refs[0].ParentSpanId)
		} else if span.ParentSpanId != -1 {
			tAdapter.ParentID = fmt.Sprintf("%v%v", sgid, span.ParentSpanId)
		}

		tAdapter.TraceID = fmt.Sprintf("%d", traceId)
		tAdapter.SpanID = fmt.Sprintf("%v%v", sgid, span.SpanId)
		tAdapter.Status = trace.STATUS_OK
		if span.IsError {
			tAdapter.Status = trace.STATUS_ERR
		}
		if span.SpanType == common.SpanType_Entry {
			tAdapter.SpanType = trace.SPAN_TYPE_ENTRY
		} else if span.SpanType == common.SpanType_Local {
			tAdapter.SpanType = trace.SPAN_TYPE_LOCAL
		} else {
			tAdapter.SpanType = trace.SPAN_TYPE_EXIT
		}
		tAdapter.EndPoint = span.Peer
		tAdapter.Tags = SkywalkingTagsV2

		// run tracing sample function
		if conf := trace.TraceSampleMatcher(sampleConfs, tAdapter.Tags); conf != nil {
			if trcid, err := strconv.ParseUint(tAdapter.TraceID, 10, 64); err == nil {
				if !trace.IgnoreErrSampleMW(tAdapter.Status, trace.IgnoreTagsSampleMW(tAdapter.Tags, conf.IgnoreTagsList, trace.DefSampleFunc))(trcid, conf.Rate, conf.Scope) {
					continue
				}
			} else {
				log.Errorf("Parse uint64 trace id failed when doing tracing sample")
			}
		}

		adapterGroup = append(adapterGroup, tAdapter)
	}
	trace.MkLineProto(adapterGroup, inputName)

	return nil
}

func (s *SkywalkingRegisterServerV2) DoServiceRegister(ctx context.Context, r *register.Services) (*register.ServiceRegisterMapping, error) {
	var sid interface{}
	var serviceID int32
	var ok bool
	serMap := register.ServiceRegisterMapping{}
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

func (s *SkywalkingRegisterServerV2) DoServiceInstanceRegister(ctx context.Context, r *register.ServiceInstances) (*register.ServiceInstanceRegisterMapping, error) {
	var ok bool
	var serInstanceID int32
	regMap := register.ServiceInstanceRegisterMapping{}

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

func (s *SkywalkingRegisterServerV2) DoEndpointRegister(ctx context.Context, r *register.Endpoints) (*register.EndpointMapping, error) {
	var epid interface{}
	var ok bool
	var endpointID int32

	reg := register.EndpointMapping{}
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
		log.Infof("DoEndpointRegister serviceID=%v endpontName=%v from=%v endpointID=%v\n", r.ServiceId,
			r.EndpointName, r.From, r.EndpointId)
		reg.Elements = append(reg.Elements, &r)
	}

	return &reg, nil
}

func (s *SkywalkingRegisterServerV2) DoNetworkAddressRegister(ctx context.Context, r *register.NetAddresses) (*register.NetAddressMapping, error) {
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

func (s *SkywalkingRegisterServerV2) DoServiceAndNetworkAddressMappingRegister(ctx context.Context, r *register.ServiceAndNetworkAddressMappings) (*common.Commands, error) {
	return new(common.Commands), nil
}

func (s *SkywalkingPingServerV2) DoPing(ctx context.Context, r *register.ServiceInstancePingPkg) (*common.Commands, error) {
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

func (s *SkywalkingJVMMetricServerV2) Collect(ctx context.Context, jvm *swV2.JVMMetricCollection) (*common.Commands, error) {
	cmds := &common.Commands{}
	log.Debugf("JVMMetricReportService %v", jvm.ServiceInstanceId)
	return cmds, nil
}

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
