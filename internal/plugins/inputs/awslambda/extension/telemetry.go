// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package lambdaext

import (
	"net"
	"net/http"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/telemetry"
	lambdatrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/trace"
)

func StartTelemetryServer(addr string, proc *lambdatrace.Processor) (*http.Server, error) {
	listener := telemetry.NewTelemetryListener()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_ = listener.HandlerTelemetry(w, r)
	})
	server := &http.Server{Addr: addr, Handler: mux}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	go func() {
		_ = server.Serve(ln)
	}()
	go func() {
		for events := range listener.GetPullChan() {
			for _, event := range events {
				_ = proc.OnTelemetryEvent(event)
			}
		}
	}()
	return server, nil
}
