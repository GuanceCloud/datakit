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
