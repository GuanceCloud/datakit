package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	dt "gitlab.jiagouyun.com/cloudcare-tools/kodo/dialtesting"
)

type debugDialtestingMock struct{}

var (
	errInit     error
	errRun      error
	debugFields map[string]interface{}
)

func (*debugDialtestingMock) debugInit(task dt.Task) error {
	return errInit
}

func (*debugDialtestingMock) debugRun(task dt.Task) error {
	return errRun
}

func (*debugDialtestingMock) getResults(task dt.Task) (tags map[string]string, fields map[string]interface{}) {
	return map[string]string{}, debugFields
}

func init() { // nolint:gochecknoinits
	defDialtestingMock = &debugDialtestingMock{}
}

func TestApiDebugDialtestingHandler(t *testing.T) {
	httpCases := []struct {
		name        string
		t           *dialtestingDebugRequest
		body        []byte
		errInit     error
		errRun      error
		errExpect   error
		expectRes   map[string]interface{}
		debugFields map[string]interface{}
	}{
		{
			name: "test dial task para wrong",
			body: []byte(`{"task_type":9,"dd":"dd"}`),
			t: &dialtestingDebugRequest{
				TaskType: "HTTP",
				Task:     &dt.HTTPTask{},
			},
			errExpect: uhttp.Error(ErrInvalidRequest, "invalid request"),
		},
		{
			name: "test dial task para wrong1",
			body: []byte(`{"task_type":"dd","dd":"dd"}`),
			t: &dialtestingDebugRequest{
				TaskType: "dd",
				Task:     &dt.HTTPTask{},
			},
			errExpect: uhttp.Error(ErrInvalidRequest, "unknown task type:DD"),
		},
		{
			name: "test dial invalid request1",
			t: &dialtestingDebugRequest{
				TaskType: "HTTP",
				Task:     &dt.HTTPTask{},
			},
			errInit:   uhttp.Error(ErrInvalidRequest, "ddd"),
			errExpect: uhttp.Error(ErrInvalidRequest, "invalid request"),
		},
		{
			name: "test dial invalid request2",
			t: &dialtestingDebugRequest{
				TaskType: "HTTP",
				Task:     &dt.HTTPTask{},
			},
			errRun:    uhttp.Error(ErrInvalidRequest, "ddd"),
			errExpect: uhttp.Error(ErrInvalidRequest, "invalid request"),
		},
		{
			name: "test dial status stop",
			t: &dialtestingDebugRequest{
				TaskType: "HTTP",
				Task: &dt.HTTPTask{
					CurStatus: "stop",
				},
			},
			errExpect: uhttp.Error(ErrInvalidRequest, "the task status is stop"),
		},
		{
			name: "test dial success1",
			t: &dialtestingDebugRequest{
				TaskType: "HTTP",
				Task:     &dt.HTTPTask{},
			},
			errInit:   nil,
			expectRes: map[string]interface{}{"Status": "success"},
		},
		{
			name: "test dial success2",
			t: &dialtestingDebugRequest{
				TaskType: "HTTP",
				Task:     &dt.HTTPTask{},
			},
			debugFields: map[string]interface{}{
				"fail_reason": "",
			},
			expectRes: map[string]interface{}{"Status": "fail"},
		},
		{
			name: "test dial success3",
			t: &dialtestingDebugRequest{
				TaskType: "TCP",
				Task:     &dt.TcpTask{},
			},
			errInit:   nil,
			expectRes: map[string]interface{}{"Status": "timeout"},
		},
		{
			name: "test dial success4",
			t: &dialtestingDebugRequest{
				TaskType: "ICMP",
				Task:     &dt.IcmpTask{},
			},
			errInit:   nil,
			expectRes: map[string]interface{}{"Status": "success"},
		},
		{
			name: "test dial success5",
			t: &dialtestingDebugRequest{
				TaskType: "WEBSOCKET",
				Task:     &dt.WebsocketTask{},
			},
			errInit:   nil,
			expectRes: map[string]interface{}{"Status": "success"},
		},
	}

	for _, tc := range httpCases {
		t.Run(tc.name, func(t *testing.T) {
			var w http.ResponseWriter
			errInit = tc.errInit
			errRun = tc.errRun
			debugFields = tc.debugFields
			var bys []byte
			if tc.name == "test dial task para wrong" || tc.name == "test dial task para wrong1" {
				var tmp map[string]interface{}
				json.Unmarshal(tc.body, &tmp)
				bys, _ = json.Marshal(tmp)
			} else {
				bys, _ = json.Marshal(tc.t)
			}
			req, err := http.NewRequest("POST", "uri", bytes.NewReader(bys))
			if err != nil {
				t.Log(err)
			}
			res, err := apiDebugDialtestingHandler(w, req)

			if err != nil {
				tu.Equals(t, tc.errExpect, err)
			} else {
				tu.Equals(t, tc.expectRes["Status"], res.(*dialtestingDebugResponse).Status)
			}
		})
	}
}
