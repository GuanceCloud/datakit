package trace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace/skywalking/v2/common"
	swV2 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace/skywalking/v2/language-agent-v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace/skywalking/v2/register"
)

type SkywalkingServerV2         struct {}
type SkywalkingRegisterServerV2 struct {}
type SkywalkingPingServerV2     struct {}

var regService     sync.Map   //key:serviceName value: serviceID
var regServiceRev  sync.Map   //key:serviceID value: serviceName
var regInstance    sync.Map   //key:uuid value:serviceID
var regEndpoint    sync.Map   //key:endpointName value: endpointID
var regEndpointRev sync.Map   //key:endpointID value: endpointName

var ServiceIdGen  = GenGlobalId(0)
var NetAddrIdGen  = GenGlobalId(10000)
var InstanceIdGen = GenGlobalId(20000)
var EndpointIdGen = GenGlobalId(30000)
//a

func (s *SkywalkingServerV2) Collect(tsc swV2.TraceSegmentReportService_CollectServer) error {
	cmd := new(common.Commands)
	for {
		sgo, err := tsc.Recv()
		if err == io.EOF {
			//return tsc.SendAndClose(&common.Commands{})
			return tsc.SendAndClose(cmd)
		}
		if err != nil {
			return err
		}
		b, _ := json.Marshal(sgo)
		log.Debugf("%#v\n", string(b))

		if err := skywalkGrpcV2ToLineProto(sgo); err != nil {
			log.Error(err)
		}
	}
	return nil
}

func skywalkGrpcV2ToLineProto(sg *swV2.UpstreamSegment) error {
	var service string

	traceId := sg.GlobalTraceIds[0].IdParts[2]
	sgid := sg.Segment.TraceSegmentId.IdParts[2]
	sid := sg.Segment.ServiceId

	sn, ok :=regServiceRev.Load(sid)
	if !ok {
		return fmt.Errorf("Service Id %v not registered", sid)
	}
	switch sn.(type) {
	case string:
		service = sn.(string)
	default:
		return fmt.Errorf("Service Name wrong type")
	}
	for _, span := range sg.Segment.Spans {
		t := TraceAdapter{}

		t.Source = "skywalking"
		t.Duration = (span.EndTime -span.StartTime)*1000
		t.TimestampUs = span.StartTime * 1000
		js ,err := json.Marshal(span)
		if err != nil {
			return err
		}
		t.Content = string(js)
		t.Class         = "tracing"
		t.ServiceName   = service
		t.OperationName = span.OperationName
		if t.OperationName == "" {
			on, ok :=regEndpointRev.Load(span.OperationNameId)
			if !ok {
				return fmt.Errorf("operation name null", sid)
			}
			switch on.(type) {
			case string:
				t.OperationName = on.(string)
			default:
				return fmt.Errorf("operation Name wrong type")
			}
		}

		if len(span.Refs) > 0 {
			t.ParentID = fmt.Sprintf("%v%v", span.Refs[0].ParentTraceSegmentId.IdParts[2], span.Refs[0].ParentSpanId)
		} else if span.ParentSpanId != -1 {
			t.ParentID = fmt.Sprintf("%v%v", sgid, span.ParentSpanId)
		}

		t.TraceID       = fmt.Sprintf("%d", traceId)
		t.SpanID        = fmt.Sprintf("%v%v", sgid, span.SpanId)
		if span.IsError {
			t.IsError   = "true"
		}
		if span.SpanType == common.SpanType_Entry {
			t.SpanType  = SPAN_TYPE_ENTRY
		} else if span.SpanType == common.SpanType_Local {
			t.SpanType  = SPAN_TYPE_LOCAL
		} else {
			t.SpanType  = SPAN_TYPE_EXIT
		}
		t.EndPoint      = span.Peer
		t.Tags = SkywalkingV2Tags
		pt, err := t.MkLineProto()
		if err != nil {
			log.Error(err)
			continue
		}

		if err := dkio.NamedFeed(pt, dkio.Logging, "tracing"); err != nil {
			log.Errorf("io feed err: %s", err)
		}
	}
	return nil
}

func (s *SkywalkingRegisterServerV2) DoServiceRegister(ctx context.Context, r *register.Services) (*register.ServiceRegisterMapping, error) {
	var sid interface{}
	var serviceID int32
	var ok bool
	serMap := register.ServiceRegisterMapping{}
	for _, s := range r.Services{
		service := s.ServiceName
		if sid, ok = regService.Load(service); !ok{
			sid = ServiceIdGen()
			regService.Store(service, sid)
			regServiceRev.Store(sid, service)
		}
		switch sid.(type) {
		case int32:
			serviceID = sid.(int32)
		default:
			log.Errorf("serviceID wrong type")
		}
		log.Infof("DoServiceRegister service: %v serviceID: %d\n", service, serviceID)
		kp := &common.KeyIntValuePair{Key:service, Value:serviceID}
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
		sid  := sin.ServiceId
		if _, ok = regInstance.Load(uuid); !ok{
			serInstanceID = InstanceIdGen()
			regInstance.Store(uuid, sid)
		}
		kp := &common.KeyIntValuePair{Key:uuid, Value:serInstanceID}
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
		sid   := v.ServiceId
		eName := v.EndpointName
		from  := v.From

		if epid, ok = regService.Load(eName); !ok{
			epid = EndpointIdGen()
			regEndpoint.Store(eName, epid)
			regEndpointRev.Store(epid, eName)
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

func (s *SkywalkingRegisterServerV2)DoNetworkAddressRegister(ctx context.Context, r *register.NetAddresses) (*register.NetAddressMapping, error) {
	reg := register.NetAddressMapping{}
	for _, addr := range r.Addresses {
		kp := common.KeyIntValuePair{}
		kp.Key   = addr
		kp.Value = NetAddrIdGen()
		reg.AddressIds = append(reg.AddressIds, &kp)
		log.Infof("DoNetworkAddressRegister addr: %v id: %v", addr, kp.Value)
	}
	return &reg, nil
}
func (s *SkywalkingRegisterServerV2)DoServiceAndNetworkAddressMappingRegister(ctx context.Context, r *register.ServiceAndNetworkAddressMappings) (*common.Commands, error) {
	return new(common.Commands), nil
}

func (s *SkywalkingPingServerV2) DoPing(ctx context.Context, r *register.ServiceInstancePingPkg) (*common.Commands, error) {
	cmds := &common.Commands{}
	if _, ok := regInstance.Load(r.ServiceInstanceUUID); !ok {
		v := r.ServiceInstanceUUID[0:8]+"-"+r.ServiceInstanceUUID[8:12]+"-"+r.ServiceInstanceUUID[12:16]+"-"+r.ServiceInstanceUUID[16:20]+"-"+r.ServiceInstanceUUID[20:32]
		cmd := &common.Command{Command:"ServiceMetadataReset"}
		kv  := &common.KeyStringValuePair{Key:"SerialNumber",
			//Value:r.ServiceInstanceUUID
			Value:v}
		cmd.Args      = append(cmd.Args, kv)
		cmds.Commands = append(cmds.Commands, cmd)
		log.Errorf("Ping %v, %v, %v", r.ServiceInstanceId, v, r.Time)
	} else {
		log.Infof("Ping %v, %v, %v", r.ServiceInstanceId, r.ServiceInstanceUUID, r.Time)
	}
	return  cmds, nil
}

func GenGlobalId(startCnt int32) func () int32 {
	var id int32 = startCnt
	var mutex sync.Mutex

	return func () int32 {
		mutex.Lock()
		id += 1
		mutex.Unlock()
		return id
	}
}