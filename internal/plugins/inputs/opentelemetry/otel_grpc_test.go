// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"context"
	"sync"
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/metrics"
	tv1 "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/collector/trace/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

func Test_grpcServer(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		reg := prometheus.NewRegistry()
		reg.MustRegister(itrace.Metrics()...)

		ipt := defaultInput()
		ipt.setup()

		ipt.GRPCConfig = &gRPC{
			Address: "0.0.0.0:0",
		}

		var wg sync.WaitGroup

		wg.Add(1)

		// setup gRPC server
		go func() {
			defer wg.Done()
			ipt.GRPCConfig.runGRPCV1(ipt)
		}()

		time.Sleep(time.Second) // wait gRPC server ok.

		// client
		conn, err := grpc.Dial(
			ipt.GRPCConfig.trueAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
			grpc.WithTimeout(5*time.Second),
		)
		assert.NoError(t, err)
		defer conn.Close()

		client := tv1.NewTraceServiceClient(conn)

		// setup test data
		req := createTestTraceData(100)

		// send
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.Export(ctx, req)
		assert.NoError(t, err)

		t.Logf("resp: %v", resp)

		ipt.GRPCConfig.stop() // stop server
		wg.Wait()

		mfs, err := reg.Gather()
		assert.NoError(t, err)

		t.Logf("metrics:\n%s", metrics.MetricFamily2Text(mfs))
	})
}
