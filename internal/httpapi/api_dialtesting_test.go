// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	dt "github.com/GuanceCloud/cliutils/dialtesting"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	tu "github.com/GuanceCloud/cliutils/testutil"
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
		hook        func()
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
				Task:     &dt.TCPTask{},
			},
			errInit:   nil,
			expectRes: map[string]interface{}{"Status": "success"},
		},
		{
			name: "test dial success4",
			t: &dialtestingDebugRequest{
				TaskType: "ICMP",
				Task: &dt.ICMPTask{
					Host: "127.0.0.1",
				},
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
		{
			name: "test internal host - private",
			t: &dialtestingDebugRequest{
				TaskType: "ICMP",
				Task: &dt.ICMPTask{
					Host: "192.168.0.1",
				},
			},
			hook: func() {
				os.Setenv("ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK", "true")
			},
			errInit:   nil,
			errExpect: uhttp.Errorf(ErrInvalidRequest, "dest host [%s] is not allowed to be tested", "192.168.0.1"),
		},
		{
			name: "test internal host cidrs",
			t: &dialtestingDebugRequest{
				TaskType: "ICMP",
				Task: &dt.ICMPTask{
					Host: "36.155.132.76",
				},
			},
			hook: func() {
				os.Setenv("ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK", "true")
				os.Setenv("ENV_INPUT_DIALTESTING_DISABLED_INTERNAL_NETWORK_CIDR_LIST", `["36.155.132.76/24"]`)
			},
			errInit:   nil,
			errExpect: uhttp.Errorf(ErrInvalidRequest, "dest host [%s] is not allowed to be tested", "36.155.132.76"),
		},
		{
			name: "test internal host ok",
			t: &dialtestingDebugRequest{
				TaskType: "ICMP",
				Task: &dt.ICMPTask{
					Host: "192.168.0.1",
				},
			},
			hook: func() {
				os.Setenv("ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK", "false")
			},
			errInit:   nil,
			expectRes: map[string]interface{}{"Status": "success"},
		},
		{
			name: "test internal host loopback",
			t: &dialtestingDebugRequest{
				TaskType: "icmp",
				Task: &dt.ICMPTask{
					Host: "localhost",
				},
			},
			hook: func() {
				os.Setenv("ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK", "true")
			},
			errInit:   nil,
			errExpect: uhttp.Errorf(ErrInvalidRequest, "dest host [%s] is not allowed to be tested", "localhost"),
		},
		{
			name: "test internal host unspecified",
			t: &dialtestingDebugRequest{
				TaskType: "icmp",
				Task: &dt.ICMPTask{
					Host: "0.0.0.0",
				},
			},
			hook: func() {
				os.Setenv("ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK", "true")
			},
			errInit:   nil,
			errExpect: uhttp.Errorf(ErrInvalidRequest, "dest host [%s] is not allowed to be tested", "0.0.0.0"),
		},
	}

	for _, tc := range httpCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Unsetenv("ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK")
			os.Unsetenv("ENV_INPUT_DIALTESTING_DISABLED_INTERNAL_NETWORK_CIDR_LIST")
			if tc.hook != nil {
				tc.hook()
			}
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
