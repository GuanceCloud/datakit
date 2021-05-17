package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	nhttp "net/http"
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

var (
	__host  = "http://127.0.0.1"
	__bind  = ":12345"
	__token = "tkn_2dc438b6693711eb8ff97aeee04b54af"
)

func TestReload(t *testing.T) {
	Start(&Option{Bind: __bind, GinLog: ".gin.log", PProf: true})
	time.Sleep(time.Second)

	n := 10

	for i := 0; i < n; i++ {
		if err := ReloadDatakit(&reloadOption{}); err != nil {
			t.Error(err)
		}

		go RestartHttpServer()
		time.Sleep(time.Second)
	}

	HttpStop()
	<-stopOkCh // wait HTTP server stop tk
	if reloadCnt != n {
		t.Errorf("reload count unmatch: expect %d, got %d", n, reloadCnt)
	}
	t.Log("HTTP server stop ok")
}

func TestAPI(t *testing.T) {

	var cases = []struct {
		api           string
		body          []byte
		method        string
		gz            bool
		expectErrCode string
	}{

		{
			api:    "/v1/ping",
			method: "GET",
			gz:     false,
		},

		{
			api:    "/v1/write/metric?input=test",
			body:   []byte(`test,t1=abc f1=1i,f2=2,f3="str"`),
			method: "POST",
			gz:     true,
		},
		{
			api:           "/v1/write/metric?input=test",
			body:          []byte(`test t1=abc f1=1i,f2=2,f3="str"`),
			method:        "POST",
			gz:            true,
			expectErrCode: "datakit.badRequest",
		},
		{
			api:           "/v1/write/metric?input=test&token=" + __token,
			body:          []byte(`test-01,category=host,host=ubt-server,level=warn,title=a\ demo message="passwd 发生了变化" 1619599490000652659`),
			method:        "POST",
			gz:            true,
			expectErrCode: "datakit.badRequest",
		},
		{
			api:           "/v1/write/metric?input=test&token=" + __token,
			body:          []byte(``),
			method:        "POST",
			gz:            true,
			expectErrCode: "datakit.badRequest",
		},
		{
			api:           "/v1/write/object?input=test&token=" + __token,
			body:          []byte(``),
			method:        "POST",
			gz:            true,
			expectErrCode: "datakit.badRequest",
		},
		{
			api:           "/v1/write/logging?input=test&token=" + __token,
			body:          []byte(``),
			method:        "POST",
			gz:            true,
			expectErrCode: "datakit.badRequest",
		},
		{
			api:           "/v1/write/keyevent?input=test&token=" + __token,
			body:          []byte(``),
			method:        "POST",
			gz:            true,
			expectErrCode: "datakit.badRequest",
		},

		// rum cases
		{
			api:           "/v1/write/rum?input=test&token=" + __token,
			body:          []byte(``),
			method:        "POST",
			gz:            true,
			expectErrCode: "datakit.badRequest",
		},

		{ // unknown RUM metric
			api:           "/v1/write/rum?input=rum-test",
			body:          []byte(`not_rum_metric,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			method:        `POST`,
			gz:            true,
			expectErrCode: "datakit.badRequest",
		},

		{ // bad line-proto
			api:           "/v1/write/rum?input=rum-test",
			body:          []byte(`not_rum_metric,t1=tag1,t2=tag2 f1=1.0f,f2=2i,f3="abc"`),
			method:        `POST`,
			gz:            true,
			expectErrCode: "datakit.badRequest",
		},

		{
			api:           "/v1/write/rum?input=rum-test",
			body:          []byte(`js_error,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			method:        `POST`,
			expectErrCode: "datakit.badRequest",
			gz:            true,
		},

		{
			api: "/v1/write/rum?input=rum-test",
			body: []byte(`error,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
			view,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
			resource,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
			long_task,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
			action,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			method: `POST`,
			gz:     true,
		},

		{
			api:           "/v1/write/rum",
			body:          []byte(`rum_app_startup,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			method:        `POST`,
			gz:            true,
			expectErrCode: "datakit.badRequest",
		},
	}

	httpBind = __bind
	io.SetTest()
	ginLog = "./gin.log"

	go func() {
		HttpStart()
	}()

	time.Sleep(time.Second)

	httpCli := &nhttp.Client{}
	var err error

	for i := len(cases) - 1; i >= 0; i-- {
		tc := cases[i]
		if tc.gz {
			tc.body, err = datakit.GZip(tc.body)
			if err != nil {
				t.Fatal(err)
			}
		}

		req, err := nhttp.NewRequest(tc.method,
			fmt.Sprintf("%s%s%s", __host, __bind, tc.api),
			bytes.NewBuffer([]byte(tc.body)))
		if err != nil {
			t.Fatal(err)
		}

		if tc.gz {
			req.Header.Set("Content-Encoding", "gzip")
		}

		resp, err := httpCli.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		respbody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		var x struct {
			ErrCode string `json:"error_code"`
			Msg     string `json:"message"`
		}

		if len(respbody) > 0 {
			if err := json.Unmarshal(respbody, &x); err != nil {
				t.Error(err.Error())
			}

			if tc.expectErrCode != "" {
				tu.Equals(t, string(tc.expectErrCode), string(x.ErrCode))
			} else {
				if resp.StatusCode != http.StatusOK {
					t.Errorf("[FAIL][%d] api %s request faild with status code: %s, body: %s\n", i, cases[i].api, resp.Status, string(respbody))
					continue
				}
				t.Logf("[%d] x: %v, body: %s", i, x, string(respbody))
			}
		}

		t.Logf("case [%d] ok: %s", i, cases[i].api)
	}
}
