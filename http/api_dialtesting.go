package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	dt "gitlab.jiagouyun.com/cloudcare-tools/kodo/dialtesting"
)

type dialtestingDebugRequest struct {
	Task *dt.HTTPTask
}

type dialtestingDebugResponse struct {
	Cost         string `json:"cost"`
	ErrorMessage string `json:"error_msg"`
	Status       string `json:"status"`
}

func apiDebugDialtestingHandler(w http.ResponseWriter, req *http.Request, whatever ...interface{}) (interface{}, error) {
	tid := req.Header.Get(uhttp.XTraceId)
	start := time.Now()
	reqDebug, err := getAPIDebugDialtestingRequest(req)
	if err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, uhttp.Error(ErrInvalidRequest, err.Error())
	}

	if reqDebug.Task.CurStatus == "stop" {
		return nil, uhttp.Error(ErrInvalidRequest, "the task status is stop")
	}

	//------------------------------------------------------------------
	// -- dialtesting debug procedure start --
	if err := defDialtestingMock.debugInit(reqDebug.Task); err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, uhttp.Error(ErrInvalidRequest, err.Error())
	}
	if err := defDialtestingMock.debugRun(reqDebug.Task); err != nil {
		l.Errorf("[%s] %s", tid, err.Error())
		return nil, uhttp.Error(ErrInvalidRequest, err.Error())
	}

	_, fields := defDialtestingMock.getResults(reqDebug.Task)

	failReason, ok := fields["fail_reason"].(string)
	status := "success"
	if ok {
		status = "fail"
	}
	return &dialtestingDebugResponse{
		Cost:         time.Since(start).String(),
		ErrorMessage: failReason,
		Status:       status,
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
