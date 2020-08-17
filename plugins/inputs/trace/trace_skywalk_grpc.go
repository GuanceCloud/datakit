package trace

import (
	"net"

	"google.golang.org/grpc"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	swV2 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace/skywalking/v2/language-agent-v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace/skywalking/v2/register"
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
