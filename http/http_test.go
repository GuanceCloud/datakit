package http

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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

var (
	__bind = ":12345"
)

type dkAPICase struct {
	api           string
	body          []byte
	method        string
	gz            bool
	expectErrCode string
}

var dkApiCases = []*dkAPICase{
	{
		api:    "/v1/write/metric?name=test",
		body:   []byte(`test,t1=abc f1=1i,f2=2,f3="str"`),
		method: "POST",
		gz:     true,
	},

	{
		api:           "/v1/write/metric?name=test",
		body:          []byte(`test t1=abc f1=1i,f2=2,f3="str"`),
		method:        "POST",
		gz:            true,
		expectErrCode: "datakit.badRequest",
	},
}

func TestAPI(t *testing.T) {
	io.SetTest()
	datakit.Cfg.MainCfg.GinLog = "./gin.log"

	go func() {
		HttpStart(__bind)
	}()

	time.Sleep(time.Second)
	runCases(t, dkApiCases)
}

func runCases(t *testing.T, cases []*dkAPICase) {

	httpCli := &http.Client{}
	var err error

	for i := len(cases) - 1; i >= 0; i-- {
		tc := cases[i]
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
