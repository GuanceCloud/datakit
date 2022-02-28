package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	dt "gitlab.jiagouyun.com/cloudcare-tools/kodo/dialtesting"
)

type debugDialtestingMock struct {
	t *testing.T
}

var (
	initErr     error
	runErr      error
	debugFields map[string]interface{}
)

func (*debugDialtestingMock) debugInit(task *dt.HTTPTask) error {
	return initErr
}

func (*debugDialtestingMock) debugRun(task *dt.HTTPTask) error {
	return runErr
}

func (*debugDialtestingMock) getResults(task *dt.HTTPTask) (tags map[string]string, fields map[string]interface{}) {
	return map[string]string{}, debugFields
}

func init() {
	defDialtestingMock = &debugDialtestingMock{}
}

func TestApiDebugDialtestingHandler(t *testing.T) {
	httpCases := []struct {
		name        string
		t           *dialtestingDebugRequest
		body        []byte
		initErr     error
		runErr      error
		expectErr   error
		expectRes   map[string]interface{}
		debugFields map[string]interface{}
	}{
		{
			name: "test dial task para wrong",
			body: []byte(`{"task":"ddd","dd":"dd"}`),
			t: &dialtestingDebugRequest{
				Task: &dt.HTTPTask{
					ExternalID: cliutils.XID("dtst_"),
					Method:     "d",
					URL:        "https://www.baidu.com",
					Name:       "_test_with_proxy",
					Frequency:  "1s",
					SuccessWhen: []*dt.HTTPSuccess{
						{
							StatusCode: []*dt.SuccessOption{
								{Is: "200"},
							},
						},
					},
				},
			},
			expectErr: uhttp.Error(ErrInvalidRequest, "invalid request"),
		},
		{
			name: "test dial invalid request1",
			t: &dialtestingDebugRequest{
				Task: &dt.HTTPTask{
					ExternalID: cliutils.XID("dtst_"),
					Method:     "d",
					URL:        "https://www.baidu.com",
					Name:       "_test_with_proxy",
					Frequency:  "1s",
					SuccessWhen: []*dt.HTTPSuccess{
						{
							StatusCode: []*dt.SuccessOption{
								{Is: "200"},
							},
						},
					},
				},
			},
			initErr:   uhttp.Error(ErrInvalidRequest, "ddd"),
			expectErr: uhttp.Error(ErrInvalidRequest, "invalid request"),
		},
		{
			name: "test dial invalid request2",
			t: &dialtestingDebugRequest{
				Task: &dt.HTTPTask{
					ExternalID: cliutils.XID("dtst_"),
					Method:     "d",
					URL:        "https://www.baidu.com",
					Name:       "_test_with_proxy",
					Frequency:  "1s",
					SuccessWhen: []*dt.HTTPSuccess{
						{
							StatusCode: []*dt.SuccessOption{
								{Is: "200"},
							},
						},
					},
				},
			},
			runErr:    uhttp.Error(ErrInvalidRequest, "ddd"),
			expectErr: uhttp.Error(ErrInvalidRequest, "invalid request"),
		},
		{
			name: "test dial status stop",
			t: &dialtestingDebugRequest{
				Task: &dt.HTTPTask{
					ExternalID: cliutils.XID("dtst_"),
					Method:     "GET",
					URL:        "https://www.baidu.com",
					Name:       "_test_with_proxy",
					Frequency:  "1s",
					CurStatus:  "stop",
					SuccessWhen: []*dt.HTTPSuccess{
						{
							StatusCode: []*dt.SuccessOption{
								{Is: "200"},
							},
						},
					},
				},
			},
			expectErr: uhttp.Error(ErrInvalidRequest, "the task status is stop"),
		},
		{
			name: "test dial success1",
			t: &dialtestingDebugRequest{
				Task: &dt.HTTPTask{
					ExternalID: cliutils.XID("dtst_"),
					Method:     "GET",
					URL:        "https://www.baidu.com",
					Name:       "_test_with_proxy",
					Frequency:  "1s",
					SuccessWhen: []*dt.HTTPSuccess{
						{
							StatusCode: []*dt.SuccessOption{
								{Is: "200"},
							},
						},
					},
				},
			},
			initErr:   nil,
			expectRes: map[string]interface{}{"Status": "success"},
		},
		{
			name: "test dial success2",
			t: &dialtestingDebugRequest{
				Task: &dt.HTTPTask{
					ExternalID: cliutils.XID("dtst_"),
					Method:     "GET",
					URL:        "https://www.baidu.com",
					Name:       "_test_with_proxy",
					Frequency:  "1s",
					SuccessWhen: []*dt.HTTPSuccess{
						{
							StatusCode: []*dt.SuccessOption{
								{Is: "220"},
							},
						},
					},
				},
			},
			debugFields: map[string]interface{}{
				"fail_reason": "",
			},
			expectRes: map[string]interface{}{"Status": "fail"},
		},
	}

	for _, tc := range httpCases {
		t.Run(tc.name, func(t *testing.T) {
			var w http.ResponseWriter
			initErr = tc.initErr
			runErr = tc.runErr
			debugFields = tc.debugFields
			var bys []byte
			if tc.name == "test dial task para wrong" {
				var tmp map[string]interface{}
				json.Unmarshal(tc.body, &tmp)
				bys, _ = json.Marshal(tmp)
			} else {
				bys, _ = json.Marshal(tc.t)
			}
			req, err := http.NewRequest("POST", "uri", bytes.NewReader(bys))
			res, err := apiDebugDialtestingHandler(w, req)

			if err != nil {
				tu.Equals(t, tc.expectErr, err)
			} else {
				tu.Equals(t, tc.expectRes["Status"], res.(*dialtestingDebugResponse).Status)
			}
		})
	}
}
