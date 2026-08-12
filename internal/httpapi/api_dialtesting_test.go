// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	dt "github.com/GuanceCloud/cliutils/dialtesting"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	previousSetup := dialtestingNetPathDebugTaskSetup
	dialtestingNetPathDebugTaskSetup = func(context.Context, *dt.NetPathTask) {}
	t.Cleanup(func() {
		dialtestingNetPathDebugTaskSetup = previousSetup
	})

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
			name: "test-dial-success-ssl",
			dr: &dialtestingDebugRequest{
				TaskType: "SSL",
				Task: &dt.SSLTask{
					Task: &dt.Task{
						Name:      "ssl-debug",
						Frequency: "1m",
					},
					Host: "example.com",
					Port: "443",
					SuccessWhen: []*dt.SSLSuccess{
						{
							ResponseTime: "1s",
						},
					},
				},
			},
			debugFields: map[string]interface{}{
				"tls_version": "TLS1.3",
			},
			errInit:   nil,
			expectRes: map[string]interface{}{"Status": "success"},
		},
		{
			name: "test-dial-success-netpath",
			dr: &dialtestingDebugRequest{
				TaskType: "netpath",
				Task: &dt.NetPathTask{
					Task:     &dt.Task{ExternalID: "netpath-debug", Name: "netpath-debug"},
					Protocol: "tcp",
					Host:     "example.com",
					Port:     "443",
				},
			},
			debugFields: map[string]interface{}{
				"fail_reason": "",
				"traceroute":  `{"runs":[],"hop_count":{"avg":0,"min":0,"max":0}}`,
			},
			errInit: nil,
			expectRes: map[string]interface{}{
				"Status":     "success",
				"Traceroute": `{"runs":[],"hop_count":{"avg":0,"min":0,"max":0}}`,
			},
		},
		{
			name: "test-dial-netpath-init-error-redacted",
			dr: &dialtestingDebugRequest{
				TaskType: "netpath",
				Task: &dt.NetPathTask{
					Task:     &dt.Task{ExternalID: "netpath-debug", Name: "netpath-debug"},
					Protocol: "icmp",
					Host:     "example.com",
				},
			},
			errInit:     assert.AnError,
			errContains: "invalid NETPATH task configuration",
		},
		{
			name: "test-dial-netpath-run-error-redacted",
			dr: &dialtestingDebugRequest{
				TaskType: "netpath",
				Task: &dt.NetPathTask{
					Task:     &dt.Task{ExternalID: "netpath-debug", Name: "netpath-debug"},
					Protocol: "icmp",
					Host:     "example.com",
				},
			},
			errRun:      assert.AnError,
			errContains: "NETPATH task run failed",
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
			errContains: "The internal network address does not support online testing. However, it can be saved and then used normally.",
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
			errContains: "The internal network address does not support online testing. However, it can be saved and then used normally.",
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
			errContains: "The internal network address does not support online testing. However, it can be saved and then used normally.",
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
			errContains: "The internal network address does not support online testing. However, it can be saved and then used normally.",
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
				if expected, ok := tc.expectRes["Traceroute"]; ok {
					assert.Equal(t, expected, res.(*dialtestingDebugResponse).Traceroute)
				}
			}
		})
	}
}

func TestAPIDebugDialtestingNetPathRejectsMissingExecutor(t *testing.T) {
	previousSetup := dialtestingNetPathDebugTaskSetup
	dialtestingNetPathDebugTaskSetup = nil
	t.Cleanup(func() {
		dialtestingNetPathDebugTaskSetup = previousSetup
	})

	body := []byte(`{
		"task_type":"netpath",
		"task":{
			"external_id":"netpath-debug",
			"status":"OK",
			"protocol":"tcp",
			"host":"example.com",
			"port":"443"
		}
	}`)
	req, err := http.NewRequest(http.MethodPost, "not-set", bytes.NewReader(body))
	require.NoError(t, err)

	_, err = apiDebugDialtestingHandler(nil, req)
	assert.ErrorContains(t, err, "NETPATH debug executor is not registered")
}

func TestAPIDebugDialtestingNetPathParseErrorRedactsPayload(t *testing.T) {
	body := []byte(`{
		"task_type":"netpath",
		"task":{
			"external_id":"netpath-debug",
			"access_key":"ak-secret",
			"post_url":"https://openway.example.com?token=tkn-secret",
			"advance_options":"invalid"
		}
	}`)
	req, err := http.NewRequest(http.MethodPost, "not-set", bytes.NewReader(body))
	require.NoError(t, err)

	_, err = apiDebugDialtestingHandler(nil, req)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "invalid NETPATH task payload")
	assert.NotContains(t, err.Error(), "ak-secret")
	assert.NotContains(t, err.Error(), "tkn-secret")
}

func TestIsAllowedHost(t *testing.T) {
	t.Run("allow when internal network check disabled", func(t *testing.T) {
		DialtestingDisableInternalNetworkTask = false

		ok, err := isAllowedHost([]string{"host-a"}, func(host string) (bool, error) {
			return false, nil
		})
		if !assert.NoError(t, err) {
			return
		}
		assert.True(t, ok)
	})

	t.Run("deny when any host is internal", func(t *testing.T) {
		DialtestingDisableInternalNetworkTask = true

		ok, err := isAllowedHost([]string{"host-a", "host-b"}, func(host string) (bool, error) {
			return host == "host-b", nil
		})
		if !assert.NoError(t, err) {
			return
		}
		assert.False(t, ok)
	})

	t.Run("return error when any host is invalid", func(t *testing.T) {
		DialtestingDisableInternalNetworkTask = true

		ok, err := isAllowedHost([]string{"host-a", "host-b"}, func(host string) (bool, error) {
			if host == "host-b" {
				return false, assert.AnError
			}
			return false, nil
		})
		if !assert.Error(t, err) {
			return
		}
		assert.False(t, ok)
	})
}

func TestIsInternalHostContextHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := isInternalHostContext(ctx, "blocked.example", nil,
		func(ctx context.Context, network, host string) ([]net.IP, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		})
	assert.ErrorIs(t, err, context.Canceled)
}

func TestDialtestingDebugDNSLookupHasFixedTimeout(t *testing.T) {
	var remaining time.Duration
	allowed, err := isAllowedDialtestingDebugHostWithChecker(
		context.Background(),
		[]string{"fixed-timeout.invalid"},
		func(ctx context.Context, hosts []string) (bool, error) {
			deadline, ok := ctx.Deadline()
			require.True(t, ok)
			remaining = time.Until(deadline)
			return true, nil
		},
	)
	require.NoError(t, err)
	assert.True(t, allowed)
	assert.Greater(t, remaining, 14*time.Second)
	assert.LessOrEqual(t, remaining, 15*time.Second)
}
