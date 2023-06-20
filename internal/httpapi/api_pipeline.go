// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	lp "github.com/GuanceCloud/cliutils/lineproto"
	uhttp "github.com/GuanceCloud/cliutils/network/http"

	clipt "github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/platypus/pkg/errchain"
	"github.com/GuanceCloud/platypus/pkg/token"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plmap"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	defaultEncode = "utf-8"

	// https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/608
	// https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/1128

	// 限制单个 pipeline 文件大小.
	PipelineScriptByteSizeLimit = 1 * 1024 * 1024

	// 由于日志采集器最大允许 32 MB，此处限制为 32 MB, 行协议暂设与日志大小相同.
	DataByteSizeLimit = 32 * 1024 * 1024
)

// request body.
type pipelineDebugRequest struct {
	Pipeline   map[string]map[string]string `json:"pipeline"`
	ScriptName string                       `json:"script_name"`
	Category   string                       `json:"category"`

	Data      []string `json:"data"`
	Multiline string   `json:"multiline"`
	Encode    string   `json:"encode"`
	Benchmark bool     `json:"benchmark"`
	Timezone  string   `json:"timezone"`
}

// response body.
type pipelineDebugResponse struct {
	Cost      string             `json:"cost"`
	Benchmark string             `json:"benchmark"`
	PlErrors  []errchain.PlError `json:"pl_errors"`
	PLResults []pipelineResult   `json:"plresults"`
}

type PlRetPoint struct {
	Name   string                 `json:"name"`
	Tags   map[string]string      `json:"tags"`
	Fields map[string]interface{} `json:"fields"`
	Time   int64                  `json:"time"`
	TimeNS int64                  `json:"time_ns"`
}

type pipelineResult struct {
	Point   *PlRetPoint `json:"point"`
	Dropped bool        `json:"dropped"`

	RunError *errchain.PlError `json:"run_error"`
}

