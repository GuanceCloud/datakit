package rum

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	influxm "github.com/influxdata/influxdb1-client/models"
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
		{
			api:    "/v1/write/rum",
			body:   []byte(`rum_app_startup,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			method: `POST`,
			gz:     true,
		},

		{
			api:    "/v1/write/rum",
			body:   []byte(`js_error,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			method: `POST`,
			gz:     true,
		},

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

	for i := len(rumAPICases) - 1; i >= 0; i-- {
		tc := rumAPICases[i]
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

			l.Debugf("x: %v, body: %s", x, string(respbody))
		}
		testutil.Equals(t, string(tc.expectErrCode), string(x.ErrCode))
		t.Logf("[%d] ok", i)
	}
}

func TestRUMAPIHandle(t *testing.T) {

	ln := `freeze,app_id=appid_43a369ea403311eb94f2a6cef984dc00,app_identifier=com.wangjiao.prof.wang.test,app_name=王教授(Dev),city=-,country=-,device=APPLE,device_uuid=D8756690-2555-422B-B185-D936F8B8CE9B,env=prod,freeze_type=Freeze,is_signin=F,isp=unknown,model=iPhone\ 12,network_type=wifi,origin_id=D8756690-2555-422B-B185-D936F8B8CE9B,os=iOS,os_version=14.3,province=-,screen_size=2532*1170,source=freeze,terminal=app,userid=0443EA2E-924F-4C9B-BA1A-F6F4DF0F34F9,version=1.3.5 freeze_stack="Backtrace of Thread 1031:
0 libsystem_kernel.dylib          0x1b7dcc2d0 mach_msg_trap + 8
1 libsystem_kernel.dylib          0x1b7dcb660 mach_msg + 76
2 0x0000000000000000              0xd8338401b7de9ad4 0x0 + 15578940678919330516
3 0x0000000000000000              0xb9784181044e83c8 0x0 + 13364503916600787912
4 AppDev                          0x1044e87d8 _mh_execute_header + 2574296
5 AppDev                          0x1044f41ac _mh_execute_header + 2621868
6 AppDev                          0x1044f211c _mh_execute_header + 2613532
7 QuartzCore                      0x18d3499e8 <redacted> + 664
8 0x0000000000000000              0xe348f0818d422eac 0x0 + 16377604484144180908
9 0x0000000000000000              0xca14130189fd0dd0 0x0 + 14561284392526613968
10 0x0000000000000000              0x5c1a710189ff5fe8 0x0 + 6636741252307967976
11 0x0000000000000000              0x2616c98189ff5378 0x0 + 2744602581132071800
12 0x0000000000000000              0x390a5c0189fef08c 0x0 + 4110198771608907916
13 0x0000000000000000              0x676a8b0189fee21c 0x0 + 7451921372164317724
14 0x0000000000000000              0x1a3b2f81a1af2784 0x0 + 1890156702421952388
15 0x0000000000000000              0xd91303818ca2cfe0 0x0 + 15641849785733009376
16 0x0000000000000000              0x473288818ca32854 0x0 + 5130313015520077908
17 0x0000000000000000              0xe3440c01043b955c 0x0 + 16376227343531480412
18 libdyld.dylib                   0x189cae6b0 <redacted> + 4

",freeze_duration="-1" 1610964279725974016`

	geo.Init()
	r := &Rum{}

	pts, err := influxm.ParsePointsWithPrecision([]byte(ln), time.Now().UTC(), "ns")
	if err != nil {
		t.Fatal(err)
	}

	for _, pt := range pts {
		influxp, esp, err := r.doHandleBody(pt, "222.65.49.189", nil)
		if err != nil {
			t.Error(err)
		}

		if influxp != nil {
			t.Logf("influx point: %s", influxp.String())
			_, err = influxm.ParsePointsWithPrecision([]byte(influxp.String()), time.Now().UTC(), "ns")
			if err != nil {
				t.Fatal(err)
			}
		}
		if esp != nil {
			t.Logf("ES point: %s", esp.String())
			_, err = influxm.ParsePointsWithPrecision([]byte(esp.String()), time.Now().UTC(), "ns")
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}
