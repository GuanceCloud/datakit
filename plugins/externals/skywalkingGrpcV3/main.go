package main

import (
	"flag"
	"fmt"
	"io"
	"net"

	"time"
	"bytes"
	"context"
	"path/filepath"
	"encoding/json"
	"encoding/base64"

	"github.com/influxdata/toml"
	"google.golang.org/grpc"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
	dkio   "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	lang   "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/skywalkingGrpcV3/skywalking/network/language/agent/v3"
	common "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/skywalkingGrpcV3/skywalking/network/common/v3"
	mgment "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/skywalkingGrpcV3/skywalking/network/management/v3"

)

type Skywalking struct {
	GrpcPort int32
	Tags     map[string]string
}

var (
	flagCfgStr    = flag.String("cfg", "", "toml config string")
	flagRPCServer = flag.String("rpc-server", "unix://"+datakit.GRPCDomainSock, "gRPC server")
	flagLog       = flag.String("log", filepath.Join(datakit.InstallDir, "externals", "skywalkingGrpcV3.log"), "log file")
	flagLogLevel  = flag.String("log-level", "info", "log file")

	l            *logger.Logger
	rpcCli       dkio.DataKitClient
	skywalkingV3 Skywalking
)

func main() {
	flag.Parse()

	cfgdata, err := base64.StdEncoding.DecodeString(*flagCfgStr)
	if err != nil {
		panic(err)
	}

	logger.SetGlobalRootLogger(*flagLog, *flagLogLevel, logger.OPT_ENC_CONSOLE|logger.OPT_SHORT_CALLER)
	l = logger.SLogger("skywalkingGrpcV3")
	l.Infof("log level: %s", *flagLogLevel)

	if err := toml.Unmarshal(cfgdata, &skywalkingV3); err != nil {
		l.Errorf("failed to parse toml `%s': %s", *flagCfgStr, err)
		return
	}

	l.Infof("gRPC dial %s...", *flagRPCServer)
	conn, err := grpc.Dial(*flagRPCServer, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second*5))
	if err != nil {
		l.Fatalf("connect RCP failed: %s", err)
	}

	l.Infof("gRPC connect %s ok", *flagRPCServer)
	defer conn.Close()

	rpcCli = dkio.NewDataKitClient(conn)

	skywalkingGrpcServRun(fmt.Sprintf(":%d", skywalkingV3.GrpcPort))
}

type SkywalkingServerV3 struct{
}

func (s *SkywalkingServerV3) Collect(tsc lang.TraceSegmentReportService_CollectServer) error {
	for {
		sgo, err := tsc.Recv()
		if err == io.EOF {
			return tsc.SendAndClose(&common.Commands{})
		}
		if err != nil {
			return err
		}
		err = skywalkGrpcToLineProto(sgo)
		if err != nil {
			return err
		}
	}
}
func skywalkGrpcToLineProto(sg *lang.SegmentObject) error {
	var lines [][]byte
	for _, span := range sg.Spans {
		t := &trace.TraceAdapter{}

		t.Source = "skywalking"

		t.Duration = (span.EndTime - span.StartTime) * 1000
		t.TimestampUs = span.StartTime * 1000
		js, err := json.Marshal(span)
		if err != nil {
			return err
		}
		t.Content = string(js)
		t.Class = "tracing"
		t.ServiceName = sg.Service
		t.OperationName = span.OperationName
		if span.SpanType == lang.SpanType_Entry {
			if len(span.Refs) > 0 {
				t.ParentID = fmt.Sprintf("%s%d", span.Refs[0].ParentTraceSegmentId,
					span.Refs[0].ParentSpanId)
			}
		} else {
			t.ParentID = fmt.Sprintf("%s%d", sg.TraceSegmentId, span.ParentSpanId)
		}

		t.TraceID = sg.TraceId
		t.SpanID = fmt.Sprintf("%s%d", sg.TraceSegmentId, span.SpanId)
		if span.IsError {
			t.IsError = "true"
		}
		if span.SpanType == lang.SpanType_Entry {
			t.SpanType = trace.SPAN_TYPE_ENTRY
		} else if span.SpanType == lang.SpanType_Local {
			t.SpanType = trace.SPAN_TYPE_LOCAL
		} else {
			t.SpanType = trace.SPAN_TYPE_EXIT
		}
		t.EndPoint = span.Peer

		t.Tags = skywalkingV3.Tags
		pt, err := trace.BuildLineProto(t)
		if err != nil {
			l.Error(err)
			continue
		}
		lines = append(lines, pt)
		l.Debug(string(pt))
	}

	if len(lines) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := rpcCli.Send(ctx, &dkio.Request{
		Lines:     bytes.Join(lines, []byte("\n")),
		Precision: "ns",
		Name:      "skywalkingGrpcV3",
		Io:        dkio.IoType_LOGGING,
	})
	if err != nil {
		l.Errorf("feed error: %s", err.Error())
		return err
	}
	l.Debugf("feed %d points, error: `%s'", r.GetPoints(), r.GetErr())
	return nil
}

type SkywalkingManagementServerV3 struct{
}

func (_ *SkywalkingManagementServerV3) ReportInstanceProperties(ctx context.Context, mng *mgment.InstanceProperties) (*common.Commands, error) {
	var kvpStr string
	cmd := &common.Commands{}

	for _, kvp := range mng.Properties {
		kvpStr += fmt.Sprintf("[%v:%v]", kvp.Key, kvp.Value)
	}
	l.Debugf("ReportInstanceProperties service:%v instance:%v properties:%v", mng.Service, mng.ServiceInstance, kvpStr)

	return cmd, nil
}

func (_ *SkywalkingManagementServerV3) KeepAlive(ctx context.Context, ping *mgment.InstancePingPkg) (*common.Commands, error) {
	cmd := &common.Commands{}

	l.Debugf("KeepAlive service:%v instance:%v", ping.Service, ping.ServiceInstance)

	return cmd, nil
}


type SkywalkingJVMMetricReportServerV3 struct {
}

func (_ *SkywalkingJVMMetricReportServerV3) Collect(ctx context.Context, jvm *lang.JVMMetricCollection) (*common.Commands, error) {
	cmd := &common.Commands{}
	//l.Debugf("JVMMetricReportService service:%v instance:%v", jvm.Service, jvm.ServiceInstance)
	return cmd, nil
}

func skywalkingGrpcServRun(addr string) {
	l.Infof("skywalking V3 gRPC starting...")

	rpcListener, err := net.Listen("tcp", addr)
	if err != nil {
		l.Errorf("start skywalking V3 gRPC server %s failed: %v", addr, err)
		return
	}

	l.Infof("start skywalking V3 gRPC server on %s ok", addr)

	rpcServer := grpc.NewServer()
	lang.RegisterTraceSegmentReportServiceServer(rpcServer, &SkywalkingServerV3{})
	lang.RegisterJVMMetricReportServiceServer(rpcServer, &SkywalkingJVMMetricReportServerV3{})
	mgment.RegisterManagementServiceServer(rpcServer, &SkywalkingManagementServerV3{})
	if err := rpcServer.Serve(rpcListener); err != nil {
		l.Error(err)
	}
}