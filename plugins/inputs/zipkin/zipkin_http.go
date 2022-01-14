package zipkin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	zipkinmodel "github.com/openzipkin/zipkin-go/model"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
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
		log.Errorf("handleZipkinTraceV1: %s", err)

		io.FeedLastError(inputName, err.Error())
	}
}

func handleZipkinTraceV1(w http.ResponseWriter, r *http.Request) error {
	reqInfo, err := trace.ParseTraceInfo(r)
	if err != nil {
		return err
	}

	var group []*trace.TraceAdapter
	switch reqInfo.ContentType {
	case "application/x-thrift":
		if zspans, err := unmarshalZipkinThriftV1(reqInfo.Body); err != nil {
			return err
		} else {
			group, err = thriftSpansToAdapters(zspans)
			if err != nil {
				log.Errorf("thriftSpansToAdapters: %s", err)

				return err
			}
		}
	case "application/json":
		zspans := []*ZipkinSpanV1{}
		if err := json.Unmarshal(reqInfo.Body, &zspans); err != nil {
			log.Errorf("json.Unmarshal: %s", err)

			return err
		} else {
			group, err = jsonV1SpansToAdapters(zspans)
			if err != nil {
				log.Errorf("jsonV1SpansToAdapters: %s", err)

				return err
			}
		}
	default:
		return fmt.Errorf("zipkin V1 unsupported Content-Type: %s", reqInfo.ContentType)
	}

	if len(group) != 0 {
		trace.MkLineProto(group, inputName)
	} else {
		log.Debug("empty zipkin v1 spans")
	}

	return nil
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
		log.Errorf("handleZipkinTraceV2: %v", err)
		io.FeedLastError(inputName, err.Error())
	}
}

func handleZipkinTraceV2(w http.ResponseWriter, r *http.Request) error {
	reqInfo, err := trace.ParseTraceInfo(r)
	if err != nil {
		return err
	}

	var group []*trace.TraceAdapter
	switch reqInfo.ContentType {
	case "application/x-protobuf":
		zspans, err := parseZipkinProtobuf3(reqInfo.Body)
		if err != nil {
			log.Errorf("parseZipkinProtobuf3: %s", err.Error())

			return err
		}
		if group, err = protobufSpansToAdapters(zspans); err != nil {
			log.Errorf("protobufSpansToAdapters: %s", err.Error())

			return err
		}
	case "application/json":
		zspans := []*zipkinmodel.SpanModel{}
		if err := json.Unmarshal(reqInfo.Body, &zspans); err != nil {
			log.Errorf("json.Unmarshal: %s", err.Error())

			return err
		}
		if group, err = parseZipkinJsonV2(zspans); err != nil {
			log.Errorf("parseZipkinJsonV2: %s", err.Error())

			return err
		}
	default:
		return fmt.Errorf("zipkin V2 unsupported Content-Type: %s", reqInfo.ContentType)
	}

	if len(group) != 0 {
		trace.MkLineProto(group, inputName)
	} else {
		log.Warn("empty zipkin v2 spans")
	}

	return nil
}
