// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package awslambda

import (
	"net"
	"net/http"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	lambdaextsrv "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/extension"
	lambdatrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/trace"
)

func NewTracingTestInput(feeder dkio.Feeder, managed bool) *Input {
	ipt := &Input{
		EnableMetricCollection: true,
		EnableLogCollection:    true,
		feeder:                 feeder,
		tags:                   map[string]string{},
		g:                      goroutine.G("awslambda-test"),
		runtimeDoneChan:        make(chan string, 32),
	}
	if managed {
		ipt.tags[LambdaInitializationType] = "lambda-managed-instances"
	}
	ipt.traceProcessor = lambdatrace.NewProcessor(lambdatrace.NewPointSink(inputName, ipt.feeder, ipt.tags), managed)
	lambdatrace.SetActiveProcessor(ipt.traceProcessor)
	return ipt
}

func StartLifecycleServerForTest(ipt *Input) (*http.Server, error) {
	return lambdaextsrv.StartLifecycleServer("127.0.0.1:0", ipt.traceProcessor)
}

func (ipt *Input) TraceProcessorForTest() *lambdatrace.Processor {
	return ipt.traceProcessor
}

func CanBindForTest(addr string) bool {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}
