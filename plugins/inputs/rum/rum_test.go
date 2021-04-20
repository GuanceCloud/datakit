package rum

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/geo"
)

type rumAPICase struct {
	api           string
	body          []byte
	method        string
	gz            bool
	expectErrCode string
	fail          bool
}

var (
	__bind      = ":12345"
	rumAPICases = []*rumAPICase{

		{ // unknown RUM metric
			api:           "/v1/write/rum",
			body:          []byte(`not_rum_metric,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			method:        `POST`,
			gz:            true,
			fail:          true,
			expectErrCode: "datakit.badRequest",
		},

		{ // bad line-proto
			api:           "/v1/write/rum",
			body:          []byte(`not_rum_metric,t1=tag1,t2=tag2 f1=1.0f,f2=2i,f3="abc"`),
			method:        `POST`,
			gz:            true,
			expectErrCode: "datakit.badRequest",
			fail:          true,
		},

		{
			api:    "/v1/write/rum",
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
	}
)

func TestRUMHandle(t *testing.T) {
	r := &Rum{}
	for _, tc := range rumAPICases {
		ifdata, esdata, err := r.handleBody([]byte(tc.body), DEFAULT_PRECISION, "1.2.3.4")
		if err != nil {
			t.Log(err)
			if !tc.fail {
				t.Fatal(err)
			}
		}

		for _, pt := range ifdata {
			t.Logf("ifdata: %s", pt.String())
		}
		for _, pt := range esdata {
			t.Logf("ifdata: %s", pt.String())
		}
	}
}

func TestAPI(t *testing.T) {

	io.SetTest()
	datakit.Cfg.MainCfg.GinLog = "./gin.log"

	ruminput := &Rum{}
	ruminput.RegHttpHandler()

	go func() {
		httpd.HttpStart(__bind)
	}()
	time.Sleep(time.Second)

	httpCli := &http.Client{}
	var err error

	for idx, tc := range rumAPICases {
		if tc.gz {
			tc.body, err = datakit.GZip(tc.body)
			if err != nil {
				t.Fatal(err)
			}
		}

		req, err := http.NewRequest(tc.method,
			fmt.Sprintf("http://0.0.0.0%s%s", __bind, tc.api),
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
				t.Fatal(err)
			}

			l.Debugf("[%d] x: %v, body: %s", idx, x, string(respbody))
		}
		testutil.Equals(t, string(tc.expectErrCode), string(x.ErrCode))
		t.Logf("[%d] ok", idx)
	}
}

func TestRUMAPIHandle(t *testing.T) {

	ln := `freeze,app_id=appid_43a369ea403311eb94f2a6cef984dc00,app_identifier=com.wangjiao.prof.wang.test,app_name=王教授(Dev),device=APPLE,device_uuid=D8756690-2555-422B-B185-D936F8B8CE9B,env=prod,freeze_type=Freeze,is_signin=F,model=iPhone\ 12,network_type=wifi,origin_id=D8756690-2555-422B-B185-D936F8B8CE9B,os=iOS,os_version=14.3,screen_size=2532*1170,source=freeze,terminal=app,userid=0443EA2E-924F-4C9B-BA1A-F6F4DF0F34F9,version=1.3.5 freeze_stack="Backtrace of Thread 1031:0 libsystem_kernel.dylib          0x1b7dcc2d0 mach_msg_trap + 8 ",freeze_duration="-1" 1610964279725974016
rum_web_page_performance,app_id=appid_43a369ea403311eb94f2a6cef984dc00,app_identifier=com.wangjiao.prof.wang.test,app_name=王教授(Dev),device=APPLE,device_uuid=D8756690-2555-422B-B185-D936F8B8CE9B,env=prod,freeze_type=Freeze,is_signin=F,model=iPhone\ 12,network_type=wifi,origin_id=D8756690-2555-422B-B185-D936F8B8CE9B,os=iOS,os_version=14.3,screen_size=2532*1170,source=freeze,terminal=app,userid=0443EA2E-924F-4C9B-BA1A-F6F4DF0F34F9,version=1.3.5 freeze_stack="Backtrace of Thread 1031:0 libsystem_kernel.dylib          0x1b7dcc2d0 mach_msg_trap + 8 ",freeze_duration="-1" 1610964279725974016`

	geo.Init()
	r := &Rum{}

	mpts, rumpts, err := r.handleBody([]byte(ln), "ns", "39.156.69.79")
	if err != nil {
		t.Error(err)
	}

	if len(mpts) != 1 {
		t.Errorf("expect 1 metric point")
	}

	for _, x := range mpts {
		t.Log(x.String())
	}

	if len(rumpts) != 1 {
		t.Errorf("expect 1 RUM point")
	}

	for _, x := range rumpts {

		t.Log(x.String())

		t.Log("RUM tags")
		for k, y := range x.Tags() {
			t.Logf("\t%s: %s", k, y)
		}

		t.Log("RUM fields")
		fields, err := x.Fields()
		if err != nil {
			t.Error()
		}

		for k, y := range fields {
			t.Logf("\t%s: %s", k, y)
		}
	}
}
