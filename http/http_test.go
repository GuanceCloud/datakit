package http

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-sourcemap/sourcemap"
	"github.com/influxdata/influxdb1-client/models"
	"github.com/stretchr/testify/assert"
	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
)

func BenchmarkHandleWriteBody(b *testing.B) {
	body := []byte(`abc,t1=b,t2=d f1=123,f2=3.4,f3="strval" 1624550216
abc,t1=b,t2=d f1=123,f2=3.4,f3="strval" 1624550216`)

	for n := 0; n < b.N; n++ {
		if _, err := handleWriteBody(body, false, &lp.Option{Precision: "s"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHandleJSONWriteBody(b *testing.B) {
	body := []byte(`[
			{
				"measurement":"abc",
				"tags": {"t1":"b", "t2":"d"},
				"fields": {"f1": 123, "f2": 3.4, "f3": "strval"},
				"time":1624550216
			},

			{
				"measurement":"def",
				"tags": {"t1":"b", "t2":"d"},
				"fields": {"f1": 123, "f2": 3.4, "f3": "strval"},
				"time":1624550216
			}
			]`)

	for n := 0; n < b.N; n++ {
		if _, err := handleWriteBody(body, true, &lp.Option{Precision: "s"}); err != nil {
			b.Fatal(err)
		}
	}
}

func TestHandleBody(t *testing.T) {
	cases := []struct {
		name string
		body []byte
		prec string
		fail bool
		js   bool
		npts int
		tags map[string]string
		opt  *lp.Option
	}{
		{
			name: `[json]tag exceed limit`,
			opt: &lp.Option{
				MaxTags: 1,
			},
			body: []byte(`[
			{
				"measurement":"abc",
				"fields": {"f1": 123, "f2": 3.4, "f3": "strval"},
				"tags": {"t1": "abc", "t2": "def"}
			}
			]`),
			fail: true,
			js:   true,
		},

		{
			name: `[json] invalid field key with .`,
			body: []byte(`[
			{
				"measurement":"abc",
				"fields": {"f1": 123, "f.2": 3.4, "f3": "strval"}
			}
			]`),
			fail: true,
			js:   true,
		},

		{
			name: `invalid field`,
			body: []byte(`[
			{
				"measurement":"abc",
				"fields": {"f1": 123, "f2": 3.4, "f3": "strval", "fx": [1,2,3]}
			}
			]`),
			fail: true,
			js:   true,
		},

		{
			name: `json body`,
			body: []byte(`[
			{
				"measurement":"abc",
				"fields": {"f1": 123, "f2": 3.4, "f3": "strval"}
			},
			{
				"measurement":"def",
				"fields": {"f1": 123, "f2": 3.4, "f3": "strval"},
				"time": 1624550216000000000
			}
			]`),
			js:   true,
			npts: 2,
		},

		{
			name: `json body with timestamp`,
			body: []byte(`[
			{
				"measurement":"abc",
				"tags": {"t1":"b", "t2":"d"},
				"fields": {"f1": 123, "f2": 3.4, "f3": "strval"},
				"time":1624550216
			},
			{
				"measurement":"def",
				"tags": {"t1":"b", "t2":"d"},
				"fields": {"f1": 123, "f2": 3.4, "f3": "strval"},
				"time":1624550216
			}
			]`),
			opt:  &lp.Option{Precision: `s`},
			npts: 2,
			js:   true,
		},

		{
			name: `raw point body with/wthout timestamp`,
			opt:  &lp.Option{Precision: `s`},
			body: []byte(`error,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
			view,t1=tag2,t2=tag2 f1=1.0,f2=2i,f3="abc" 1621239130
			resource,t1=tag3,t2=tag2 f1=1.0,f2=2i,f3="abc" 1621239130
			long_task,t1=tag4,t2=tag2 f1=1.0,f2=2i,f3="abc"
			action,t1=tag5,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			npts: 5,
		},

		{
			name: "invalid line protocol",
			body: []byte(`test t1=abc f1=1i,f2=2,f3="str"`),
			npts: 1,
			fail: true,
		},

		{
			name: "multi-line protocol",
			body: []byte(`test,t1=abc f1=1i,f2=2,f3="str"
test,t1=abc f1=1i,f2=2,f3="str"
test,t1=abc f1=1i,f2=2,f3="str"`),
			npts: 3,
		},
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pts, err := handleWriteBody(tc.body, tc.js, tc.opt)

			if tc.fail {
				tu.NotOk(t, err, "case[%d] expect fail, but ok", i)
				t.Logf("[%d] handle body failed: %s", i, err)
				return
			}

			if err != nil && !tc.fail {
				t.Errorf("[FAIL][%d] handle body failed: %s", i, err)
				return
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
		})
	}
}

func TestRUMHandleBody(t *testing.T) {
	cases := []struct {
		name           string
		body           []byte
		prec           string
		fail           bool
		js             bool
		npts           int
		appidWhiteList []string
	}{
		{
			name: `invalid json`,
			body: []byte(`[{
"measurement": "error",
"tags": {"t1": "tv1"},
"fields": 1.0, "f2": 2}
}]`),
			npts: 1,
			fail: true,
			js:   true,
		},

		{
			name: `valid app_id`,
			body: []byte(`[{
"measurement": "error",
"tags": {"app_id": "appid01"},
"fields": {"f1": 1.0, "f2": 2}
}]`),
			npts:           1,
			js:             true,
			appidWhiteList: []string{"appid01"},
		},

		{
			name: `invalid app_id`,
			body: []byte(`[{
"measurement": "error",
"tags": {"app_id": "appid01"},
"fields": {"f1": 1.0, "f2": 2}
}]`),
			npts:           1,
			js:             true,
			appidWhiteList: []string{"appid02"},
			fail:           true,
		},

		{
			name: `invalid json, no tags`,
			body: []byte(`[{
"measurement": "error",
"fields": {"f1": 1.0, "f2": 2}
}]`),
			npts:           1,
			js:             true,
			appidWhiteList: []string{"appid02"},
			fail:           true,
		},

		{
			name: `Precision ms`,
			prec: "ms",
			body: []byte(`error,app_id=appid01,t2=tag2 f1=1.0,f2=2i,f3="abc"
			action,app_id=appid01,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			npts:           2,
			appidWhiteList: []string{"appid01"},
		},

		{
			name: "app_id not in white-list",
			prec: "ms",
			body: []byte(`error,app_id=appid01,t2=tag2 f1=1.0,f2=2i,f3="abc"
			action,app_id=appid01,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			npts:           2,
			appidWhiteList: []string{"appid02"},
			fail:           true,
		},

		{
			name: `Precision ns`,
			prec: "n",
			body: []byte(`error,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
			view,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
			resource,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
			long_task,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
			action,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			npts: 5,
		},

		{
			// 行协议指标带换行
			name: "line break in point",
			prec: "ms",
			body: []byte(`error,sdk_name=Web\ SDK,sdk_version=2.0.1,app_id=appid_16b35953792f4fcda0ca678d81dd6f1a,env=production,version=1.0.0,userid=60f0eae1-01b8-431e-85c9-a0b7bcb391e1,session_id=8c96307f-5ef0-4533-be8f-c84e622578cc,is_signin=F,os=Mac\ OS,os_version=10.11.6,os_version_major=10,browser=Chrome,browser_version=90.0.4430.212,browser_version_major=90,screen_size=1920*1080,network_type=4g,view_id=addb07a3-5ab9-4e30-8b4f-6713fc54fb4e,view_url=http://172.16.5.9:5003/,view_host=172.16.5.9:5003,view_path=/,view_path_group=/,view_url_query={},error_source=source,error_type=ReferenceError error_starttime=1621244127493,error_message="displayDate is not defined",error_stack="ReferenceError
  at onload @ http://172.16.5.9:5003/:25:30" 1621244127493`),
			npts: 1,
		},
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pts, err := doHandleRUMBody(tc.body, tc.prec, tc.js, nil, tc.appidWhiteList)

			if tc.fail {
				tu.NotOk(t, err, "case[%d] expect fail, but ok", i)
				t.Logf("[%d] handle body failed: %s", i, err)
				return
			}

			if err != nil && !tc.fail {
				t.Errorf("[FAIL][%d] handle body failed: %s", i, err)
				return
			}

			tu.Equals(t, tc.npts, len(pts))

			for _, pt := range pts {
				lp := pt.String()
				t.Logf("\t%s", lp)
				_, err := models.ParsePointsWithPrecision([]byte(lp), time.Now(), "n")
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestParsePoint(t *testing.T) {
	cases := []struct {
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

func TestRestartAPI(t *testing.T) {
	tokens := []string{
		"http://1.2.3.4?token=tkn_abc123",
		"http://4.3.2.1?token=tkn_abc456",
	}

	dw = &dataway.DataWayCfg{URLs: tokens}
	if err := dw.Apply(); err != nil {
		t.Error(err)
	}

	cases := []struct {
		token string
		fail  bool
	}{
		{
			token: "tkn_abc123",
			fail:  false,
		},
		{
			token: "tkn_abc456",
			fail:  true,
		},
		{
			token: "",
			fail:  true,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := checkToken(r); err != nil {
			w.WriteHeader(ErrInvalidToken.HttpCode)
			if err := json.NewEncoder(w).Encode(err); err != nil {
				t.Error(err)
			}
		} else {
			w.WriteHeader(200)
		}
	}))

	defer ts.Close() //nolint:errcheck

	time.Sleep(time.Second)

	for _, tc := range cases {
		resp, err := http.Post(fmt.Sprintf("%s?token=%s", ts.URL, tc.token), "", nil)
		if err != nil {
			t.Errorf("error: %s", err)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
			continue
		}
		resp.Body.Close() //nolint:errcheck

		if !tc.fail {
			tu.Equals(t, 200, resp.StatusCode)
		} else {
			tu.Equals(t, ErrInvalidToken.HttpCode, resp.StatusCode)
		}

		t.Logf("resp: %s", string(body))
	}
}

func TestApiGetDatakitLastError(t *testing.T) {
	const uri string = "/v1/lasterror"

	cases := []struct {
		body []byte
		fail bool
	}{
		{
			[]byte(`{"input":"fakeCPU","err_content":"cpu has broken down"}`),
			false,
		},
		{
			[]byte(`{"input":"fakeCPU","err_content":""}`),
			true,
		},
		{
			[]byte(`{"input":"","err_content":"cpu has broken down"}`),
			true,
		},
		{
			[]byte(`{"input":"","err_content":""}`),
			true,
		},
		{
			[]byte(`{"":"fakeCPU","err_content":"cpu has broken down"}`),
			true,
		},
		{
			[]byte(`{"input":"fakeCPU","":"cpu has broken down"}`),
			true,
		},
		{
			[]byte(`{"":"fakeCPU","":"cpu has broken down"}`),
			true,
		},
		{
			[]byte(``),
			true,
		},
	}

	for _, fakeError := range cases {
		fakeEM := errMessage{}
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", uri, bytes.NewReader(fakeError.body))
		if err != nil {
			t.Errorf("create newrequest failed:%s", err)
		}
		em, err := doAPIGetDatakitLastError(req, rr)
		if err != nil {
			if fakeError.fail {
				t.Logf("expect error: %s", err)
				continue
			}
			t.Errorf("api test failed:%s", err)
		}
		err = json.Unmarshal(fakeError.body, &fakeEM)
		if err != nil {
			t.Errorf("json.Unmarshal: %s", err)
		}
		tu.Equals(t, fakeEM.ErrContent, em.ErrContent)
		tu.Equals(t, fakeEM.Input, em.Input)
	}
}

// go test -v -run ^TestReloadServer$ gitlab.jiagouyun.com/cloudcare-tools/datakit/http
/*
func TestReloadServer(t *testing.T) {
	t.Log("start HTTPStart")

	apiConfig = &APIConfig{
		RUMOriginIPHeader: "X-Forwarded-For",
		Listen:            "localhost:54321",
		Disable404Page:    false,
	}

	go HTTPStart()

	t.Log("start sleep")

	time.Sleep(5 * time.Second)

	t.Log("start reload")

	ReloadTheNormalServer()

	t.Log("start loop")

	time.Sleep(60 * time.Second)
} */

func TestHandleSourcemap(t *testing.T) {
	bodyStr := `resource,sdk_name=df_web_rum_sdk,sdk_version=2.0.15,app_id=appid_e6208285ab6947dbaef25d1f1e4749bd,env=production,version=1.0.0,userid=5499c5f6-9b35-43dc-b752-d291a4677b07,session_id=e0e08c94-3096-419e-8eae-1273d045153c,session_type=user,is_signin=F,os=Mac\ OS,os_version=10.14.6,os_version_major=10,browser=Chrome,browser_version=95.0.4638.69,browser_version_major=95,screen_size=2560*1440,network_type=4g,view_id=6f272fe3-5ab1-430c-98b4-45d3d091126c,view_referrer=http://localhost:5500/dist/,view_url=http://localhost:5500/dist/,view_host=localhost:5500,view_path=/dist/,view_path_group=/dist/,view_url_query={},resource_url=https://static.dataflux.cn/browser-sdk/v2/dataflux-rum.js,resource_url_host=static.dataflux.cn,resource_url_path=/browser-sdk/v2/dataflux-rum.js,resource_url_path_group=/browser-sdk/?/dataflux-rum.js,resource_url_query={},resource_type=js,resource_status=200,resource_status_group=2xx,resource_method=GET duration=0 1636524705407
	long_task,sdk_name=df_web_rum_sdk,sdk_version=2.0.15,app_id=appid_e6208285ab6947dbaef25d1f1e4749bd,env=production,version=1.0.0,userid=5499c5f6-9b35-43dc-b752-d291a4677b07,session_id=e0e08c94-3096-419e-8eae-1273d045153c,session_type=user,is_signin=F,os=Mac\ OS,os_version=10.14.6,os_version_major=10,browser=Chrome,browser_version=95.0.4638.69,browser_version_major=95,screen_size=2560*1440,network_type=4g,view_id=6f272fe3-5ab1-430c-98b4-45d3d091126c,view_referrer=http://localhost:5500/dist/,view_url=http://localhost:5500/dist/,view_host=localhost:5500,view_path=/dist/,view_path_group=/dist/,view_url_query={} duration=158000000 1636524705407
	resource,sdk_name=df_web_rum_sdk,sdk_version=2.0.15,app_id=appid_e6208285ab6947dbaef25d1f1e4749bd,env=production,version=1.0.0,userid=5499c5f6-9b35-43dc-b752-d291a4677b07,session_id=e0e08c94-3096-419e-8eae-1273d045153c,session_type=user,is_signin=F,os=Mac\ OS,os_version=10.14.6,os_version_major=10,browser=Chrome,browser_version=95.0.4638.69,browser_version_major=95,screen_size=2560*1440,network_type=4g,view_id=6f272fe3-5ab1-430c-98b4-45d3d091126c,view_referrer=http://localhost:5500/dist/,view_url=http://localhost:5500/dist/,view_host=localhost:5500,view_path=/dist/,view_path_group=/dist/,view_url_query={},resource_url=http://localhost:5500/dist/,resource_url_host=localhost,resource_url_path=/dist/,resource_url_path_group=/dist/,resource_url_query={},resource_type=document,resource_method=GET duration=25300000,resource_ttfb=4900000,resource_trans=1300000,resource_first_byte=16100000 1636524705288
	resource,sdk_name=df_web_rum_sdk,sdk_version=2.0.15,app_id=appid_e6208285ab6947dbaef25d1f1e4749bd,env=production,version=1.0.0,userid=5499c5f6-9b35-43dc-b752-d291a4677b07,session_id=e0e08c94-3096-419e-8eae-1273d045153c,session_type=user,is_signin=F,os=Mac\ OS,os_version=10.14.6,os_version_major=10,browser=Chrome,browser_version=95.0.4638.69,browser_version_major=95,screen_size=2560*1440,network_type=4g,view_id=6f272fe3-5ab1-430c-98b4-45d3d091126c,view_referrer=http://localhost:5500/dist/,view_url=http://localhost:5500/dist/,view_host=localhost:5500,view_path=/dist/,view_path_group=/dist/,view_url_query={},resource_url=http://localhost:5500/dist/bundle.js,resource_url_host=localhost,resource_url_path=/dist/bundle.js,resource_url_path_group=/dist/bundle.js,resource_url_query={},resource_type=js,resource_method=GET duration=15100000,resource_ttfb=2500000,resource_trans=1100000,resource_first_byte=14000000 1636524705407
	long_task,sdk_name=df_web_rum_sdk,sdk_version=2.0.15,app_id=appid_e6208285ab6947dbaef25d1f1e4749bd,env=production,version=1.0.0,userid=5499c5f6-9b35-43dc-b752-d291a4677b07,session_id=e0e08c94-3096-419e-8eae-1273d045153c,session_type=user,is_signin=F,os=Mac\ OS,os_version=10.14.6,os_version_major=10,browser=Chrome,browser_version=95.0.4638.69,browser_version_major=95,screen_size=2560*1440,network_type=4g,view_id=6f272fe3-5ab1-430c-98b4-45d3d091126c,view_referrer=http://localhost:5500/dist/,view_url=http://localhost:5500/dist/,view_host=localhost:5500,view_path=/dist/,view_path_group=/dist/,view_url_query={} duration=118000000 1636524705588
	long_task,sdk_name=df_web_rum_sdk,sdk_version=2.0.15,app_id=appid_e6208285ab6947dbaef25d1f1e4749bd,env=production,version=1.0.0,userid=5499c5f6-9b35-43dc-b752-d291a4677b07,session_id=e0e08c94-3096-419e-8eae-1273d045153c,session_type=user,is_signin=F,os=Mac\ OS,os_version=10.14.6,os_version_major=10,browser=Chrome,browser_version=95.0.4638.69,browser_version_major=95,screen_size=2560*1440,network_type=4g,view_id=6f272fe3-5ab1-430c-98b4-45d3d091126c,view_referrer=http://localhost:5500/dist/,view_url=http://localhost:5500/dist/,view_host=localhost:5500,view_path=/dist/,view_path_group=/dist/,view_url_query={} duration=703000000 1636524705819
	error,sdk_name=df_web_rum_sdk,sdk_version=2.0.15,app_id=appid_e6208285ab6947dbaef25d1f1e4749bd,env=production,version=1.0.0,userid=5499c5f6-9b35-43dc-b752-d291a4677b07,session_id=e0e08c94-3096-419e-8eae-1273d045153c,session_type=user,is_signin=F,os=Mac\ OS,os_version=10.14.6,os_version_major=10,browser=Chrome,browser_version=95.0.4638.69,browser_version_major=95,screen_size=2560*1440,network_type=4g,view_id=6f272fe3-5ab1-430c-98b4-45d3d091126c,view_referrer=http://localhost:5500/dist/,view_url=http://localhost:5500/dist/,view_host=localhost:5500,view_path=/dist/,view_path_group=/dist/,view_url_query={},error_source=source,error_type=TypeError error_message="l.aa is not a function",error_stack="TypeError
		at <anonymous> @ http://localhost:5500/dist/bundle.js:1:821" 1636524715601
	error,sdk_name=df_web_rum_sdk,sdk_version=2.0.15,app_id=appid_e6208285ab6947dbaef25d1f1e4749bd,env=production,version=1.0.0,userid=5499c5f6-9b35-43dc-b752-d291a4677b07,session_id=e0e08c94-3096-419e-8eae-1273d045153c,session_type=user,is_signin=F,os=Mac\ OS,os_version=10.14.6,os_version_major=10,browser=Chrome,browser_version=95.0.4638.69,browser_version_major=95,screen_size=2560*1440,network_type=4g,view_id=6f272fe3-5ab1-430c-98b4-45d3d091126c,view_referrer=http://localhost:5500/dist/,view_url=http://localhost:5500/dist/,view_host=localhost:5500,view_path=/dist/,view_path_group=/dist/,view_url_query={},error_source=source,error_type=TypeError error_message="l.aa is not a function",error_stack="TypeError
		at <anonymous> @ http://localhost:5500/dist/bundle.js:1:821" 1636524725599
	view,sdk_name=df_web_rum_sdk,sdk_version=2.0.15,app_id=appid_e6208285ab6947dbaef25d1f1e4749bd,env=production,version=1.0.0,userid=5499c5f6-9b35-43dc-b752-d291a4677b07,session_id=e0e08c94-3096-419e-8eae-1273d045153c,session_type=user,is_signin=F,os=Mac\ OS,os_version=10.14.6,os_version_major=10,browser=Chrome,browser_version=95.0.4638.69,browser_version_major=95,screen_size=2560*1440,network_type=4g,view_id=6f272fe3-5ab1-430c-98b4-45d3d091126c,view_referrer=http://localhost:5500/dist/,view_url=http://localhost:5500/dist/,view_host=localhost:5500,view_path=/dist/,view_path_group=/dist/,view_url_query={},view_loading_type=initial_load,view_apdex_level=0,is_active=true view_error_count=2,view_resource_count=3,view_long_task_count=3,view_action_count=0,cumulative_layout_shift=0,loading_time=1252500000,dom_interactive=310000000,dom_content_loaded=331900000,dom_complete=1249000000,first_paint_time=17400000,resource_load_time=917200000,time_to_interactive=302100000,dom=939000000,dom_ready=324000000,time_spent=23318700000 1636524705288`
	body := []byte(bodyStr)
	mapFileContent := `{"version":3,"file":"bundle.js","mappings":"qBAeAA,EAAOC,QAfP,WACEC,QAAQC,IAAI,gBACZD,QAAQC,IAAI,gBACZD,QAAQC,IAAI,gBACZD,QAAQC,IAAI,gBACZD,QAAQC,IAAI,gBACZD,QAAQC,IAAI,sBACZD,QAAQC,IAAI,gBACZD,QAAQC,IAAI,gBACZD,QAAQC,IAAI,gBACZD,QAAQC,IAAI,gBACZD,QAAQC,IAAI,gBACZD,QAAQC,IAAI,mBCXVC,EAA2B,GAG/B,SAASC,EAAoBC,GAE5B,IAAIC,EAAeH,EAAyBE,GAC5C,QAAqBE,IAAjBD,EACH,OAAOA,EAAaN,QAGrB,IAAID,EAASI,EAAyBE,GAAY,CAGjDL,QAAS,IAOV,OAHAQ,EAAoBH,GAAUN,EAAQA,EAAOC,QAASI,GAG/CL,EAAOC,QCpBfI,EAAoBK,EAAKV,IACxB,IAAIW,EAASX,GAAUA,EAAOY,WAC7B,IAAOZ,EAAiB,QACxB,IAAM,EAEP,OADAK,EAAoBQ,EAAEF,EAAQ,CAAEG,EAAGH,IAC5BA,GCLRN,EAAoBQ,EAAI,CAACZ,EAASc,KACjC,IAAI,IAAIC,KAAOD,EACXV,EAAoBY,EAAEF,EAAYC,KAASX,EAAoBY,EAAEhB,EAASe,IAC5EE,OAAOC,eAAelB,EAASe,EAAK,CAAEI,YAAY,EAAMC,IAAKN,EAAWC,MCJ3EX,EAAoBY,EAAI,CAACK,EAAKC,IAAUL,OAAOM,UAAUC,eAAeC,KAAKJ,EAAKC,G,yCCEhFrB,QAAQC,IAAI,KCEd,MAEAwB,aAAY,KAUVC,EAAEC,OARD,KAEH,IAAID,EAAI,I","sources":["webpack:///./src/func1.js","webpack:///webpack/bootstrap","webpack:///webpack/runtime/compat get default export","webpack:///webpack/runtime/define property getters","webpack:///webpack/runtime/hasOwnProperty shorthand","webpack:///./src/func.js","webpack:///./src/index.js"],"sourcesContent":["function add () {\n  console.log(\"xxxxxxxxxxxx\")\n  console.log(\"xxxxxxxxxxxx\")\n  console.log(\"xxxxxxxxxxxx\")\n  console.log(\"xxxxxxxxxxxx\")\n  console.log(\"xxxxxxxxxxxx\")\n  console.log(\"xxxxxx\\n\\n\\nxxxxxx\")\n  console.log(\"xxxxxxxxxxxx\")\n  console.log(\"xxxxxxxxxxxx\")\n  console.log(\"xxxxxxxxxxxx\")\n  console.log(\"xxxxxxxxxxxx\")\n  console.log(\"xxxxxxxxxxxx\")\n  console.log(\"xxxxxxxxxxxx\")\n}\n\nmodule.exports = add","// The module cache\nvar __webpack_module_cache__ = {};\n\n// The require function\nfunction __webpack_require__(moduleId) {\n\t// Check if module is in cache\n\tvar cachedModule = __webpack_module_cache__[moduleId];\n\tif (cachedModule !== undefined) {\n\t\treturn cachedModule.exports;\n\t}\n\t// Create a new module (and put it into the cache)\n\tvar module = __webpack_module_cache__[moduleId] = {\n\t\t// no module.id needed\n\t\t// no module.loaded needed\n\t\texports: {}\n\t};\n\n\t// Execute the module function\n\t__webpack_modules__[moduleId](module, module.exports, __webpack_require__);\n\n\t// Return the exports of the module\n\treturn module.exports;\n}\n\n","// getDefaultExport function for compatibility with non-harmony modules\n__webpack_require__.n = (module) => {\n\tvar getter = module && module.__esModule ?\n\t\t() => (module['default']) :\n\t\t() => (module);\n\t__webpack_require__.d(getter, { a: getter });\n\treturn getter;\n};","// define getter functions for harmony exports\n__webpack_require__.d = (exports, definition) => {\n\tfor(var key in definition) {\n\t\tif(__webpack_require__.o(definition, key) && !__webpack_require__.o(exports, key)) {\n\t\t\tObject.defineProperty(exports, key, { enumerable: true, get: definition[key] });\n\t\t}\n\t}\n};","__webpack_require__.o = (obj, prop) => (Object.prototype.hasOwnProperty.call(obj, prop))","export default function run() {\n  debugger\n  console.log(\"a\")\n}","import run from './func'\nimport run1 from './func1'\n\nrun()\nrun1()\n\nsetInterval(() => {\n  a()\n}, 10000)\n\nlet b = {}\nfunction a(){\n  c()\n}\n\nfunction c(){\n  b.aa()\n}"],"names":["module","exports","console","log","__webpack_module_cache__","__webpack_require__","moduleId","cachedModule","undefined","__webpack_modules__","n","getter","__esModule","d","a","definition","key","o","Object","defineProperty","enumerable","get","obj","prop","prototype","hasOwnProperty","call","setInterval","b","aa"],"sourceRoot":""}`

	tmpDir := os.TempDir()

	datakit.DataDir = path.Join(tmpDir, "data")

	sourcemapFileName := GetSourcemapZipFileName("appid_e6208285ab6947dbaef25d1f1e4749bd", "production", "1.0.0")
	rumDir := GetRumSourcemapDir()

	err := os.MkdirAll(rumDir, os.ModePerm)
	assert.NoError(t, err)

	buf := new(bytes.Buffer)

	w := zip.NewWriter(buf)

	f, err := w.Create("dist/bundle.js.map")

	assert.NoError(t, err)
	_, err = f.Write([]byte(mapFileContent))
	assert.NoError(t, err)
	err = w.Close()
	assert.NoError(t, err)

	zipFilePath := filepath.Join(rumDir, sourcemapFileName)

	err = ioutil.WriteFile(zipFilePath, buf.Bytes(), os.ModePerm)

	assert.NoError(t, err)
	defer os.Remove(zipFilePath) //nolint:errcheck

	loadSourcemapFile()

	pts, err := lp.ParsePoints(body, &lp.Option{
		Time:      time.Now(),
		Precision: "ms",
		Callback: func(p models.Point) (models.Point, error) {
			if string(p.Name()) == "error" {
				_ = handleSourcemap(p)
			}
			return p, nil
		},
	})

	assert.NoError(t, err)
	assert.Greater(t, len(pts), 0)

	for _, p := range pts {
		if p.Name() == "error" {
			fields, err := p.Fields()
			if err != nil {
				continue
			}
			if _, ok := fields["error_stack"]; ok {
				tags := p.Tags()
				if errorStackSource, ok := tags["error_stack_source_base64"]; !ok {
					assert.Fail(t, "error stack transform failed")
				} else {
					decodBytes, err := base64.StdEncoding.DecodeString(errorStackSource)
					assert.NoError(t, err)
					assert.Contains(t, string(decodBytes), "webpack")
				}
			}
		}
	}

	t.Run("updateSourcemapCache", func(t *testing.T) {
		updateSourcemapCache("invalid")
		assert.Equal(t, len(sourcemapCache), 1)

		updateSourcemapCache(zipFilePath)
		fileName := filepath.Base(zipFilePath)
		_, ok := sourcemapCache[fileName]
		assert.True(t, ok)

		deleteSourcemapCache(zipFilePath)
		_, ok = sourcemapCache[fileName]
		assert.False(t, ok)
	})

	t.Run("getSourceMapString", func(t *testing.T) {
		updateSourcemapCache(zipFilePath)
		fileName := filepath.Base(zipFilePath)
		cases := []struct {
			src           string
			dest          string
			desc          string
			sourceMapItem map[string]*sourcemap.Consumer
		}{
			{
				src:  "http://a.com/1.js",
				dest: "http://a.com/1.js",
				desc: "split with colon, length less than 3",
			},
			{
				src:  "http://a.com/1.js:a:1",
				dest: "http://a.com/1.js:a:1",
				desc: "invalid row number",
			},
			{
				src:  "http://a.com/1.js:1:a",
				dest: "http://a.com/1.js:1:a",
				desc: "invalid col number",
			},
			{
				src:           "http://localhost:5500/dist/bundle.js:1:821",
				dest:          "webpack:///./src/index.js:17:4",
				desc:          "it shourld work",
				sourceMapItem: sourcemapCache[fileName],
			},
			{
				src:           "http:///localhost:5500/dist/bundle.js:1:821",
				dest:          "http:///localhost:5500/dist/bundle.js:1:821",
				desc:          "invalid url",
				sourceMapItem: sourcemapCache[fileName],
			},
			{
				src:  "http://localhost:5500/dist/bundle.js:1:821",
				dest: "http://localhost:5500/dist/bundle.js:1:821",
				desc: "it shourld not work without sourcemap file",
			},
		}

		for _, item := range cases {
			assert.Equal(t, getSourceMapString(item.src, item.sourceMapItem), item.dest, item.desc)
		}
	})
}
