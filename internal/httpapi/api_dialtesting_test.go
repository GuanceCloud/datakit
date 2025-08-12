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
	"github.com/stretchr/testify/assert"
)

type debugDialtestingMock struct{}

var (
	errInit     error
	errRun      error
	debugFields map[string]interface{}
)

func (*debugDialtestingMock) debugInit(task dt.ITask, vars map[string]dt.Variable) error {
	return errInit
}

func (*debugDialtestingMock) debugRun(task dt.ITask) error {
	return errRun
}

func (*debugDialtestingMock) getResults(task dt.ITask) (tags map[string]string, fields map[string]interface{}) {
	return map[string]string{}, debugFields
}

func (*debugDialtestingMock) getVars(task dt.ITask) dt.Vars {
	return task.GetPostScriptVars()
}

func init() { // nolint:gochecknoinits
	defDialtestingMock = &debugDialtestingMock{}
}

func TestApiDebugDialtestingHandler(t *testing.T) {
	httpCases := []struct {
		name        string
		dr          *dialtestingDebugRequest
		body        []byte
		errInit     error
		errRun      error
		errContains string
		expectRes   map[string]interface{}
		debugFields map[string]interface{}
		hook        func()
	}{
		{
			name:        "test-dial-task-para-wrong",
			body:        []byte(`{"task_type":9,"dd":"dd"}`),
			dr:          &dialtestingDebugRequest{},
			errContains: "json: cannot unmarshal",
		},

		{
			name:        "test-dial-task-para-wrong1",
			body:        []byte(`{"task_type":"dd","dd":"dd"}`),
			dr:          &dialtestingDebugRequest{},
			errContains: "unknown task type:DD",
		},

		{
			name: "test-dial-invalid-request1",
			dr: &dialtestingDebugRequest{
				TaskType: "HTTP",
				Task:     &dt.HTTPTask{},
			},
			errInit:     uhttp.Error(ErrInvalidRequest, "ddd"),
			errContains: "invalid request",
		},

		{
			name: "test-dial-status-stop",
			dr: &dialtestingDebugRequest{
				TaskType: "HTTP",
				Task: &dt.HTTPTask{
					Task: &dt.Task{
						CurStatus: "stop",
					},
				},
			},
			errContains: "the task status is stop",
		},
		{
			name: "test-dial-success1",
			dr: &dialtestingDebugRequest{
				TaskType: "HTTP",
				Task:     &dt.HTTPTask{},
			},
			errInit:   nil,
			expectRes: map[string]interface{}{"Status": "success"},
		},
		{
			name: "test-dial-success2",
			dr: &dialtestingDebugRequest{
				TaskType: "HTTP",
				Task:     &dt.HTTPTask{},
			},
			debugFields: map[string]interface{}{
				"fail_reason": "",
			},
			expectRes: map[string]interface{}{"Status": "fail"},
		},
		{
			name: "test-dial-success3",
			dr: &dialtestingDebugRequest{
				TaskType: "TCP",
				Task:     &dt.TCPTask{},
			},
			errInit:   nil,
			expectRes: map[string]interface{}{"Status": "success"},
		},
		{
			name: "test-dial-success4",
			dr: &dialtestingDebugRequest{
				TaskType: "ICMP",
				Task: &dt.ICMPTask{
					Host: "127.0.0.1",
				},
			},
			errInit:   nil,
			expectRes: map[string]interface{}{"Status": "success"},
		},
		{
			name: "test-dial-success5",
			dr: &dialtestingDebugRequest{
				TaskType: "WEBSOCKET",
				Task:     &dt.WebsocketTask{},
			},
			errInit:   nil,
			expectRes: map[string]interface{}{"Status": "success"},
		},
		{
			name: "test-internal-host-private",
			dr: &dialtestingDebugRequest{
				TaskType: "ICMP",
				Task: &dt.ICMPTask{
					Host: "192.168.0.1",
				},
			},
			hook: func() {
				os.Setenv("ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK", "true")
			},
			errInit:     nil,
			errContains: "is not allowed to be tested",
		},
		{
			name: "test-internal-host-illegal-host",
			dr: &dialtestingDebugRequest{
				TaskType: "http",
				Task: &dt.HTTPTask{
					URL: "http://①0.43.239.255:5000", // invalid URL characters
				},
			},
			hook: func() {
				os.Setenv("ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK", "true")
			},
			errInit:     nil,
			errContains: "lookup ip failed: lookup ①0.43.239.255: no such host",
		},
		{
			name: "test-internal-host-cidrs",
			dr: &dialtestingDebugRequest{
				TaskType: "ICMP",
				Task: &dt.ICMPTask{
					Host: "36.155.132.76",
				},
			},
			hook: func() {
				os.Setenv("ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK", "true")
				os.Setenv("ENV_INPUT_DIALTESTING_DISABLED_INTERNAL_NETWORK_CIDR_LIST", `["36.155.132.76/24"]`)
			},
			errInit:     nil,
			errContains: "is not allowed to be tested",
		},
		{
			name: "test-internal-host-ok",
			dr: &dialtestingDebugRequest{
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
			name: "test-internal-host-loopback",
			dr: &dialtestingDebugRequest{
				TaskType: "icmp",
				Task: &dt.ICMPTask{
					Host: "localhost",
				},
			},
			hook: func() {
				os.Setenv("ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK", "true")
			},
			errInit:     nil,
			errContains: "is not allowed to be tested",
		},

		{
			name: "test-internal-host-unspecified",
			dr: &dialtestingDebugRequest{
				TaskType: "icmp",
				Task: &dt.ICMPTask{
					Host: "0.0.0.0",
				},
			},
			hook: func() {
				os.Setenv("ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK", "true")
			},
			errInit:     nil,
			errContains: "is not allowed to be tested",
		},
	}

	for _, tc := range httpCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				os.Unsetenv("ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK")
				os.Unsetenv("ENV_INPUT_DIALTESTING_DISABLED_INTERNAL_NETWORK_CIDR_LIST")

				DialtestingDisableInternalNetworkTask = false
				DialtestingEnableDebugAPI = false
				DialtestingDisabledInternalNetworkCidrList = []string{}
			}()

			if tc.hook != nil {
				tc.hook()
			}

			parseDialtestingEnvs()

			var (
				w   http.ResponseWriter
				bys []byte
			)

			errInit = tc.errInit
			errRun = tc.errRun
			debugFields = tc.debugFields

			if tc.body != nil {
				bys = tc.body
			} else {
				bys, _ = json.Marshal(tc.dr)
			}

			req, err := http.NewRequest("POST", "not-set", bytes.NewReader(bys))
			if err != nil {
				t.Log(err)
			}

			res, err := apiDebugDialtestingHandler(w, req)

			if err != nil {
				assert.ErrorContains(t, err, tc.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectRes["Status"], res.(*dialtestingDebugResponse).Status)
			}
		})
	}
}
