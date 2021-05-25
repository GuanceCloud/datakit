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

func TestHandleBody(t *testing.T) {
	var cases = []struct {
		body []byte
		prec string
		fail bool
		npts int
		tags map[string]string
	}{
		{
			prec: "s",
			body: []byte(`error,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
			view,t1=tag2,t2=tag2 f1=1.0,f2=2i,f3="abc" 1621239130
			resource,t1=tag3,t2=tag2 f1=1.0,f2=2i,f3="abc" 1621239130
			long_task,t1=tag4,t2=tag2 f1=1.0,f2=2i,f3="abc"
			action,t1=tag5,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			npts: 5,
		},

		{
			body: []byte(`test t1=abc f1=1i,f2=2,f3="str"`),
			npts: 1,
			fail: true,
		},

		{
			body: []byte(`test,t1=abc f1=1i,f2=2,f3="str"
test,t1=abc f1=1i,f2=2,f3="str"
test,t1=abc f1=1i,f2=2,f3="str"`),
			npts: 3,
		},
	}

	for i, tc := range cases {
		pts, err := handleWriteBody(tc.body, tc.tags, tc.prec)

		if tc.fail {
			tu.NotOk(t, err, "case[%d] expect fail, but ok", i)
			t.Logf("[%d] handle body failed: %s", i, err)
			continue
		}

		if err != nil && !tc.fail {
			t.Errorf("[FAIL][%d] handle body failed: %s", i, err)
			continue
		}

		tu.Equals(t, tc.npts, len(pts))

		t.Logf("----------- [%d] -----------", i)
		for _, pt := range pts {
			s := pt.String()
			fs, err := pt.Fields()
			if err != nil {
				t.Error(err)
				continue
			}

			x, err := models.NewPoint(pt.Name(), models.NewTags(pt.Tags()), fs, pt.Time())
			if err != nil {
				t.Error(err)
				continue
			}

			t.Logf("\t%s, key: %s, hash: %d", s, x.Key(), x.HashID())
		}
	}
}

func TestRUMHandleBody(t *testing.T) {

	var cases = []struct {
		body []byte
		prec string
		fail bool
		npts int
	}{
		{
			prec: "ms",
			body: []byte(`error,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
			view,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc" 1621239130000
			resource,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc" 1621239130000
			long_task,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
			action,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			npts: 5,
		},

		{
			prec: "n",
			body: []byte(`error,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
			view,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
			resource,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
			long_task,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
			action,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			npts: 5,
		},
		{
			prec: "ms",
			// 行协议指标带换行
			body: []byte(`error,sdk_name=Web\ SDK,sdk_version=2.0.1,app_id=appid_16b35953792f4fcda0ca678d81dd6f1a,env=production,version=1.0.0,userid=60f0eae1-01b8-431e-85c9-a0b7bcb391e1,session_id=8c96307f-5ef0-4533-be8f-c84e622578cc,is_signin=F,os=Mac\ OS,os_version=10.11.6,os_version_major=10,browser=Chrome,browser_version=90.0.4430.212,browser_version_major=90,screen_size=1920*1080,network_type=4g,view_id=addb07a3-5ab9-4e30-8b4f-6713fc54fb4e,view_url=http://172.16.5.9:5003/,view_host=172.16.5.9:5003,view_path=/,view_path_group=/,view_url_query={},error_source=source,error_type=ReferenceError error_starttime=1621244127493,error_message="displayDate is not defined",error_stack="ReferenceError
  at onload @ http://172.16.5.9:5003/:25:30" 1621244127493`),
			npts: 1,
		},
	}

	for i, tc := range cases {
		pts, err := handleRUMBody(tc.body, tc.prec, "")

		if tc.fail {
			tu.NotOk(t, err, "case[%d] expect fail, but ok", i)
			t.Logf("[%d] handle body failed: %s", i, err)
			continue
		}

		if err != nil && !tc.fail {
			t.Errorf("[FAIL][%d] handle body failed: %s", i, err)
			continue
		}

		tu.Equals(t, tc.npts, len(pts))

		t.Logf("----------- [%d] -----------", i)
		for _, pt := range pts {
			lp := pt.String()
			t.Logf("\t%s", lp)
			_, err := models.ParsePointsWithPrecision([]byte(lp), time.Now(), "n")
			if err != nil {
				t.Error(err)
			}
		}
	}
}

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
			body:   []byte(`abc,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc" 1621315267`),
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
			api:    "/v1/write/rum?input=unknown-RUM-metric",
			body:   []byte(`not_rum_metric,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			method: `POST`,
			gz:     true,
			fail:   true,
		},

		{ // bad line-proto
			api:    "/v1/write/rum?input=bad-line-proto",
			body:   []byte(`not_rum_metric,t1=tag1,t2=tag2 f1=1.0f,f2=2i,f3="abc"`),
			method: `POST`,
			gz:     true,
			fail:   true,
		},

		{
			api:    "/v1/write/rum?input=rum-test",
			body:   []byte(`error,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			method: `POST`,
			gz:     true,
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
			api:    "/v1/write/rum?precision=ms",
			body:   []byte(`resource,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc" 1621315267`),
			method: `POST`,
			gz:     true,
		},

		{
			api:    "/v1/write/xxx?precision=ms",
			body:   []byte(`view,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc" 1621315267`),
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

		if tc.fail {
			tu.Assert(t, resp.StatusCode != http.StatusOK, "[%d] http should be failed, but ok, api: %s", i, tc.api)
		} else {
			tu.Assert(t, resp.StatusCode == http.StatusOK,
				"[%d] request failed, http code %d, body: %s, api: %s", i, resp.Status, string(respbody), tc.api)
		}

		if len(respbody) > 0 {

			var x struct {
				ErrCode string `json:"error_code"`
				Msg     string `json:"message"`
			}

			if err := json.Unmarshal(respbody, &x); err != nil {
				t.Error(err.Error())
			}
		}
		t.Logf("case [%d] %s ok", i, cases[i].api)
	}
}
