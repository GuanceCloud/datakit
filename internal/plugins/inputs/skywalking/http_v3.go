// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package skywalking

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	agentv3 "github.com/GuanceCloud/tracing-protos/skywalking-gen-go/language/agent/v3"
	loggingv3 "github.com/GuanceCloud/tracing-protos/skywalking-gen-go/logging/v3"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"google.golang.org/protobuf/proto"
)

func httpStatusRespFunc(resp http.ResponseWriter, req *http.Request, err error) {
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	resp.WriteHeader(http.StatusOK)
}

func handleSkyTraceV3(resp http.ResponseWriter, req *http.Request) {
	log.Debugf("### receiving trace data from path %s", req.URL.Path)

	var (
		segment *agentv3.SegmentObject
		err     error
	)
	switch req.Header.Get("Content-Type") {
	case "application/x-protobuf":
		segment, err = readSegmentObjectV3(req)
	default:
		err = errors.New("unrecognized HTTP content type")
	}
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	dktrace := parseSegmentObjectV3(segment)
	if len(dktrace) != 0 && afterGatherRun != nil {
		afterGatherRun.Run(inputName, itrace.DatakitTraces{dktrace})
	}

	resp.WriteHeader(http.StatusOK)
}

func readSegmentObjectV3(req *http.Request) (*agentv3.SegmentObject, error) {
	bts, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	segment := &agentv3.SegmentObject{}
	err = proto.Unmarshal(bts, segment)

	return segment, err
}

func (ipt *Input) handleSkyMetricV3(resp http.ResponseWriter, req *http.Request) {
	log.Debugf("### receiving metric data from path %s", req.URL.Path)

	var (
		jvm   *agentv3.JVMMetricCollection
		start = time.Now()
		err   error
	)
	switch req.Header.Get("Content-Type") {
	case "application/x-protobuf":
		jvm, err = readJvmMetricColV3(req)
	default:
		err = errors.New("unrecognized HTTP content type")
	}
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	metrics := processMetricsV3(jvm, start, ipt)
	if len(metrics) != 0 {
		if err := ipt.feeder.Feed(jvmMetricName, point.Metric, metrics, &dkio.Option{CollectCost: time.Since(start)}); err != nil {
			ipt.feeder.FeedLastError(jvmMetricName, err.Error())
		}
	}

	resp.WriteHeader(http.StatusOK)
}

func readJvmMetricColV3(req *http.Request) (*agentv3.JVMMetricCollection, error) {
	bts, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	jvm := &agentv3.JVMMetricCollection{}
	err = proto.Unmarshal(bts, jvm)

	return jvm, err
}

func handleSkyLoggingV3(resp http.ResponseWriter, req *http.Request) {
	log.Debugf("### receiving logging data from path %s", req.URL.Path)

	var (
		logdata *loggingv3.LogData
		err     error
	)
	switch req.Header.Get("Content-Type") {
	case "application/x-protobuf":
		logdata, err = readLoggingV3(req)
	default:
		err = errors.New("unrecognized HTTP content type")
	}
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	if pt, err := processLogV3(logdata); err != nil {
		log.Error(err.Error())
	} else {
		if err = dkio.Feed(logdata.Service, datakit.Logging, []*dkpt.Point{pt}, nil); err != nil {
			log.Error(err.Error())
		}
	}
}

func readLoggingV3(req *http.Request) (*loggingv3.LogData, error) {
	bts, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	logdata := &loggingv3.LogData{}
	err = proto.Unmarshal(bts, logdata)

	return logdata, err
}

func handleProfilingV3(resp http.ResponseWriter, req *http.Request) {
	log.Debugf("### receiving profiling data from path %s", req.URL.Path)

	resp.WriteHeader(http.StatusOK)
}
