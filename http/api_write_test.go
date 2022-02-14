package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/influxdata/influxdb1-client/models"
	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	plw "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/worker"
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

type apiWriteMock struct {
	t *testing.T
}

func (x *apiWriteMock) sendToPipLine(t *plw.Task) error {
	x.t.Helper()
	x.t.Log("under mock impl: sendToPipLine")

	for _, td := range t.Data {
		r := plw.NewResult()
		if err := td.Handler(r); err != nil {
			x.t.Error(err)
		} else {
			x.t.Logf("result: %v", r)
		}
	}

	return nil
}

func (x *apiWriteMock) sendToIO(string, string, []*io.Point, *io.Option) error {
	x.t.Helper()
	x.t.Log("under mock impl: sendToIO")
	return nil // do nothing
}

func (x *apiWriteMock) geoInfo(ip string) map[string]string {
	x.t.Helper()
	x.t.Log("under mock impl: geoInfo")
	return nil // do nothing
}

func TestAPIWrite(t *testing.T) {
	router := gin.New()
	router.Use(uhttp.RequestLoggerMiddleware)
	router.POST("/v1/write/:category", wrap(apiWrite, &apiWriteMock{t: t}))

	ts := httptest.NewServer(router)
	defer ts.Close()

	cases := []struct {
		name, method, url string
		body              []byte
		expectBody        interface{}
		expectStatusCode  int
		contentType       string

		fail bool
	}{
		//--------------------------------------------
		// loging cases
		//--------------------------------------------
		{
			name:             `write-logging-empty-data`,
			method:           "POST",
			url:              "/v1/write/logging",
			body:             []byte(``),
			expectStatusCode: 400,
			expectBody:       ErrEmptyBody,
		},

		{
			name:             `[ok]write-logging(json)`,
			method:           "POST",
			url:              "/v1/write/logging",
			body:             []byte(`[{"measurement":"abc", "tags": {"t1": "xxx"}, "fields":{"f1": 1.0}}]`),
			contentType:      "application/json",
			expectStatusCode: 200,
		},
		{
			name:             `[ok]write-logging(line-proto)`,
			method:           "POST",
			url:              "/v1/write/logging",
			body:             []byte(`xxx-source,t1=1 f1=1i`),
			expectStatusCode: 200,
		},

		{
			name:             `write-logging-json-with-invalid-content-type`,
			method:           "POST",
			url:              "/v1/write/logging",
			body:             []byte(`[{"measurement":"abc", "tags": {"t1": "xxx"}, "fields":{"f1": 1.0}}]`),
			contentType:      "application/xml", // invalid content-type
			expectStatusCode: 400,
			expectBody:       ErrInvalidLinePoint,
		},

		{
			name:             `write-logging-with-source`,
			method:           "POST",
			url:              "/v1/write/logging?source=abc",
			body:             []byte(`xxx-source,t1=1 f1=1i`),
			expectStatusCode: 200,
		},

		//--------------------------------------------
		// RUM cases
		//--------------------------------------------
		{
			name:             `write-rum-unknown-rum-measurement`,
			method:           "POST",
			url:              "/v1/write/rum",
			body:             []byte(`unknown,t1=1 f1=1i`),
			expectStatusCode: 400,
			expectBody:       ErrUnknownRUMMeasurement,
		},

		{
			name:             `write-rum-invalid-json`,
			method:           "POST",
			url:              "/v1/write/rum",
			contentType:      "application/json",
			body:             []byte(`view,t1=1 f1=1i`), // invalid json
			expectStatusCode: 400,
			expectBody:       ErrInvalidJSONPoint,
		},

		{
			name:             `[ok]write-rum-json`,
			method:           "POST",
			url:              "/v1/write/rum?disable_pipeline=1",
			contentType:      "application/json",
			body:             []byte(`[{"measurement":"view","tags":{"t1": "1"}, "fields":{"f1":"1i"}}]`),
			expectStatusCode: 200,
		},

		//--------------------------------------------
		// object cases
		//--------------------------------------------
		{
			name:             `[ok]write-object-json`,
			method:           "POST",
			url:              "/v1/write/object",
			contentType:      "application/json",
			body:             []byte(`[{"measurement":"view","tags":{"t1": "1", "name": "some-obj-name"}, "fields":{"f1":"1i", "message": "dump object message"}}]`),
			expectStatusCode: 200,
		},

		{
			name:             `[ok]write-object-json-missing-name-tag`,
			method:           "POST",
			url:              "/v1/write/object",
			contentType:      "application/json",
			body:             []byte(`[{"measurement":"object-class","tags":{"t1": "1"}, "fields":{"f1":"1i", "message": "dump object message"}}]`),
			expectStatusCode: 400,
			expectBody:       ErrInvalidObjectPoint,
		},
	}

	errEq := func(e1, e2 *uhttp.HttpError) bool {
		return e1.ErrCode == e2.ErrCode
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var resp *http.Response
			var err error
			switch tc.method {
			case "GET":
				resp, err = http.Get(fmt.Sprintf("%s%s", ts.URL, tc.url)) //nolint:bodyclose
				if err != nil {
					t.Logf("http.Get: %s", err)
				}

			case "POST":
				resp, err = http.Post(fmt.Sprintf("%s%s", ts.URL, tc.url), tc.contentType, bytes.NewBuffer(tc.body)) //nolint:bodyclose
				if err != nil {
					t.Error(err)
					return
				}
			default: //
				t.Error("TODO")
				return
			}

			if resp == nil {
				t.Logf("no response")
				return
			}

			defer resp.Body.Close() // nolint:errcheck

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("ioutil.ReadAll: %s", err)
				return
			}

			t.Logf("body: %s", string(body)[:len(body)%256]) // remove too-long-body display

			defer resp.Body.Close() //nolint:errcheck

			if tc.expectBody != nil {
				switch x := tc.expectBody.(type) {
				case *uhttp.HttpError:
					var e uhttp.HttpError
					if err := json.Unmarshal(body, &e); err != nil {
						t.Error(err)
					} else {
						tu.Assert(t, errEq(x, &e), "\n%+#v\nexpect to equal\n%+#v", x, &e)
					}

				default:
					t.Logf("get expect type: %s", reflect.TypeOf(tc.expectBody).String())

					j, err := json.Marshal(tc.expectBody)
					if err != nil {
						t.Errorf("json.Marshal: %s", err)
						return
					}

					tu.Equals(t, string(j), string(body))
				}
			}

			tu.Equals(t, tc.expectStatusCode, resp.StatusCode)
		})
	}
}
