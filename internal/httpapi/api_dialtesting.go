// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	dt "github.com/GuanceCloud/cliutils/dialtesting"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/gin-gonic/gin"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
)

type dialtestingDebugRequest struct {
	Task     interface{} `json:"task"`
	TaskType string      `json:"task_type"`
}

type dialtestingDebugResponse struct {
	Cost         string                 `json:"cost"`
	ErrorMessage string                 `json:"error_msg"`
	Status       string                 `json:"status"`
	Traceroute   string                 `json:"traceroute"`
	Fields       map[string]interface{} `json:"fields"`
}

func apiDebugDialtestingHandler(w http.ResponseWriter, req *http.Request, whatever ...interface{}) (interface{}, error) {
	var (
		tid        = req.Header.Get(uhttp.XTraceID)
		start      = time.Now()
		t          dt.Task
		traceroute string
		status     = "success"
	)

	reqDebug, err := getAPIDebugDialtestingRequest(req)
	if err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, uhttp.Error(ErrInvalidRequest, err.Error())
	}

	taskType := strings.ToUpper(reqDebug.TaskType)
	switch taskType {
	case dt.ClassHTTP:
		t = &dt.HTTPTask{Option: map[string]string{"userAgent": fmt.Sprintf("DataKit/%s dialtesting", datakit.Version)}}
	case dt.ClassTCP:
		t = &dt.TCPTask{}
	case dt.ClassWebsocket:
		t = &dt.WebsocketTask{}
	case dt.ClassICMP:
		t = &dt.ICMPTask{}
	default:
		l.Errorf("unknown task type: %s", taskType)
		return nil, uhttp.Error(ErrInvalidRequest, fmt.Sprintf("unknown task type:%s", taskType))
	}

	bys, err := json.Marshal(reqDebug.Task)
	if err != nil {
		l.Errorf(`json.Marshal: %s`, err.Error())
		return nil, err
	}

	if err := json.Unmarshal(bys, &t); err != nil {
		l.Errorf(`json.Unmarshal: %s`, err.Error())
		return nil, err
	}
	if strings.ToLower(t.Status()) == dt.StatusStop {
		return nil, uhttp.Error(ErrInvalidRequest, "the task status is stop")
	}

	// -- dialtesting debug procedure start --
	if err := defDialtestingMock.debugInit(t); err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, uhttp.Error(ErrInvalidRequest, err.Error())
	}
	if err := defDialtestingMock.debugRun(t); err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, uhttp.Error(ErrInvalidRequest, err.Error())
	}

	_, fields := defDialtestingMock.getResults(t)

	failReason, ok := fields["fail_reason"].(string)
	if ok {
		status = "fail"
	}
	if taskType == dt.ClassTCP || taskType == dt.ClassICMP {
		traceroute, _ = fields["traceroute"].(string)
	}

	return &dialtestingDebugResponse{
		Cost:         time.Since(start).String(),
		ErrorMessage: failReason,
		Status:       status,
		Traceroute:   traceroute,
		Fields:       fields,
	}, nil
}

func getAPIDebugDialtestingRequest(req *http.Request) (*dialtestingDebugRequest, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, uhttp.Error(ErrInvalidRequest, err.Error())
	}

	var reqDebug dialtestingDebugRequest
	if err := json.Unmarshal(body, &reqDebug); err != nil {
		return nil, uhttp.Error(ErrInvalidRequest, err.Error())
	}

	return &reqDebug, nil
}

func apiUploadSourcemap(c *gin.Context) {
	context := getContext(c)

	// check token valid
	token := c.Query("token")
	if len(token) == 0 {
		context.fail(dcaError{Code: 401, ErrorCode: "auth.failed", ErrorMsg: "auth failed"})
		return
	}
	localTokens := apiServer.dw.GetTokens()
	if len(token) == 0 || len(localTokens) == 0 || (token != localTokens[0]) {
		context.fail(dcaError{Code: 401, ErrorCode: "auth.failed", ErrorMsg: "auth failed"})
		return
	}

	uploadSourcemap(c)
}

func uploadSourcemap(c *gin.Context) {
	context := getContext(c)

	var param rumQueryParam

	if c.ShouldBindQuery(&param) != nil {
		context.fail(dcaError{ErrorCode: "query.parse.error", ErrorMsg: "query string parse error"})
		return
	}

	if param.ApplicationID == "" {
		context.fail(dcaError{ErrorCode: "query.param.required", ErrorMsg: "app_id required"})
		return
	}

	if param.Platform == "" {
		param.Platform = SourceMapDirWeb
	}

	if param.Platform != SourceMapDirWeb && param.Platform != SourceMapDirMini &&
		param.Platform != SourceMapDirAndroid && param.Platform != SourceMapDirIOS {
		l.Errorf("platform [%s] not supported", param.Platform)
		context.fail(dcaError{
			ErrorCode: "param.invalid",
			ErrorMsg:  fmt.Sprintf("platform [%s] not supported, please use web, miniapp, android or ios", param.Platform),
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		l.Errorf("get file failed: %s", err.Error())
		context.fail(dcaError{ErrorCode: "upload.file.required", ErrorMsg: "make sure the file was uploaded correctly"})
		return
	}

	fileName := GetSourcemapZipFileName(param.ApplicationID, param.Env, param.Version)
	rumDir := filepath.Join(GetRumSourcemapDir(), param.Platform)
	if !path.IsDir(rumDir) {
		if err := os.MkdirAll(rumDir, datakit.ConfPerm); err != nil {
			context.fail(dcaError{
				ErrorCode: "dir.create.failed",
				ErrorMsg:  "rum dir created failed",
			})
			return
		}
	}
	dst := filepath.Clean(filepath.Join(rumDir, fileName))

	// check filename
	if !strings.HasPrefix(dst, rumDir) {
		context.fail(dcaError{
			ErrorCode: "param.invalid",
			ErrorMsg:  "invalid param, should not contain illegal char, such as  '../, /'",
		})
		return
	}

	if err := c.SaveUploadedFile(file, dst); err != nil {
		l.Errorf("save upload file error: %s", err.Error())
		context.fail(dcaError{ErrorCode: "upload.file.error", ErrorMsg: "upload failed"})
		return
	}
	runSourcemapCallback(dst)
	context.success(fmt.Sprintf("uploaded to [%s]!", dst))
}
