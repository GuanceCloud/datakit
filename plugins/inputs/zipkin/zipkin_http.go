package zipkin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	zpkmodel "github.com/openzipkin/zipkin-go/model"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/zipkin/corev1"
)

func handleZipkinTraceV1(resp http.ResponseWriter, req *http.Request) {
	contentType, body, err := itrace.ParseTracingRequest(req)
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	var dktrace itrace.DatakitTrace
	switch contentType {
	case "application/x-thrift":
		var zspans []*corev1.Span
		if zspans, err = unmarshalZipkinThriftV1(body); err == nil {
			dktrace = thriftV1SpansToDkTrace(zspans)
		}
	case "application/json":
		var zspans []*ZipkinSpanV1
		if err = json.NewDecoder(req.Body).Decode(&zspans); err == nil {
			dktrace = jsonV1SpansToDkTrace(zspans)
		}
	default:
		err = fmt.Errorf("zipkin V1 unsupported Content-Type: %s", contentType)
	}

	if err != nil {
		log.Errorf("convert zipkin trace [%s] to datakit trace failed: %s", contentType, err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	if len(dktrace) == 0 {
		log.Warn("empty datakit trace")
	} else {
		afterGatherRun.Run(inputName, dktrace, false)
	}

	resp.WriteHeader(http.StatusOK)
}

func handleZipkinTraceV2(resp http.ResponseWriter, req *http.Request) {
	contentType, body, err := itrace.ParseTracingRequest(req)
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	buf, err := io.ReadAll(body)
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	var (
		zpkmodels []*zpkmodel.SpanModel
		dktrace   itrace.DatakitTrace
	)
	switch contentType {
	case "application/x-protobuf":
		zpkmodels, err = parseZipkinProtobuf3(buf)
	case "application/json":
		err = json.Unmarshal(buf, &zpkmodels)
	default:
		err = fmt.Errorf("zipkin V2 unsupported Content-Type: %s", contentType)
	}

	if err != nil {
		log.Errorf("convert zipkin trace [%s] to datakit trace failed: %s", contentType, err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	dktrace = spanModeleV2ToDkTrace(zpkmodels)
	if len(dktrace) == 0 {
		log.Warn("empty datakit trace")
	} else {
		afterGather.Run(inputName, dktrace, false)
	}

	resp.WriteHeader(http.StatusOK)
}
