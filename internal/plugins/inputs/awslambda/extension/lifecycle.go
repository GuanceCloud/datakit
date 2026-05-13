// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

// Package lambdaext provides AWS Lambda extension functionality for Datakit.
package lambdaext

import (
	"io"
	"net"
	"net/http"
	"strconv"
)

type LifecycleProcessor interface {
	OnUniversalStart(requestID string, headers map[string]string, payload []byte) (uint64, uint64)
	OnUniversalEnd(requestID string, headers map[string]string, payload []byte) error
}

func StartLifecycleServer(addr string, proc LifecycleProcessor) (*http.Server, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/lambda/start-invocation", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(io.LimitReader(r.Body, 6*1024*1024))
		_ = r.Body.Close()
		headers := flattenHeaders(r.Header)
		requestID := headers["Lambda-Runtime-Aws-Request-Id"]
		if requestID == "" {
			requestID = headers["lambda-runtime-aws-request-id"]
		}
		traceID, parentID := proc.OnUniversalStart(requestID, headers, body)
		w.Header().Set("x-datadog-trace-id", strconv.FormatUint(traceID, 10))
		w.Header().Set("x-datadog-parent-id", strconv.FormatUint(parentID, 10))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})
	mux.HandleFunc("/lambda/end-invocation", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(io.LimitReader(r.Body, 6*1024*1024))
		_ = r.Body.Close()
		headers := flattenHeaders(r.Header)
		requestID := headers["Lambda-Runtime-Aws-Request-Id"]
		if requestID == "" {
			requestID = headers["lambda-runtime-aws-request-id"]
		}
		_ = proc.OnUniversalEnd(requestID, headers, body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})
	mux.HandleFunc("/lambda/hello", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{}`)
	})

	server := &http.Server{Addr: addr, Handler: mux}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	server.Addr = ln.Addr().String()
	go func() {
		_ = server.Serve(ln)
	}()
	return server, nil
}

func flattenHeaders(headers http.Header) map[string]string {
	out := make(map[string]string, len(headers))
	for k, values := range headers {
		if len(values) > 0 {
			out[k] = values[0]
		}
	}
	return out
}
