// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package zipkin handle Zipkin APM traces.
package zipkin

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	zpkmodel "github.com/openzipkin/zipkin-go/model"
	zpkprotov2 "github.com/openzipkin/zipkin-go/proto/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/zipkin/compiled/thrift-0.16.0/zipkincore"
)

func httpStatusRespFunc(resp http.ResponseWriter, req *http.Request, err error) {
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
	} else {
		resp.WriteHeader(http.StatusOK)
	}
}

func handleZipkinTraceV1(resp http.ResponseWriter, req *http.Request) {
	log.Debugf("### receiving trace data from path: %s", req.URL.Path)

	pbuf := bufpool.GetBuffer()
	defer bufpool.PutBuffer(pbuf)

	_, err := io.Copy(pbuf, req.Body)
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	param := &itrace.TraceParameters{
		URLPath: apiv1Path,
		Media:   req.Header.Get("Content-Type"),
		Encode:  req.Header.Get("Content-Encoding"),
		Body:    pbuf,
	}
	if err = parseZipkinTraceV1(param); err != nil {
		log.Errorf("### parse zipkin trace v1 failed: %s", err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	resp.WriteHeader(http.StatusOK)
}

func parseZipkinTraceV1(param *itrace.TraceParameters) error {
	var (
		body io.ReadCloser
		err  error
	)
	if param.Encode == "gzip" {
		if body, err = gzip.NewReader(param.Body); err != nil {
			return err
		}
		defer body.Close() // nolint: errcheck
	} else {
		body = io.NopCloser(param.Body)
	}

	var dktrace itrace.DatakitTrace
	switch param.Media {
	case "application/x-thrift":
		var zspans []*zipkincore.Span
		if zspans, err = unmarshalZipkinThriftV1(body); err == nil {
			dktrace = thriftV1SpansToDkTrace(zspans)
		}
	case "application/json":
		var zspans []*ZipkinSpanV1
		if err = json.NewDecoder(body).Decode(&zspans); err == nil {
			dktrace = jsonV1SpansToDkTrace(zspans)
		}
	default:
		err = fmt.Errorf("### zipkin V1 unsupported Content-Type: %s", param.Media)
	}
	if err != nil {
		return err
	}

	if len(dktrace) != 0 && afterGatherRun != nil {
		afterGatherRun.Run(inputName, itrace.DatakitTraces{dktrace}, false)
	}

	return nil
}

func handleZipkinTraceV2(resp http.ResponseWriter, req *http.Request) {
	log.Debugf("### receiving trace data from path: %s", req.URL.Path)

	pbuf := bufpool.GetBuffer()
	defer bufpool.PutBuffer(pbuf)

	_, err := io.Copy(pbuf, req.Body)
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	param := &itrace.TraceParameters{
		URLPath: apiv2Path,
		Media:   req.Header.Get("Content-Type"),
		Encode:  req.Header.Get("Content-Encoding"),
		Body:    pbuf,
	}
	if err = parseZipkinTraceV2(param); err != nil {
		log.Errorf("### parse zipkin trace v2 failed: %s", err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	resp.WriteHeader(http.StatusOK)
}

func parseZipkinTraceV2(param *itrace.TraceParameters) error {
	var (
		buf []byte
		err error
	)
	if param.Encode == "gzip" {
		var body io.ReadCloser
		if body, err = gzip.NewReader(param.Body); err != nil {
			return err
		}
		defer body.Close() // nolint: errcheck

		if buf, err = io.ReadAll(body); err != nil {
			return err
		}
	} else {
		buf = param.Body.Bytes()
	}

	var (
		zpkmodels []*zpkmodel.SpanModel
		dktrace   itrace.DatakitTrace
	)
	switch param.Media {
	case "application/x-protobuf":
		// zpkmodels, err = parseZipkinProtobuf3(buf)
		zpkmodels, err = zpkprotov2.ParseSpans(buf, false)
	case "application/json":
		err = json.Unmarshal(buf, &zpkmodels)
	default:
		err = fmt.Errorf("### zipkin V2 unsupported Content-Type: %s", param.Media)
	}
	if err != nil {
		return err
	}

	dktrace = spanModeleV2ToDkTrace(zpkmodels)
	if len(dktrace) != 0 && afterGatherRun != nil {
		afterGatherRun.Run(inputName, itrace.DatakitTraces{dktrace}, false)
	}

	return nil
}
