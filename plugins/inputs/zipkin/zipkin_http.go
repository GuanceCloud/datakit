// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package zipkin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	zpkmodel "github.com/openzipkin/zipkin-go/model"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/zipkin/corev1"
)

type parameters struct {
	media string
	body  *bytes.Buffer
}

func handleZipkinTraceV1(resp http.ResponseWriter, req *http.Request) {
	media, encode, buf, err := itrace.ParseTracerRequest(req)
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	param := &parameters{
		media: media,
		body:  bytes.NewBuffer(buf),
	}

	log.Debugf("### path: %s, Content-Type: %s, Encode-Type: %s, body-size: %s", req.URL.Path, media, encode, len(buf))

	if wpool == nil {
		if err = parseZipkinTraceV1(param); err != nil {
			log.Error(err.Error())
			resp.WriteHeader(http.StatusBadRequest)

			return
		}
	} else {
		job, err := workerpool.NewJob(workerpool.WithInput(param),
			workerpool.WithProcess(parseZipkinTraceV1Adapter),
			workerpool.WithProcessCallback(func(input, output interface{}, cost time.Duration, isTimeout bool) {
				log.Debugf("### job status: input: %v, output: %v, cost: %dms, timeout: %v", input, output, cost/time.Millisecond, isTimeout)
			}),
			workerpool.WithTimeout(jobTimeout),
		)
		if err != nil {
			log.Error(err.Error())
			resp.WriteHeader(http.StatusBadRequest)

			return
		}

		if err = wpool.MoreJob(job); err != nil {
			log.Error(err)
			resp.WriteHeader(http.StatusTooManyRequests)

			return
		}
	}

	resp.WriteHeader(http.StatusOK)
}

func parseZipkinTraceV1Adapter(input interface{}) (output interface{}) {
	param, ok := input.(*parameters)
	if !ok {
		return errors.New("type assertion failed")
	}

	return parseZipkinTraceV1(param)
}

func parseZipkinTraceV1(param *parameters) error {
	var (
		dktrace itrace.DatakitTrace
		err     error
	)
	switch param.media {
	case "application/x-thrift":
		var zspans []*corev1.Span
		if zspans, err = unmarshalZipkinThriftV1(io.NopCloser(param.body)); err == nil {
			dktrace = thriftV1SpansToDkTrace(zspans)
		}
	case "application/json":
		var zspans []*ZipkinSpanV1
		if err = json.NewDecoder(param.body).Decode(&zspans); err == nil {
			dktrace = jsonV1SpansToDkTrace(zspans)
		}
	default:
		err = fmt.Errorf("### zipkin V1 unsupported Content-Type: %s", param.media)
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
	media, _, buf, err := itrace.ParseTracerRequest(req)
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	param := &parameters{
		media: media,
		body:  bytes.NewBuffer(buf),
	}

	log.Debugf("### path: %s, Content-Type: %s, body-size: %s", req.URL.Path, media, len(buf))

	if wpool == nil {
		if err = parseZipkinTraceV2(param); err != nil {
			log.Error(err.Error())
			resp.WriteHeader(http.StatusBadRequest)

			return
		}
	} else {
		job, err := workerpool.NewJob(workerpool.WithInput(param),
			workerpool.WithProcess(parseZipkinTraceV2Adapter),
			workerpool.WithProcessCallback(func(input, output interface{}, cost time.Duration, isTimeout bool) {
				log.Debugf("### job status: input: %v, output: %v, cost: %dms, timeout: %v", input, output, cost/time.Millisecond, isTimeout)
			}),
			workerpool.WithTimeout(jobTimeout),
		)
		if err != nil {
			log.Error(err.Error())
			resp.WriteHeader(http.StatusBadRequest)

			return
		}

		if err = wpool.MoreJob(job); err != nil {
			log.Error(err)
			resp.WriteHeader(http.StatusTooManyRequests)

			return
		}
	}

	resp.WriteHeader(http.StatusOK)
}

func parseZipkinTraceV2Adapter(input interface{}) (output interface{}) {
	param, ok := input.(*parameters)
	if !ok {
		return errors.New("type assertion failed")
	}

	return parseZipkinTraceV2(param)
}

func parseZipkinTraceV2(param *parameters) error {
	var (
		zpkmodels []*zpkmodel.SpanModel
		dktrace   itrace.DatakitTrace
		err       error
	)
	switch param.media {
	case "application/x-protobuf":
		zpkmodels, err = parseZipkinProtobuf3(param.body.Bytes())
	case "application/json":
		err = json.Unmarshal(param.body.Bytes(), &zpkmodels)
	default:
		err = fmt.Errorf("### zipkin V2 unsupported Content-Type: %s", param.media)
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
