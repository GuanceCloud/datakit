package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/influxdb1-client/models"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
)

func BenchmarkHandleWriteBody(b *testing.B) {

	body := []byte(`abc,t1=b,t2=d f1=123,f2=3.4,f3="strval" 1624550216
abc,t1=b,t2=d f1=123,f2=3.4,f3="strval" 1624550216`)

	for n := 0; n < b.N; n++ {
		handleWriteBody(body, "s", nil, false)
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
		handleWriteBody(body, "s", nil, true)
	}
}

func TestHandleBody(t *testing.T) {
	var cases = []struct {
		body []byte
		prec string
		fail bool
		js   bool
		npts int
		tags map[string]string
	}{

		{
			body: []byte(`[
			{
				"measurement":"abc",
				"fields": {"f1": 123, "f2": 3.4, "f3": "strval", "fx": [1,2,3]}
			}
			]`),
			fail: true, // invalid field
			js:   true,
		},

		{
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
			prec: "s",
			npts: 2,
			js:   true,
		},

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
		pts, err := handleWriteBody(tc.body, tc.prec, tc.tags, tc.js)

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
		js   bool
		npts int
	}{

		{
			body: []byte(`[{
"measurement": "error",
"tags": {"t1": "tv1"},
"fields": {"f1": 1.0, "f2": 2}
}]`),
			npts: 1,
			js:   true,
		},

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
		pts, err := handleRUMBody(tc.body, tc.prec, "", tc.js)

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
			json.NewEncoder(w).Encode(err)
		} else {
			w.WriteHeader(200)
		}
	}))

	defer ts.Close()

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
		resp.Body.Close()

		if !tc.fail {
			tu.Equals(t, 200, resp.StatusCode)
		} else {
			tu.Equals(t, ErrInvalidToken.HttpCode, resp.StatusCode)
		}

		t.Logf("resp: %s", string(body))
	}
}
