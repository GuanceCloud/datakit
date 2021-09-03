package traceZipkin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	zipkinmodel "github.com/openzipkin/zipkin-go/model"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

func ZipkinTraceHandleV1(w http.ResponseWriter, r *http.Request) {
	log.Debugf("trace handle with path: %s", r.URL.Path)
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Stack crash: %v", r)
			log.Errorf("Stack info :%s", string(debug.Stack()))
		}
	}()

	if err := handleZipkinTraceV1(w, r); err != nil {
		io.FeedLastError(inputName, err.Error())
		log.Errorf("%v", err)
	}
}

func ZipkinTraceHandleV2(w http.ResponseWriter, r *http.Request) {
	log.Debugf("trace handle with path: %s", r.URL.Path)
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Stack crash: %v", r)
			log.Errorf("Stack info :%s", string(debug.Stack()))
		}
	}()

	if err := handleZipkinTraceV2(w, r); err != nil {
		io.FeedLastError(inputName, err.Error())
		log.Errorf("%v", err)
	}
}

func handleZipkinTraceV1(w http.ResponseWriter, r *http.Request) error {
	reqInfo, err := trace.ParseHttpReq(r)
	if err != nil {
		return err
	}

	var group []*trace.TraceAdapter
	if reqInfo.ContentType == "application/x-thrift" {
		if zspans, err := unmarshalZipkinThriftV1(reqInfo.Body); err != nil {
			return err
		} else {
			group, err = thriftSpansToAdapters(zspans, zpkThriftV1Filters...)
		}
	} else if reqInfo.ContentType == "application/json" {
		zspans := []*ZipkinSpanV1{}
		if err := json.Unmarshal(reqInfo.Body, &zspans); err != nil {
			return err
		} else {
			group, err = jsonV1SpansToAdapters(zspans, zpkJsonV1Filters...)
		}
	} else {
		return fmt.Errorf("Zipkin V1 unsupported Content-Type: %s", reqInfo.ContentType)
	}
	if err != nil {
		log.Error(err)

		return err
	}

	if len(group) != 0 {
		trace.MkLineProto(group, inputName)
	} else {
		log.Debug("empty zipkin v1 spans")
	}

	return nil
}

func handleZipkinTraceV2(w http.ResponseWriter, r *http.Request) error {
	reqInfo, err := trace.ParseHttpReq(r)
	if err != nil {
		return err
	}

	var group []*trace.TraceAdapter
	if reqInfo.ContentType == "application/x-protobuf" {
		if zspans, err := parseZipkinProtobuf3(reqInfo.Body); err != nil {
			return err
		} else {
			group, err = protobufSpansToAdapters(zspans, zpkProtoBufV2Filters...)
		}
	} else if reqInfo.ContentType == "application/json" {
		zspans := []*zipkinmodel.SpanModel{}
		if err := json.Unmarshal(reqInfo.Body, &zspans); err != nil {
			return err
		} else {
			group, err = parseZipkinJsonV2(zspans, zpkJsonV2Filters...)
		}
	} else {
		return fmt.Errorf("Zipkin V2 unsupported Content-Type: %s", reqInfo.ContentType)
	}
	if err != nil {
		log.Error(err)

		return err
	}

	if len(group) != 0 {
		trace.MkLineProto(group, inputName)
	} else {
		log.Debug("empty zipkin v2 spans")
	}

	return nil
}
