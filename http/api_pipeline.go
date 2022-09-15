// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/multiline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	defaultNotSetMakePoint = "default_not_set"
	defaultNotSetService   = "default_not_set_service"

	// https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/608
	maxLenPipelineBefore = 3 * 4096
	maxLenDataBefore     = 3 * 8192
)

type pipelineDebugRequest struct {
	Pipeline   string `json:"pipeline"`
	ScriptName string `json:"script_name"`
	Category   string `json:"category"`
	Data       string `json:"data"`
	Multiline  string `json:"multiline"`
	Encode     string `json:"encode"`
	Benchmark  bool   `json:"benchmark"`
}

type pipelineDebugResult struct {
	Measurement string                 `json:"measurement"`
	Tags        map[string]string      `json:"tags"`
	Fields      map[string]interface{} `json:"fields"`
	Time        int64                  `json:"time"`
	TimeNS      int64                  `json:"time_ns"`
	Dropped     bool                   `json:"dropped"`
}

type pipelineDebugResponse struct {
	Cost         string                 `json:"cost"`
	Benchmark    string                 `json:"benchmark"`
	ErrorMessage string                 `json:"error_msg"`
	PLResults    []*pipelineDebugResult `json:"plresults"`
}

// func plAPIDebugCallback(ret *pipeline.Result) (*pipeline.Result, error) {
// 	return pipeline.ResultUtilsLoggingProcessor(ret, false, nil), nil
// }

func apiDebugPipelineHandler(w http.ResponseWriter, req *http.Request, whatever ...interface{}) (interface{}, error) {
	tid := req.Header.Get(uhttp.XTraceId)

	reqDebug, err := getAPIDebugPipelineRequest(req)
	if err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, err
	}

	if err := checkRequest(reqDebug); err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, err
	}
	category := getPointCategory(reqDebug.Category)
	if category == "" {
		err := uhttp.Error(ErrInvalidCategory, "invalid category")
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, err
	}

	//------------------------------------------------------------------
	// -- pipeline debug procedure start --

	// STEP 1: check pipeline
	decodePipeline, err := base64.StdEncoding.DecodeString(reqDebug.Pipeline)
	if err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, uhttp.Error(ErrInvalidPipeline, err.Error())
	}

	scriptInfo, err := pipeline.NewPipeline(category, reqDebug.ScriptName+".p", string(decodePipeline))
	if err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, uhttp.Error(ErrCompiledFailed, err.Error())
	}

	// STEP 2: get logging data
	data, err := getDecodeData(reqDebug)
	if err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, uhttp.Error(ErrInvalidData, err.Error())
	}

	dataLines, err := getDataLines(data, reqDebug.Multiline)
	if err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, uhttp.Error(ErrInvalidData, err.Error())
	}

	opt := &point.PointOption{
		Category: category,
		Time:     time.Now(),
	}

	// STEP 3: pipeline processing
	start := time.Now()
	res := []*pipeline.Result{}
	for _, line := range dataLines {
		switch category {
		case datakit.Logging:
			pt, err := point.NewPoint(reqDebug.ScriptName, nil, map[string]interface{}{pipeline.FieldMessage: line}, opt)
			if err != nil {
				l.Errorf("[%s] %s", tid, err.Error())
				return nil, uhttp.Error(ErrInvalidData, err.Error())
			}
			if single := getSinglePointResult(scriptInfo, pt, opt); single != nil {
				res = append(res, single)
			}
		default:
			pts, err := lp.ParsePoints([]byte(line), nil)
			if err != nil {
				l.Errorf("[%s] %s", tid, err.Error())
				return nil, uhttp.Error(ErrInvalidData, err.Error())
			}
			newPts := point.WrapPoint(pts)
			for _, pt := range newPts {
				if single := getSinglePointResult(scriptInfo, pt, opt); single != nil {
					res = append(res, single)
				}
			}
		}
	}

	// STEP 4 (optional): benchmark
	var benchmarkResult testing.BenchmarkResult
	if reqDebug.Benchmark {
		benchmarkResult = testing.Benchmark(func(b *testing.B) {
			b.Helper()
			for n := 0; n < b.N; n++ {
				for _, line := range dataLines {
					pt, _ := point.NewPoint(defaultNotSetMakePoint,
						nil,
						map[string]interface{}{pipeline.FieldMessage: line}, opt)
					_, _, _ = scriptInfo.Run(pt, nil, opt)
				}
			}
		})
	}

	// -- pipeline debug procedure end --
	//------------------------------------------------------------------

	return getReturnResult(start, res, reqDebug, &benchmarkResult), nil
}

func getSinglePointResult(scriptInfo *pipeline.Pipeline, pt *point.Point, opt *point.PointOption) *pipeline.Result {
	pt, drop, err := scriptInfo.Run(pt, nil, opt)
	if err != nil || pt == nil {
		return nil
	}
	fields, err := pt.Fields()
	if err != nil {
		return nil
	}
	tags := pt.Tags()

	if _, ok := tags["service"]; !ok {
		tags["service"] = defaultNotSetService
	}

	return &pipeline.Result{
		Output: &pipeline.Output{
			Drop:        drop,
			Measurement: pt.Name(),
			Time:        pt.Time(),
			Tags:        tags,
			Fields:      fields,
		},
	}
}

