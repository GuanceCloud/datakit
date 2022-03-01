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

	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/multiline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/worker"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	categoryPipelineLogging = "logging"

	// https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/608
	maxLenPipelineBefore = 3 * 4096
	maxLenDataBefore     = 3 * 8192
	maxLenDataAfter      = 8192
)

type pipelineDebugRequest struct {
	Pipeline  string
	Source    string
	Service   string
	Category  string
	Data      string
	Multiline string
	Encode    string
	Benchmark bool
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

	//------------------------------------------------------------------
	// -- pipeline debug procedure start --

	// STEP 1: check pipeline
	decodePipeline, err := base64.StdEncoding.DecodeString(reqDebug.Pipeline)
	if err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, uhttp.Error(ErrInvalidPipeline, err.Error())
	}

	ng, err := worker.ParsePlScript(string(decodePipeline))
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

	// STEP 3: pipeline processing
	start := time.Now()
	res := worker.RunAsPlTask(reqDebug.Category, reqDebug.Source, reqDebug.Service, dataLines, ng)

	// STEP 4 (optional): benchmark
	var benchmarkResult testing.BenchmarkResult
	if reqDebug.Benchmark {
		benchmarkResult = testing.Benchmark(func(b *testing.B) {
			b.Helper()
			for n := 0; n < b.N; n++ {
				worker.RunAsPlTask(reqDebug.Category, reqDebug.Source, reqDebug.Service, []string{dataLines[0]}, ng)
			}
		})
	}

	// -- pipeline debug procedure end --
	//------------------------------------------------------------------

	return getReturnResult(start, res, reqDebug, &benchmarkResult), nil
}

func getReturnResult(start time.Time, res []*worker.Result,
	reqDebug *pipelineDebugRequest,
	benchmarkResult *testing.BenchmarkResult) *pipelineDebugResponse {
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

	multi, err := multiline.New(pattern, maxLenDataAfter)
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

	if multi.CacheLines() != 0 {
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
	if reqDebug.Category != categoryPipelineLogging {
		return uhttp.Error(ErrInvalidCategory, "invalid category")
	}

	if len(reqDebug.Pipeline) > maxLenPipelineBefore ||
		len(reqDebug.Source) > maxLenPipelineBefore ||
		len(reqDebug.Service) > maxLenPipelineBefore ||
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
