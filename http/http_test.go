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

	"github.com/influxdata/influxdb1-client/models"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

var (
	__host  = "http://127.0.0.1"
	__bind  = ":12345"
	__token = "tkn_2dc438b6693711eb8ff97aeee04b54af"
)

func TestParsePoint(t *testing.T) {

	var cases = []struct {
		body []byte
		prec string
		npts int
		fail bool
	}{
		{
			body: []byte(`m1,t1=abc f1=123 123`),
			prec: "h",
			npts: 1,
		},

		{
			body: []byte(`m1,t1=abc f1=123 123`),
			prec: "m",
			npts: 1,
		},

		{
			body: []byte(`m1,t1=abc f1=123 123`),
			prec: "s",
			npts: 1,
		},

		{
			body: []byte(`m1,t1=abc f1=123 123`),
			prec: "ms",
			npts: 1,
		},

		{
			body: []byte(`m1,t1=abc f1=123 123`),
			prec: "u",
			npts: 1,
		},

		{
			body: []byte(`m1,t1=abc f1=123 123`),
			prec: "n",
			npts: 1,
		},
	}

	for _, tc := range cases {
		points, err := models.ParsePointsWithPrecision(tc.body, time.Now(), tc.prec)
		if tc.fail {
			tu.NotOk(t, err, "")
		} else {
			tu.Equals(t, tc.npts, len(points))
			for _, pt := range points {
				t.Log(pt.String())
			}
		}
	}
}

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
		api    string
		body   []byte
		method string
		gz     bool
		fail   bool
	}{
		{
			api:    "/v1/write/metric?precision=mss",
			body:   []byte(`rum_app_startup,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc" 1621315267`),
			method: `POST`,
			fail:   true,
		},

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
			api:    "/v1/write/metric?input=test",
			body:   []byte(`test t1=abc f1=1i,f2=2,f3="str"`),
			method: "POST",
			gz:     true,
			fail:   true,
		},
		{
			api:    "/v1/write/metric?input=test&token=" + __token,
			body:   []byte(`test-01,category=host,host=ubt-server,level=warn,title=a\ demo message="passwd 发生了变化" 1619599490000652659`),
			method: "POST",
			gz:     true,
		},
		{
			api:    "/v1/write/metric?input=test&token=" + __token,
			body:   []byte(``),
			method: "POST",
			gz:     true,
			fail:   true,
		},
		{
			api:    "/v1/write/object?input=test&token=" + __token,
			body:   []byte(``),
			method: "POST",
			gz:     true,
			fail:   true,
		},
		{
			api:    "/v1/write/logging?input=test&token=" + __token,
			body:   []byte(``),
			method: "POST",
			gz:     true,
			fail:   true,
		},
		{
			api:    "/v1/write/keyevent?input=test&token=" + __token,
			body:   []byte(``),
			method: "POST",
			gz:     true,
			fail:   true,
		},

		// rum cases
		{
			api:    "/v1/write/rum?input=test&token=" + __token,
			body:   []byte(``),
			method: "POST",
			gz:     true,
			fail:   true,
		},

		{ // unknown RUM metric
			api:    "/v1/write/rum?input=rum-test",
			body:   []byte(`not_rum_metric,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			method: `POST`,
			gz:     true,
			fail:   true,
		},

		{ // bad line-proto
			api:    "/v1/write/rum?input=rum-test",
			body:   []byte(`not_rum_metric,t1=tag1,t2=tag2 f1=1.0f,f2=2i,f3="abc"`),
			method: `POST`,
			gz:     true,
			fail:   true,
		},

		{
			api:    "/v1/write/rum?input=rum-test",
			body:   []byte(`js_error,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			method: `POST`,
			gz:     true,
		},

		{
			api:    "/v1/write/rum",
			body:   []byte(`rum_app_startup,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			method: `POST`,
			gz:     true,
		},

		{
			api:    "/v1/write/rum?precision=ms",
			body:   []byte(`rum_app_startup,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc" 1621315267`),
			method: `POST`,
			gz:     true,
		},

		{
			api:    "/v1/write/xxx?precision=ms",
			body:   []byte(`rum_app_startup,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc" 1621315267`),
			method: `POST`,
			fail:   true,
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

	for i, tc := range cases {
		if tc.gz {
			tc.body, err = datakit.GZip(tc.body)
			if err != nil {
				t.Fatal(err)
			}
		}

		urlstr := fmt.Sprintf("%s%s%s", __host, __bind, tc.api)
		t.Logf("request URL: %s", urlstr)
		req, err := nhttp.NewRequest(tc.method, urlstr,
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

		if tc.fail {
			tu.Assert(t, resp.StatusCode != http.StatusOK, "[%d] http should be failed, but ok", i)
		} else {
			tu.Assert(t, resp.StatusCode == http.StatusOK,
				"[%d] request failed, http code %d, body: %s", i, resp.Status, string(respbody))
		}

		if len(respbody) > 0 {
			if err := json.Unmarshal(respbody, &x); err != nil {
				t.Error(err.Error())
			}
		}

		t.Logf("case [%d] %s ok", i, cases[i].api)
	}
}
