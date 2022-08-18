// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package zipkin

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	zpkmodel "github.com/openzipkin/zipkin-go/model"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/zipkin/corev1"
)

type parameters struct {
	media  string
	encode string
	body   *bytes.Buffer
}

func handleZipkinTraceV1(resp http.ResponseWriter, req *http.Request) {
	log.Debugf("### received tracing data from path: %s", req.URL.Path)

	var (
		readbodycost = time.Now()
		enterWPoolOK = false
	)
	pbuf := bufpool.GetBuffer()
	defer func() {
		if !enterWPoolOK {
			bufpool.PutBuffer(pbuf)
		}
	}()

	_, err := io.Copy(pbuf, req.Body)
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	param := &parameters{
		media:  req.Header.Get("Content-Type"),
		encode: req.Header.Get("Content-Encoding"),
		body:   pbuf,
	}

	log.Debugf("### path: %s, Content-Type: %s, Encode-Type: %s, body-size: %dkb, read-body-cost: %dms",
		req.URL.Path, param.media, param.encode, pbuf.Len()>>10, time.Since(readbodycost)/time.Millisecond)

	if wpool == nil {
		if err = parseZipkinTraceV1(param); err != nil {
			log.Error(err.Error())
			resp.WriteHeader(http.StatusBadRequest)

			return
		}
	} else {
		job, err := workerpool.NewJob(workerpool.WithInput(param),
			workerpool.WithProcess(parseZipkinTraceV1Adapter),
			workerpool.WithProcessCallback(func(input, output interface{}, cost time.Duration) {
				if param, ok := input.(*parameters); ok {
					bufpool.PutBuffer(param.body)
				}
				log.Debugf("### job status: input: %v, output: %v, cost: %dms", input, output, cost/time.Millisecond)
			}),
		)
		if err != nil {
			log.Error(err.Error())
			resp.WriteHeader(http.StatusBadRequest)

			return
		}

		if err = wpool.MoreJob(job); err != nil {
			log.Error(err.Error())
			resp.WriteHeader(http.StatusTooManyRequests)

			return
		}
		enterWPoolOK = true
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
		body io.ReadCloser
		err  error
	)
	if param.encode == "gzip" {
		if body, err = gzip.NewReader(param.body); err != nil {
			return err
		}
		defer body.Close() // nolint: errcheck
	} else {
		body = io.NopCloser(param.body)
	}

	var dktrace itrace.DatakitTrace
	switch param.media {
	case "application/x-thrift":
		var zspans []*corev1.Span
		if zspans, err = unmarshalZipkinThriftV1(body); err == nil {
			dktrace = thriftV1SpansToDkTrace(zspans)
		}
	case "application/json":
		var zspans []*ZipkinSpanV1
		if err = json.NewDecoder(body).Decode(&zspans); err == nil {
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
	log.Debugf("### received tracing data from path: %s", req.URL.Path)

	var (
		readbodycost = time.Now()
		enterWPoolOK = false
	)
	pbuf := bufpool.GetBuffer()
	defer func() {
		if !enterWPoolOK {
			bufpool.PutBuffer(pbuf)
		}
	}()

	_, err := io.Copy(pbuf, req.Body)
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	param := &parameters{
		media:  req.Header.Get("Content-Type"),
		encode: req.Header.Get("Content-Encoding"),
		body:   pbuf,
	}

	log.Debugf("### path: %s, Content-Type: %s, Encode-Type: %s, body-size: %dkb, read-body-cost: %dms",
		req.URL.Path, param.media, param.encode, pbuf.Len()>>10, time.Since(readbodycost)/time.Millisecond)

	if wpool == nil {
		if err = parseZipkinTraceV2(param); err != nil {
			log.Error(err.Error())
			resp.WriteHeader(http.StatusBadRequest)

			return
		}
	} else {
		job, err := workerpool.NewJob(workerpool.WithInput(param),
			workerpool.WithProcess(parseZipkinTraceV2Adapter),
			workerpool.WithProcessCallback(func(input, output interface{}, cost time.Duration) {
				if param, ok := input.(*parameters); ok {
					bufpool.PutBuffer(param.body)
				}
				log.Debugf("### job status: input: %v, output: %v, cost: %dms", input, output, cost/time.Millisecond)
			}),
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
		enterWPoolOK = true
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
		buf []byte
		err error
	)
	if param.encode == "gzip" {
		var body io.ReadCloser
		if body, err = gzip.NewReader(param.body); err != nil {
			return err
		}
		defer body.Close() // nolint: errcheck

		if buf, err = io.ReadAll(body); err != nil {
			return err
		}
	} else {
		buf = param.body.Bytes()
	}

	var (
		zpkmodels []*zpkmodel.SpanModel
		dktrace   itrace.DatakitTrace
	)
	switch param.media {
	case "application/x-protobuf":
		zpkmodels, err = parseZipkinProtobuf3(buf)
	case "application/json":
		err = json.Unmarshal(buf, &zpkmodels)
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
