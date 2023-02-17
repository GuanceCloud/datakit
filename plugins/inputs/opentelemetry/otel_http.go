// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"net/http"

	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/opentelemetry/compiled/v1/collector/trace"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func httpStatusRespFunc(resp http.ResponseWriter, req *http.Request, err error) {
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	buf, err := proto.Marshal(&trace.ExportTraceServiceResponse{})
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusInternalServerError)

		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.Write(buf)
}

func handleOTELTraces(resp http.ResponseWriter, req *http.Request) {
	media, _, buf, err := itrace.ParseTracerRequest(req)
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	tsreq := &trace.ExportTraceServiceRequest{}
	switch media {
	case "application/x-protobuf":
		err = proto.Unmarshal(buf, tsreq)
	case "application/json":
		err = protojson.Unmarshal(buf, tsreq)
	default:
		log.Error("unrecognized Content-Type")
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	dktraces := parseResourceSpans(tsreq.ResourceSpans)
	if len(dktraces) != 0 && afterGatherRun != nil {
		afterGatherRun.Run(inputName, dktraces, false)
	}
}