func apiPipelineDebugHandler(w http.ResponseWriter, req *http.Request, whatever ...interface{}) (interface{}, error) {
	tid := req.Header.Get(uhttp.XTraceID)

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, uhttp.Error(ErrInvalidRequest, err.Error())
	}

	var reqBody pipelineDebugRequest
	if err := json.Unmarshal(body, &reqBody); err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, uhttp.Error(ErrInvalidRequest, err.Error())
	}

	category := normalizeCategory(reqBody.Category)
	if category == clipt.UnknownCategory {
		err := uhttp.Error(ErrInvalidCategory, "invalid category")
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, err
	}

	// get all base64 encoded scripts for the current category
	if len(reqBody.Pipeline) == 0 ||
		len(reqBody.Pipeline[reqBody.Category]) == 0 {
		err := uhttp.Error(ErrInvalidPipeline, "invalid pipeline")
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, err
	}

	scriptContent := reqBody.Pipeline[reqBody.Category]
	if _, ok := scriptContent[reqBody.ScriptName]; !ok {
		err := uhttp.Error(ErrInvalidPipeline, "invalid pipeline")
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, err
	}

	// decode pipeline script content
	// script name will automatically add the file suffix
	scriptContent, err = decodePipeline(category, scriptContent)
	if err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, uhttp.Error(ErrInvalidData, err.Error())
	}

	// parse pipeline script
	plRunner, err := parsePipeline(category, reqBody.ScriptName+".p", scriptContent)
	if err != nil {
		var plerrs []errchain.PlError
		if plerr, ok := err.(*errchain.PlError); ok { //nolint:errorlint
			plerrs = append(plerrs, *plerr)
		} else {
			plerrs = append(plerrs, *errchain.NewErr(
				reqBody.ScriptName+".p",
				token.LnColPos{
					Pos: 0,
					Ln:  1,
					Col: 1,
				},
				err.Error(),
			))
		}
		return &pipelineDebugResponse{
			PlErrors: plerrs,
		}, nil
	}

	// decode log or line protocol data
	// conv data to point
	ptName := pointName(category.String(), reqBody.ScriptName)
	pts, err := decodeDataAndConv2Point(category, ptName, reqBody.Encode,
		reqBody.Data)
	if err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, uhttp.Error(ErrInvalidData, err.Error())
	}

	opt := &point.PointOption{
		Category: category.URL(),
		Time:     time.Now(),
	}

	var runResult []pipelineResult

	start := time.Now()

	ptsLi := &pldebugFeed{}

	buks := plmap.NewAggBuks(ptsLi.uploadfn)
	for _, pt := range pts {
		// run pipeline
		pt, drop, err := plRunner.Run(category, pt, nil, opt, newPlTestSingal(), buks)

		if err != nil {
			plerr, ok := err.(*errchain.PlError) //nolint:errorlint
			if !ok {
				plerr = errchain.NewErr(reqBody.ScriptName+".p", token.LnColPos{
					Pos: 0,
					Ln:  1,
					Col: 1,
				}, err.Error())
			}

			runResult = append(runResult, pipelineResult{
				RunError: plerr,
			})
		} else {
			fields, err := pt.Fields()
			if err != nil {
				l.Errorf("Fields: %s", err.Error())
				return nil, err
			}

			runResult = append(runResult, pipelineResult{
				Point: &PlRetPoint{
					Name:   pt.Name(),
					Tags:   pt.Tags(),
					Fields: fields,
					Time:   pt.Time().Unix(),
					TimeNS: int64(pt.Time().Nanosecond()),
				},
				Dropped: drop,
			})
		}
	}
	buks.StopAllBukScanner()
	ptsLi.Lock()
	defer ptsLi.Unlock()
	for _, pt := range ptsLi.d {
		fields, err := pt.Fields()
		if err != nil {
			l.Errorf("Fields: %s", err.Error())
			return nil, err
		}
		runResult = append(runResult, pipelineResult{
			Point: &PlRetPoint{
				Name:   pt.Name(),
				Tags:   pt.Tags(),
				Fields: fields,
				Time:   pt.Time().Unix(),
				TimeNS: int64(pt.Time().Nanosecond()),
			},
		})
	}

	var benchmarkInfo string
	if reqBody.Benchmark && len(pts) > 0 {
		benchmarkInfo = benchPipeline(plRunner, category, pts[0], opt)
	}

	return &pipelineDebugResponse{
		PLResults: runResult,
		Benchmark: benchmarkInfo,
		Cost:      time.Since(start).String(),
	}, nil
}

func parsePipeline(category clipt.Category, scriptName string, scripts map[string]string) (*pipeline.Pipeline, error) {
	success, faild := pipeline.NewPipelineMulti(category, scripts, nil)
	if err, ok := faild[scriptName]; ok && err != nil {
		return nil, err
	}
	if pl, ok := success[scriptName]; !ok {
		return nil, uhttp.Error(ErrInvalidPipeline, "invalid pipeline")
	} else {
		return pl, nil
	}
}

func normalizeCategory(category string) clipt.Category {
	switch category {
	case datakit.Metric, datakit.MetricDeprecated, datakit.CategoryMetric:
		return clipt.Metric
	case datakit.Network, datakit.CategoryNetwork:
		return clipt.Network
	case datakit.KeyEvent, datakit.CategoryKeyEvent:
		return clipt.KeyEvent
	case datakit.Object, datakit.CategoryObject:
		return clipt.Object
	case datakit.CustomObject, datakit.CategoryCustomObject:
		return clipt.CustomObject
	case datakit.Tracing, datakit.CategoryTracing:
		return clipt.Tracing
	case datakit.RUM, datakit.CategoryRUM:
		return clipt.RUM
	case datakit.Security, datakit.CategorySecurity:
		return clipt.Security
	case datakit.Logging, datakit.CategoryLogging:
		return clipt.Logging
	case datakit.Profiling, datakit.CategoryProfiling:
		return clipt.Profiling
	default:
		return clipt.UnknownCategory
	}
}

