package cmds

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

type Content struct {
	Content []Workspace `json:"content"`
}

func TestWorkspaceQuery(t *testing.T) {
	expect := Content{}
	expectBody := `{"content":[{"token":{"ws_uuid":"abc","bill_state":"normal",
	"ver_type":"free","token":"123","db_uuid":"abcd",
	"status":0,"creator":"","expire_at":-1,"create_at":0,"update_at":0,"delete_at":0},
	"data_usage":{"data_metric":97109,"data_logging":13009,"data_tracing":12427,
	"data_rum":0,"is_over_usage":false}}]}`
	if err := json.Unmarshal([]byte(expectBody), &expect); err != nil {
		errorf("json.Unmarshal:%s\n", err)
	}
	cases := []struct {
		body    Content
		expect  Content
		bodyStr string
		flag    bool
	}{
		{
			body:   Content{},
			expect: expect,
			bodyStr: `{"content":[{"token":{"ws_uuid":"abc","bill_state":"normal",
			"ver_type":"free","token":"123","db_uuid":"abcd"},
			"data_usage":{"data_metric":97109,"data_tracing":12427,"data_rum":0}}]}`,
		},
		{
			body:    Content{},
			expect:  expect,
			bodyStr: expectBody,
			flag:    true,
		},
	}

	for _, tc := range cases {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(tc.bodyStr))
		}))
		result, err := doWorkspace(ts.URL)
		if err != nil {
			t.Error(err)
		}
		tu.Ok(t, err)
		if err = json.Unmarshal(result, &tc.body); err != nil {
			errorf("json.Unmarshal:%s\n", err)
		}
		if tc.flag {
			tu.Equals(t, tc.expect, tc.body)
		} else {
			tu.Assert(t, tc.expect.Content[0].DataUsage.DataLogging != tc.body.Content[0].DataUsage.DataLogging,
				"epxect `%d', got `%d'", tc.expect.Content[0].DataUsage.DataLogging, tc.body.Content[0].DataUsage.DataLogging)
		}
		ts.Close()
	}
}
