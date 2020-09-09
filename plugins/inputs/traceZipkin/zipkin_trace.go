package traceZipkin

import (
	"fmt"
	"net/http"
	"runtime/debug"

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
		log.Errorf("%v", err)
	}
}

func handleZipkinTraceV1(w http.ResponseWriter, r *http.Request) error {
	reqInfo, err := trace.ParseHttpReq(r)
	if err != nil {
		return err
	}

	if reqInfo.ContentType == "application/x-thrift" {
		return parseZipkinThriftV1(reqInfo.Body)
	} else if reqInfo.ContentType == "application/json" {
		return parseZipkinJsonV1(reqInfo.Body)
	} else {
		return fmt.Errorf("Zipkin V1 unsupported Content-Type: %s", reqInfo.ContentType)
	}
}

func handleZipkinTraceV2(w http.ResponseWriter, r *http.Request) error {
	reqInfo, err := trace.ParseHttpReq(r)
	if err != nil {
		return err
	}

	if reqInfo.ContentType == "application/x-protobuf" {
		return parseZipkinProtobufV2(reqInfo.Body)
	} else if reqInfo.ContentType == "application/json" {
		return parseZipkinJsonV2(reqInfo.Body)
	} else {
		return fmt.Errorf("Zipkin V2 unsupported Content-Type: %s", reqInfo.ContentType)
	}
}
