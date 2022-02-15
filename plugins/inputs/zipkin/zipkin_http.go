package zipkin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	zpkmodel "github.com/openzipkin/zipkin-go/model"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
)

func ZipkinTraceHandleV1(resp http.ResponseWriter, req *http.Request) {
	log.Debugf("%s: listen on path: %s", inputName, req.URL.Path)

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Stack crash: %v", r)
			log.Errorf("Stack info :%s", string(debug.Stack()))
		}
	}()

	if err := handleZipkinTraceV1(req); err != nil {
		log.Errorf("handleZipkinTraceV1: %s", err)

		io.FeedLastError(inputName, err.Error())
	}
}

func handleZipkinTraceV1(req *http.Request) error {
	reqInfo, err := itrace.ParseTraceInfo(req)
	if err != nil {
		return err
	}

	var dktrace itrace.DatakitTrace
	switch reqInfo.ContentType {
	case "application/x-thrift":
		if zspans, err := unmarshalZipkinThriftV1(reqInfo.Body); err != nil {
			return err
		} else {
			dktrace = thriftSpansToDkTrace(zspans)
		}
	case "application/json":
		var zspans []*ZipkinSpanV1
		if err := json.Unmarshal(reqInfo.Body, &zspans); err != nil {
			log.Errorf("json.Unmarshal: %s", err)

			return err
		} else {
			dktrace, err = jsonV1SpansToDkTrace(zspans)
			if err != nil {
				log.Errorf("jsonV1SpansToAdapters: %s", err)

				return err
			}
		}
	default:
		return fmt.Errorf("zipkin V1 unsupported Content-Type: %s", reqInfo.ContentType)
	}

	if len(dktrace) == 0 {
		log.Warn("empty datakit trace")
	} else {
		afterGather.Run(inputName, dktrace, false)
	}

	return nil
}

func ZipkinTraceHandleV2(resp http.ResponseWriter, req *http.Request) {
	log.Debugf("%s: listen on path: %s", inputName, req.URL.Path)

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Stack crash: %v", r)
			log.Errorf("Stack info :%s", string(debug.Stack()))
		}
	}()

	if err := handleZipkinTraceV2(req); err != nil {
		log.Errorf("handleZipkinTraceV2: %v", err)
		io.FeedLastError(inputName, err.Error())
	}
}

func handleZipkinTraceV2(req *http.Request) error {
	reqInfo, err := itrace.ParseTraceInfo(req)
	if err != nil {
		return err
	}

	var (
		zpkmodels []*zpkmodel.SpanModel
		dktrace   itrace.DatakitTrace
	)
	switch reqInfo.ContentType {
	case "application/x-protobuf":
		if zpkmodels, err = parseZipkinProtobuf3(reqInfo.Body); err == nil {
			dktrace, err = spanModelsToDkTrace(zpkmodels)
		}
	case "application/json":
		if err = json.Unmarshal(reqInfo.Body, &zpkmodels); err == nil {
			dktrace, err = spanModelsToDkTrace(zpkmodels)
		}
	default:
		return fmt.Errorf("zipkin V2 unsupported Content-Type: %s", reqInfo.ContentType)
	}

	if err != nil {
		log.Errorf("convert trace to adapters failed: %s", err.Error())

		return err
	}

	if len(dktrace) == 0 {
		log.Warn("empty datakit trace")
	} else {
		afterGather.Run(inputName, dktrace, false)
	}

	return nil
}
