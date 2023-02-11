// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/influxdata/influxdb1-client/models"
	"github.com/stretchr/testify/assert"
	lp "github.com/GuanceCloud/cliutils/lineproto"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	tu "github.com/GuanceCloud/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
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

			opt: func() *lp.Option {
				o := lp.NewDefaultOption()
				o.MaxTags = 1
				return o
			}(),

			body: []byte(`[
			{
				"measurement":"abc",
				"fields": {"f1": 123, "f2": 3.4, "f3": "strval"},
				"tags": {"t1": "abc", "t2": "def"}
			}
			]`),
			js:   true,
			npts: 1,
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

			opt: func() *lp.Option {
				o := lp.NewDefaultOption()
				o.Precision = "s"
				return o
			}(),

			npts: 2,
			js:   true,
		},

		{
			name: `raw point body with/wthout timestamp`,
			opt: func() *lp.Option {
				o := lp.NewDefaultOption()
				o.PrecisionV2 = lp.Second
				return o
			}(),
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

func (x *apiWriteMock) sendToIO(string, string, []*point.Point, *io.Option) error {
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
	const timestamp = 1000000000 // 2001-09-09 01:46:40 +0000 UTC

	router := gin.New()
	router.Use(uhttp.RequestLoggerMiddleware)
	router.POST("/v1/write/:category", rawHTTPWraper(nil, apiWrite, &apiWriteMock{t: t}))

	ts := httptest.NewServer(router)
	defer ts.Close()

	cases := []struct {
		name, method, url string
		body              []byte
		expectBody        interface{}
		expectStatusCode  int
		contentType       string

		globalHostTags, globalEnvTags map[string]string

		fail bool
	}{
		//--------------------------------------------
		// logging cases
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
			name:             `write-logging-with-point-in-key`,
			method:           "POST",
			url:              "/v1/write/logging",
			body:             []byte(`some-source,tag1=1,tag2=2,tag.3=3 f1=1`),
			expectStatusCode: 400,
		},

		{
			name:             `unstrict-write-logging-with-point-in-key`,
			method:           "POST",
			url:              "/v1/write/logging",
			body:             []byte(`some-source,tag1=1,tag2=2,tag.3=3 f1=1`),
			expectStatusCode: 400,
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
			expectBody:       ErrInvalidCategory,
		},

		{
			name:             `write-rum-invalid-json`,
			method:           "POST",
			url:              "/v1/write/rum",
			contentType:      "application/json",
			body:             []byte(`view,t1=1 f1=1i`), // invalid json
			expectStatusCode: 400,
			expectBody:       ErrInvalidCategory,
		},

		{
			name:             `[ok]write-rum-json`,
			method:           "POST",
			url:              "/v1/write/rum?disable_pipeline=1",
			contentType:      "application/json",
			body:             []byte(`[{"measurement":"view","tags":{"t1": "1"}, "fields":{"f1":"1i"}}]`),
			expectStatusCode: 400,
		},

		//--------------------------------------------
		// metric cases
		//--------------------------------------------
		{
			name:             `metric-json-point-key-with-point`,
			method:           "POST",
			url:              "/v1/write/metric",
			contentType:      "application/json",
			body:             []byte(`[{"measurement":"view","tags":{"t1": "1", "name": "some-obj-name"}, "fields":{"f1.1":1, "f2": 3.14}}]`),
			expectStatusCode: 200,
		},
		{
			name:             `metric-json-point-key-with-point-seconds`,
			method:           "POST",
			url:              "/v1/write/metric?echo_line_proto=1&precision=s",
			contentType:      "application/json",
			body:             []byte(`[{"measurement":"view","tags":{"t1": "1", "name": "some-obj-name"}, "fields":{"f1.1":1, "f2": 3.14}, "time":` + fmt.Sprintf("%d", timestamp) + `}]`),
			expectStatusCode: 200,
			expectBody: &uhttp.BodyResp{
				Content: []*sinkcommon.JSONPoint{
					{
						Measurement: "view",
						Tags: map[string]string{
							"t1": "1", "name": "some-obj-name",
						},
						Fields: map[string]interface{}{
							"f1.1": 1, "f2": 3.14,
						},
						Time: time.Unix(timestamp, 0).UTC(),
					},
				},
			},
		},
		{
			name:             `metric-json-point-key-with-point-nanoseconds`,
			method:           "POST",
			url:              "/v1/write/metric?echo_line_proto=1",
			contentType:      "application/json",
			body:             []byte(`[{"measurement":"view","tags":{"t1": "1", "name": "some-obj-name"}, "fields":{"f1.1":1, "f2": 3.14}, "time":` + fmt.Sprintf("%d", timestamp*1000*1000*1000) + `}]`),
			expectStatusCode: 200,
			expectBody: &uhttp.BodyResp{
				Content: []*sinkcommon.JSONPoint{
					{
						Measurement: "view",
						Tags: map[string]string{
							"t1": "1", "name": "some-obj-name",
						},
						Fields: map[string]interface{}{
							"f1.1": 1, "f2": 3.14,
						},
						Time: time.Unix(0, timestamp*1000*1000*1000).UTC(),
					},
				},
			},
		},

		{
			name:             `metric-point-key-with-point`,
			method:           "POST",
			url:              "/v1/write/metric",
			body:             []byte(`measurement,t1=1,t2=2 f1=1,f2=3,f3.14=3.14`),
			expectStatusCode: 200,
		},
		{
			name:             `metric-point-key-with-point-seconds`,
			method:           "POST",
			url:              "/v1/write/metric?echo_line_proto=1&precision=s",
			body:             []byte(`measurement,t1=1,t2=2 f1=1,f2=2,f3.14=3.14 ` + fmt.Sprintf("%d", timestamp)),
			expectStatusCode: 200,
			expectBody: &uhttp.BodyResp{
				Content: []*sinkcommon.JSONPoint{
					{
						Measurement: "measurement",
						Tags: map[string]string{
							"t1": "1", "t2": "2",
						},
						Fields: map[string]interface{}{
							"f1": 1, "f2": 2, "f3.14": 3.14,
						},
						Time: time.Unix(timestamp, 0).UTC(),
					},
				},
			},
		},
		{
			name:             `metric-point-key-with-point-nanoseconds`,
			method:           "POST",
			url:              "/v1/write/metric?echo_line_proto=1&precision=n",
			body:             []byte(`measurement,t1=1,t2=2 f1=1,f2=2,f3.14=3.14 ` + fmt.Sprintf("%d", timestamp*1000*1000*1000)),
			expectStatusCode: 200,
			expectBody: &uhttp.BodyResp{
				Content: []*sinkcommon.JSONPoint{
					{
						Measurement: "measurement",
						Tags: map[string]string{
							"t1": "1", "t2": "2",
						},
						Fields: map[string]interface{}{
							"f1": 1, "f2": 2, "f3.14": 3.14,
						},
						Time: time.Unix(0, timestamp*1000*1000*1000).UTC(),
					},
				},
			},
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

		// global-host-tag
		{
			name: `with-global-host-tags`,
			globalHostTags: map[string]string{
				"host": "my-testing",
			},
			method:           "POST",
			url:              "/v1/write/object?echo_line_proto=1",
			contentType:      "application/json",
			body:             []byte(`[{"measurement":"object-class","tags":{"name": "1"}, "fields":{"f1":1, "message": "dump object message"}, "time": 123}]`),
			expectStatusCode: 200,
			expectBody: &uhttp.BodyResp{
				Content: []*sinkcommon.JSONPoint{
					{
						Measurement: "object-class",
						Tags: map[string]string{
							"name": "1", "host": "my-testing",
						},
						Fields: map[string]interface{}{
							"f1": 1, "message": "dump object message",
						},
						Time: time.Unix(0, 123).UTC(),
					},
				},
			},
		},

		// global-env-tag
		{
			name: `with-global-env-tags`,
			globalHostTags: map[string]string{
				"host": "my-testing",
			},

			globalEnvTags: map[string]string{
				"cluster": "my-cluster",
			},

			method:           "POST",
			url:              "/v1/write/object?echo_line_proto=1&ignore_global_host_tags=1&ignore_global_tags=1&global_election_tags=1", // global-host-tag ignored
			contentType:      "application/json",
			body:             []byte(`[{"measurement":"object-class","tags":{"name": "1"}, "fields":{"f1":1, "message": "dump object message"}, "time": 123}]`),
			expectStatusCode: 200,
			expectBody: &uhttp.BodyResp{
				Content: []*sinkcommon.JSONPoint{
					{
						Measurement: "object-class",
						Tags: map[string]string{
							"name":    "1",
							"cluster": "my-cluster",
						},
						Fields: map[string]interface{}{
							"f1": 1, "message": "dump object message",
						},
						Time: time.Unix(0, 123).UTC(),
					},
				},
			},
		},
	}

	errEq := func(e1, e2 *uhttp.HttpError) bool {
		return e1.ErrCode == e2.ErrCode
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			point.ClearGlobalTags()

			if tc.globalHostTags != nil {
				for k, v := range tc.globalHostTags {
					point.SetGlobalHostTags(k, v)
				}
			}

			if tc.globalEnvTags != nil {
				for k, v := range tc.globalEnvTags {
					point.SetGlobalElectionTags(k, v)
				}
			}

			var resp *http.Response
			var err error
			switch tc.method {
			case "POST":
				resp, err = http.Post(fmt.Sprintf("%s%s", ts.URL, tc.url), tc.contentType, bytes.NewBuffer(tc.body))
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

// go test -v -timeout 30s -run ^Test_getTimeFromInt64$ gitlab.jiagouyun.com/cloudcare-tools/datakit/http
func Test_getTimeFromInt64(t *testing.T) {
	const timestamp = 1000000000 // 2001-09-09 01:46:40 +0000 UTC

	cases := []struct {
		name string
		in   int64
		opt  *lp.Option
		out  string
	}{
		{
			name: "seconds_1",
			in:   timestamp,
			opt:  &lp.Option{Precision: "s"},
			out:  "2001-09-09 01:46:40 +0000 UTC",
		},
		{
			name: "nanoseconds_1",
			in:   timestamp * 1000 * 1000 * 1000, // nanoseconds
			out:  "2001-09-09 01:46:40 +0000 UTC",
		},
		{
			name: "hours",
			in:   1,
			opt:  &lp.Option{Precision: "h"},
			out:  "1970-01-01 01:00:00 +0000 UTC",
		},
		{
			name: "minutes",
			in:   2,
			opt:  &lp.Option{Precision: "m"},
			out:  "1970-01-01 00:02:00 +0000 UTC",
		},
		{
			name: "seconds",
			in:   3,
			opt:  &lp.Option{Precision: "s"},
			out:  "1970-01-01 00:00:03 +0000 UTC",
		},
		{
			name: "millseconds",
			in:   4,
			opt:  &lp.Option{Precision: "ms"},
			out:  "1970-01-01 00:00:00.000004 +0000 UTC",
		},
		{
			name: "microseconds",
			in:   5,
			opt:  &lp.Option{Precision: "u"},
			out:  "1970-01-01 00:00:00.005 +0000 UTC",
		},
		{
			name: "nanoseconds",
			in:   6,
			opt:  &lp.Option{Precision: "n"},
			out:  "1970-01-01 00:00:00.000000006 +0000 UTC",
		},
		{
			name: "nanoseconds_with_no_precision",
			in:   7,
			out:  "1970-01-01 00:00:00.000000007 +0000 UTC",
		},
	}

	for _, tc := range cases {
		out := getTimeFromInt64(tc.in, tc.opt)
		// fmt.Printf("name = %s, out = %s\n", tc.name, out.String())
		assert.Equal(t, tc.out, out.String())
	}
}