func getReturnResult(start time.Time, res []*pipeline.Result,
	reqDebug *pipelineDebugRequest,
	benchmarkResult *testing.BenchmarkResult,
) *pipelineDebugResponse {
	var returnres pipelineDebugResponse
	cost := time.Since(start)
	returnres.Cost = cost.String()

	var results []*pipelineDebugResult
	for _, v := range res {
		var tm, tmns int64
		if t, err := v.GetTime(); err != nil {
			l.Debugf("GetTime failed: %v", err)
		} else {
			tmstr := fmt.Sprintf("%d", t.Unix())
			tmnsstr := fmt.Sprintf("%d", t.UnixNano())
			nsstr := strings.ReplaceAll(tmnsstr, tmstr, "")

			n, err := strconv.ParseInt(nsstr, 10, 64)
			if err != nil {
				l.Debugf("strconv.ParseInt failed: %v", err)
			} else {
				tm = t.Unix()
				tmns = n
			}
		}

		results = append(results, &pipelineDebugResult{
			Measurement: v.GetMeasurement(),
			Tags:        v.GetTags(),
			Fields:      v.GetFields(),
			Time:        tm,
			TimeNS:      tmns,
			Dropped:     v.IsDropped(),
		})
	}
	returnres.PLResults = results
	if reqDebug.Benchmark {
		returnres.Benchmark = benchmarkResult.String()
	}

	return &returnres
}

func getDataLines(originBytes []byte, pattern string) ([]string, error) {
	var outArr []string

	multi, err := multiline.New([]string{pattern})
	if err != nil {
		return nil, err
	}

	lines, err := getBytesLines(originBytes)
	if err != nil {
		return nil, err
	}

	for _, v := range lines {
		res := multi.ProcessLineString(v)
		if len(res) != 0 {
			outArr = append(outArr, res)
		}
	}

	if multi.BuffLength() > 0 {
		outArr = append(outArr, multi.FlushString())
	}

	return outArr, nil
}

func getBytesLines(bys []byte) ([]string, error) {
	var arr []string
	scanner := bufio.NewScanner(bytes.NewBuffer(bys))
	for scanner.Scan() {
		arr = append(arr, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return arr, nil
}

func checkRequest(reqDebug *pipelineDebugRequest) error {
	switch reqDebug.Category {
	case datakit.CategoryMetric,
		datakit.CategoryNetwork,
		datakit.CategoryKeyEvent,
		datakit.CategoryObject,
		datakit.CategoryCustomObject,
		datakit.CategoryLogging,
		datakit.CategoryTracing,
		datakit.CategoryRUM,
		datakit.CategorySecurity:
	default:
		return uhttp.Error(ErrInvalidCategory, "invalid category")
	}

	if len(reqDebug.Pipeline) > maxLenPipelineBefore ||
		len(reqDebug.ScriptName) > maxLenPipelineBefore ||
		len(reqDebug.Category) > maxLenPipelineBefore ||
		len(reqDebug.Data) > maxLenDataBefore ||
		len(reqDebug.Multiline) > maxLenPipelineBefore ||
		len(reqDebug.Encode) > maxLenPipelineBefore {
		return uhttp.Error(ErrBadReq, "too large")
	}

	return nil
}

func getAPIDebugPipelineRequest(req *http.Request) (*pipelineDebugRequest, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, uhttp.Error(ErrInvalidRequest, err.Error())
	}

	var reqDebug pipelineDebugRequest
	if err := json.Unmarshal(body, &reqDebug); err != nil {
		return nil, uhttp.Error(ErrInvalidRequest, err.Error())
	}

	return &reqDebug, nil
}

func getPointCategory(category string) string {
	switch category {
	case datakit.Metric, datakit.MetricDeprecated, datakit.CategoryMetric, "metrics":
		return datakit.Metric
	case datakit.Network, datakit.CategoryNetwork:
		return datakit.Network
	case datakit.KeyEvent, datakit.CategoryKeyEvent:
		return datakit.KeyEvent
	case datakit.Object, datakit.CategoryObject:
		return datakit.Object
	case datakit.CustomObject, datakit.CategoryCustomObject:
		return datakit.CustomObject
	case datakit.Tracing, datakit.CategoryTracing:
		return datakit.Tracing
	case datakit.RUM, datakit.CategoryRUM:
		return datakit.RUM
	case datakit.Security, datakit.CategorySecurity:
		return datakit.Security
	case datakit.Logging, datakit.CategoryLogging:
		return datakit.Logging
	case datakit.Profiling, datakit.CategoryProfiling:
		return datakit.Profiling
	default:
		return ""
	}
}

//------------------------------------------------------------------------------
// -- decoding functions' area start --

func getDecodeData(reqDebug *pipelineDebugRequest) ([]byte, error) {
	decodeData, err := base64.StdEncoding.DecodeString(reqDebug.Data)
	if err != nil {
		return nil, uhttp.Error(ErrInvalidData, err.Error())
	}

	var data []byte

	var encode string
	if reqDebug.Encode != "" {
		encode = strings.ToLower(reqDebug.Encode)
	}
	switch encode {
	case "gbk", "gb18030":
		data, err = GbToUtf8(decodeData, encode)
		if err != nil {
			return nil, uhttp.Error(ErrInvalidData, err.Error())
		}
	case "utf8", "utf-8":
		fallthrough
	default:
		data = decodeData
	}

	return data, nil
}

// GbToUtf8 Gb to UTF-8.
func GbToUtf8(s []byte, encoding string) ([]byte, error) {
	var t transform.Transformer
	switch encoding {
	case "gbk":
		t = simplifiedchinese.GBK.NewDecoder()
	case "gb18030":
		t = simplifiedchinese.GB18030.NewDecoder()
	}
	reader := transform.NewReader(bytes.NewReader(s), t)
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

// -- decode functions' area end --
//------------------------------------------------------------------------------