func pointName(category, name string) string {
	switch category {
	case datakit.RUM:
		return datakit.CategoryRUM
	case datakit.Security:
		return datakit.CategorySecurity
	case datakit.Tracing:
		return datakit.CategoryTracing
	case datakit.Profiling:
		return datakit.Profiling
	default:
		return name
	}
}

func benchPipeline(runner *pipeline.Pipeline, cat clipt.Category, pt *point.Point, ptOPt *point.PointOption) string {
	benchmarkResult := testing.Benchmark(func(b *testing.B) {
		b.Helper()
		for n := 0; n < b.N; n++ {
			_, _, _ = runner.Run(cat, pt, nil, ptOPt, newPlTestSingal())
		}
	})
	return benchmarkResult.String()
}

//------------------------------------------------------------------------------
// -- decoding functions' area start --

func decodePipeline(category clipt.Category, scripts map[string]string) (map[string]string, error) {
	decodedScripts := map[string]string{}
	for scriptName, scriptContent := range scripts {
		scriptContent, err := decodeBase64Content(scriptContent, defaultEncode)
		if err != nil {
			return nil, err
		}
		if len(scriptContent) > PipelineScriptByteSizeLimit {
			return nil, fmt.Errorf("script size exceeds 1MB limit")
		}
		// since the incoming script name has no suffix, add it here
		decodedScripts[scriptName+".p"] = scriptContent
	}
	return decodedScripts, nil
}

func decodeDataAndConv2Point(category clipt.Category, name, encode string, data []string) ([]*point.Point, error) {
	result := []*point.Point{}

	opt := &point.PointOption{
		Category: category.URL(),
		Time:     time.Now(),
	}

	for _, line := range data {
		line, err := decodeBase64Content(line, encode)
		if err != nil {
			return nil, err
		}
		if len(line) > DataByteSizeLimit {
			return nil, fmt.Errorf("data size exceeds 32MB limit")
		}
		switch category { //nolint:exhaustive
		case clipt.Logging:
			pt, err := point.NewPoint(name, nil, map[string]interface{}{
				pipeline.FieldMessage: line,
			}, opt)
			if err != nil {
				return nil, err
			}
			result = append(result, pt)
		case clipt.Metric:
			pts, err := lp.ParsePoints([]byte(line), &lp.Option{EnablePointInKey: true})
			if err != nil {
				return nil, err
			}
			newPts := point.WrapPoint(pts)
			result = append(result, newPts...)
		default:
			pts, err := lp.ParsePoints([]byte(line), nil)
			if err != nil {
				return nil, err
			}
			newPts := point.WrapPoint(pts)
			result = append(result, newPts...)
		}
	}
	return result, nil
}

func decodeBase64Content(content string, encode string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", err
	}

	if encode != "" {
		encode = strings.ToLower(encode)
	}
	switch encode {
	case "gbk", "gb18030":
		data, err = GbToUtf8(data, encode)
		if err != nil {
			return "", err
		}
	case "utf8", "utf-8":
		fallthrough
	default:
	}

	return string(data), nil
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

type plTestSignal struct {
	tn      time.Time
	timeout time.Duration
}

func (s *plTestSignal) ExitSignal() bool {
	return time.Since(s.tn) >= s.timeout
}

func (s *plTestSignal) Reset() {
	s.tn = time.Now()
}

func newPlTestSingal() *plTestSignal {
	return &plTestSignal{
		tn:      time.Now(),
		timeout: time.Millisecond * 500,
	}
}

type pldebugFeed struct {
	d []*point.Point
	sync.Mutex
}

func (f *pldebugFeed) uploadfn(name string, data any) error {
	f.Lock()
	defer f.Unlock()

	if v, ok := data.([]*point.Point); ok {
		for _, v := range v {
			if v != nil {
				f.d = append(f.d, v)
			}
		}
	}
	return nil
}
