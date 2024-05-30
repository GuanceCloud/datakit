// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	T "testing"
	"time"

	lp "github.com/GuanceCloud/cliutils/lineproto"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/ipdb"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

func BenchmarkHandleWriteBody(b *testing.B) {
	body := []byte(`abc,t1=b,t2=d f1=123,f2=3.4,f3="strval" 1624550216
abc,t1=b,t2=d f1=123,f2=3.4,f3="strval" 1624550216`)

	for n := 0; n < b.N; n++ {
		if _, err := HandleWriteBody(body, point.LineProtocol, point.WithPrecision(point.PrecS)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHandleProtobufWriteBody(b *testing.B) {
	kvs := point.NewTags(map[string]string{
		"t1": "b",
		"t2": "d",
	})
	kvs = append(kvs, point.NewKVs(map[string]interface{}{
		"f1": 123.0,
		"f2": 3.4,
		"f3": "strval",
	})...)

	pt := point.NewPointV2("abc", kvs, append(point.CommonLoggingOptions(), point.WithTime(
		time.Unix(1624550216, 0)))...)

	enc := point.GetEncoder(point.WithEncEncoding(point.Protobuf))
	defer point.PutEncoder(enc)
	pts, _ := enc.Encode([]*point.Point{pt})
	body := pts[0]
	for n := 0; n < b.N; n++ {
		if _, err := HandleWriteBody(body, point.Protobuf, point.WithPrecision(point.PrecS)); err != nil {
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
		if _, err := HandleWriteBody(body, point.JSON, point.WithPrecision(point.PrecS)); err != nil {
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
		enc  point.Encoding
		npts int

		opts      []point.Option
		expectPts []*point.Point
	}{
		{
			name: `invalid-json`,
			body: []byte(`[
			{
				"measurement":"abc",
				"fields": {"f1": 123, "f.2": 3.4, "f3": "strval"},
				"time": 123
			}, # extra ','
			]`),
			enc:  point.JSON,
			fail: true,
		},

		{
			name: `json-tag-exceed-limit`,

			opts: []point.Option{
				point.WithMaxTags(1),
			},

			body: []byte(`[
			{
				"measurement":"abc",
				"fields": {"f1": 123, "f2": 3.4, "f3": "strval"},
				"tags": {"t1": "abc", "t2": "def"},
				"time": 123
			}
			]`),
			enc:  point.JSON,
			npts: 1,
		},

		{
			name: `json-invalid-field-key-with-dot`,
			body: []byte(`[
			{
				"measurement":"abc",
				"fields": {"f1": 123, "f.2": 3.4, "f3": "strval"},
				"time": 123
			}
			]`),
			opts: []point.Option{
				point.WithDotInKey(false),
			},
			enc: point.JSON,

			npts: 1,
			expectPts: []*point.Point{
				point.NewPointV2("abc",
					point.NewKVs(nil).AddV2("f1", 123.0 /* int value in json are float */, true).
						AddV2("f_2", 3.4, true).
						AddV2("f3", "strval", true), point.WithTimestamp(123)),
			},
		},

		{
			name: `invalid-field`,
			body: []byte(`[
			{
				"measurement":"abc",
				"time":123,
				"fields": {"f1": 123, "f2": 3.4, "f3": "strval", "fx": [1,2,3]}
			}
			]`),
			enc: point.JSON,

			expectPts: []*point.Point{
				point.NewPointV2("abc",
					point.NewKVs(nil).AddV2("f1", 123.0, true).
						AddV2("f2", 3.4, true).
						AddV2("fx", []float64{1, 2, 3}, true).
						AddV2("f3", "strval", true), point.WithTimestamp(123)),
			},
		},

		{
			name: `json-body`,
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
			enc:  point.JSON,
			npts: 2,
		},

		{
			name: `json-body-with-timestamp`,
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

			opts: []point.Option{
				point.WithPrecision(point.PrecS),
			},

			npts: 2,
			enc:  point.JSON,
		},

		{
			name: `raw-point-body-with/wthout-timestamp`,

			opts: []point.Option{
				point.WithPrecision(point.PrecS),
			},

			body: []byte(`error,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
            view,t1=tag2,t2=tag2 f1=1.0,f2=2i,f3="abc" 1621239130
            resource,t1=tag3,t2=tag2 f1=1.0,f2=2i,f3="abc" 1621239130
            long_task,t1=tag4,t2=tag2 f1=1.0,f2=2i,f3="abc"
            action,t1=tag5,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			npts: 5,
		},

		{
			name: "invalid-line-protocol",
			body: []byte(`test t1=abc f1=1i,f2=2,f3="str"`),
			fail: true,
		},

		{
			name: "multi-line protocol",
			body: []byte(`test,t1=abc f1=1i,f2=2,f3="str"
test,t1=abc f1=1i,f2=2,f3="str"
test,t1=abc f1=1i,f2=2,f3="str"`),
			npts: 3,
		},

		{
			name: "lp-on-metric-drop-sting-field",
			body: []byte(`
m1,t1=a f1=1i,f2="abc"`),
			npts: 1,
			opts: point.DefaultMetricOptions(),
		},

		{
			name: "lp-on-object",
			body: []byte(`
m1,name=a f1=1i,f2="abc"`),
			npts: 1,
			opts: point.DefaultObjectOptions(),
		},

		{
			name: "lp-on-object-add-default-name",
			body: []byte(`
m1 f1=1i,f2="abc"`),
			npts: 1,
			opts: point.DefaultObjectOptions(),
		},

		{
			name: "lp-on-logging-add-default-status",
			body: []byte(`
m1 f1=1i,f2="abc" 123`),
			npts: 1,
			opts: point.DefaultLoggingOptions(),
		},
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pts, err := HandleWriteBody(tc.body, tc.enc, tc.opts...)

			if len(tc.expectPts) > 0 {
				for i, pt := range tc.expectPts {
					ok, why := pt.EqualWithReason(pts[i])
					assert.Truef(t, ok, "why: %s, expect: %s, got %s", why, pt.Pretty(), pts[i].Pretty())
				}
			}

			if tc.npts > 0 {
				assert.Len(t, pts, tc.npts)
			}

			if tc.fail {
				assert.Error(t, err, "case[%d] expect fail, but ok", i)
				t.Logf("[%d] handle body failed: %s", i, err)
				return
			}

			if err != nil && !tc.fail {
				t.Errorf("[FAIL][%d] handle body failed: %s", i, err)
				return
			}

			for _, pt := range pts {
				t.Log(pt.Pretty())
			}
		})
	}

	for _, tc := range []struct {
		name         string
		body, expect []byte
		enc          point.Encoding
		opts         []point.Option
	}{
		{
			"object-no-name",
			[]byte(`[{ "measurement": "some", "fields": { "f1": 123 }, "time": 123 }]`),
			[]byte(`some f1=123,name="default" 123`),
			point.JSON,
			point.DefaultObjectOptions(),
		},
		{
			"logging-no-status",
			[]byte(`[{ "measurement": "some", "fields": { "f1": 123 }, "time": 123 }]`),
			[]byte(`some f1=123,status="unknown" 123`),
			point.JSON,
			point.DefaultLoggingOptions(),
		},
		{
			"logging-replace-dot-with_",
			[]byte(`[{ "measurement": "some", "fields": { "f.1": 123,"status":"info" }, "time": 123 }]`),
			[]byte(`some f_1=123,status="info" 123`),
			point.JSON,
			point.DefaultLoggingOptions(),
		},
		{
			"default-measurement",
			[]byte(`[{ "measurement": "", "fields": { "f1": 123 }, "time": 123 }]`),
			[]byte(`__default f1=123 123`),
			point.JSON,
			nil,
		},
		{
			"drop-string-on-M",
			[]byte(`[{ "measurement": "some", "fields": { "f1": 123, "str": "drop-me" }, "time": 123 }]`),
			[]byte(`some f1=123 123`),
			point.JSON,
			point.DefaultMetricOptions(),
		},
		{
			"drop-nil",
			[]byte(`[{ "measurement": "some", "fields": { "f1": 123, "f2": null }, "time": 123 }]`),
			[]byte(`some f1=123 123`),
			point.JSON,
			nil,
		},
	} {
		t.Run("auto-fix-"+tc.name, func(t *T.T) {
			pts, err := HandleWriteBody(tc.body, tc.enc, tc.opts...)
			assert.NoError(t, err, "got error %s", err)
			require.Len(t, pts, 1)
			assert.Equal(t, string(tc.expect), pts[0].LineProto())

			for _, pt := range pts {
				t.Logf("pt: %s", pt.Pretty())
				t.Logf("lp: %s", pt.LineProto())
			}
		})
	}
}

type apiWriteMock struct {
	t *testing.T
}

func (x *apiWriteMock) Feed(point.Category, []*point.Point, []io.FeedOption) error {
	x.t.Helper()
	x.t.Log("mock feed impl")
	return nil // do nothing
}

func (x *apiWriteMock) Query(ip string) (*ipdb.IPdbRecord, error) {
	x.t.Helper()
	x.t.Log("mock geoInfo impl")
	return nil, nil // do nothing
}

func (x *apiWriteMock) GetSourceIP(req *http.Request) (string, string) {
	return "", ""
}

func TestAPIWrite(t *testing.T) {
	router := gin.New()
	router.Use(uhttp.RequestLoggerMiddleware)
	router.POST("/v1/write/:category", rawHTTPWraper(nil, apiWrite, &apiWriteMock{t: t}))

	ts := httptest.NewServer(router)
	defer ts.Close()

	time.Sleep(time.Second)

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
			url:              "/v1/write/logging?strict=1",
			body:             []byte(`some-source,tag1=1,tag2=2,tag.3=3 f1=1`),
			expectStatusCode: 400,
		},

		{
			name:             `unstrict-write-logging-with-point-in-key`,
			method:           "POST",
			url:              "/v1/write/logging?strict=1",
			body:             []byte(`some-source,tag1=1,tag2=2,tag.3=3 f1=1`),
			expectStatusCode: 400,
		},

		{
			name:             `write-logging-json`,
			method:           "POST",
			url:              "/v1/write/logging?strict=1",
			body:             []byte(`[{"measurement":"abc", "tags": {"t1": "xxx"}, "fields":{"f1": 1.0}}]`),
			contentType:      "application/json",
			expectStatusCode: 400,
		},

		{
			name:             `write-logging-json-loose`,
			method:           "POST",
			url:              "/v1/write/logging",
			body:             []byte(`[{"measurement":"abc", "tags": {"t1": "xxx"}, "fields":{"f1": 1.0}}]`),
			contentType:      "application/json",
			expectStatusCode: 200,
		},

		{
			name:             `write-json-with-precision`,
			method:           "POST",
			url:              "/v1/write/metric?echo_json=1&precision=s",
			body:             []byte(`[{"measurement":"abc", "tags": {"t1": "xxx"}, "fields":{"f1": 1.0}, "time":123}]`),
			contentType:      "application/json",
			expectStatusCode: 200,
			expectBody: []*point.JSONPoint{
				{
					Measurement: "abc",
					Tags: map[string]string{
						"t1": "xxx",
					},
					Fields: map[string]interface{}{
						"f1": 1.0,
					},
					Time: 123000000000,
				},
			},
		},

		{
			name:             `write-logging(line-proto-loose)`,
			method:           "POST",
			url:              "/v1/write/logging?loose=1",
			body:             []byte(`xxx-source,t1=1 f1=1i`),
			expectStatusCode: 200,
		},

		{
			name:             `write-logging-lineproto`,
			method:           "POST",
			url:              "/v1/write/logging?strict=1",
			body:             []byte(`xxx-source,t1=1 f1=1i`),
			expectStatusCode: 400,
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
			name:             `write-logging-with-source-loose`,
			method:           "POST",
			url:              "/v1/write/logging?source=abc&loose=1",
			body:             []byte(`xxx-source,t1=1 f1=1i`),
			expectStatusCode: 200,
		},

		//--------------------------------------------
		// RUM cases
		//--------------------------------------------
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
			name:             `write-rum-json`,
			method:           "POST",
			url:              "/v1/write/rum?disable_pipeline=1",
			contentType:      "application/json",
			body:             []byte(`[{"measurement":"view","tags":{"t1": "1"}, "fields":{"f1":"1i"}}]`),
			expectStatusCode: 200,
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
			name:        `metric-invalid-json-point`,
			method:      "POST",
			url:         "/v1/write/metric",
			contentType: "application/json",

			// the JSON missing leading `['
			body: []byte(`{"measurement":"view","tags":{"t1": "1", "name": "some-obj-name"}, "fields":{"f1.1":1, "f2": 3.14}}]`),

			expectBody:       ErrInvalidJSONPoint,
			expectStatusCode: 400,
		},
		{
			name:             `metric-json-point-key-with-point-seconds`,
			method:           "POST",
			url:              "/v1/write/metric?precision=s&echo_json=1",
			contentType:      "application/json",
			body:             []byte(`[{"measurement":"view","tags":{"t1": "1", "name": "some-obj-name"}, "fields":{"f1.1":1, "f2": 3.14}, "time": 123}]`),
			expectStatusCode: 200,
			expectBody: []*point.JSONPoint{
				{
					Measurement: "view",
					Tags: map[string]string{
						"t1": "1", "name": "some-obj-name",
					},
					Fields: map[string]interface{}{
						"f1.1": 1, "f2": 3.14,
					},
					Time: 123000000000,
				},
			},
		},

		{
			name:             `metric-json-point-key-with-point-nanoseconds`,
			method:           "POST",
			url:              "/v1/write/metric?echo_json=1&precision=n",
			contentType:      "application/json",
			body:             []byte(`[{"measurement":"view","tags":{"t1": "1", "name": "some-obj-name"}, "fields":{"f1.1":1, "f2": 3.14}, "time":123000000000}]`),
			expectStatusCode: 200,
			expectBody: []*point.JSONPoint{
				{
					Measurement: "view",
					Tags: map[string]string{
						"t1": "1", "name": "some-obj-name",
					},
					Fields: map[string]interface{}{
						"f1.1": 1, "f2": 3.14,
					},
					Time: 123000000000,
				},
			},
		},

		{
			name:             `metric-point-key-with-dot`,
			method:           "POST",
			url:              "/v1/write/metric",
			body:             []byte(`measurement,t1=1,t2=2 f1=1,f2=3,f3.14=3.14`),
			expectStatusCode: 200,
		},

		{
			name:             `metric-point-key-with-point-seconds-echo-json`,
			method:           "POST",
			url:              "/v1/write/metric?echo_json=1&precision=s",
			body:             []byte(`measurement,t1=1,t2=2 f1=1,f2=2,f3.14=3.14 123`),
			expectStatusCode: 200,
			expectBody: []*point.JSONPoint{
				{
					Measurement: "measurement",
					Tags: map[string]string{
						"t1": "1",
						"t2": "2",
					},
					Fields: map[string]interface{}{
						"f1":    1,
						"f2":    2,
						"f3.14": 3.14,
					},
					Time: 123000000000,
				},
			},
		},

		{
			name:             `metric-point-key-with-point-seconds-echo-lineproto`,
			method:           "POST",
			url:              "/v1/write/metric?echo_line_proto=1&precision=s",
			body:             []byte(`measurement,t1=1,t2=2 f1=1,f2=2,f3.14=3.14 123`),
			expectStatusCode: 200,
			expectBody:       []byte(`measurement,t1=1,t2=2 f1=1,f2=2,f3.14=3.14 123000000000`),
		},

		{
			name:   `multi-line-proto`,
			method: "POST",
			url:    "/v1/write/metric?echo_line_proto=1&precision=n",
			body: []byte(`measurement-1,t1=1,t2=2 f1=1,f2=2,f3.14=3.14 123
measurement-2,t1=1,t2=2 f1=1,f2=2,f3.14=3.14 123`),
			expectStatusCode: 200,
			expectBody: []byte(`measurement-1,t1=1,t2=2 f1=1,f2=2,f3.14=3.14 123
measurement-2,t1=1,t2=2 f1=1,f2=2,f3.14=3.14 123`),
		},

		{
			name:   `multi-line-proto-with-dyn-precision`,
			method: "POST",
			url:    "/v1/write/metric?echo_line_proto=1",
			body: []byte(`measurement-1,t1=1,t2=2 f1=1,f2=2,f3.14=3.14 123
measurement-2,t1=1,t2=2 f1=1,f2=2,f3.14=3.14 123`),
			expectStatusCode: 200,
			expectBody: []byte(`measurement-1,t1=1,t2=2 f1=1,f2=2,f3.14=3.14 123000000000
measurement-2,t1=1,t2=2 f1=1,f2=2,f3.14=3.14 123000000000`),
		},

		{
			name:             `metric-point-key-with-point-nanoseconds`,
			method:           "POST",
			url:              "/v1/write/metric?echo_json=1&precision=n",
			body:             []byte(`measurement,t1=1,t2=2 f1=1,f2=2,f3.14=3.14 123000000000`),
			expectStatusCode: 200,
			expectBody: []*point.JSONPoint{
				{
					Measurement: "measurement",
					Tags: map[string]string{
						"t1": "1", "t2": "2",
					},
					Fields: map[string]interface{}{
						"f1": 1, "f2": 2, "f3.14": 3.14,
					},
					Time: 123000000000,
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
			name:             `write-object-json-missing-name-tag`,
			method:           "POST",
			url:              "/v1/write/object?strict=1",
			contentType:      "application/json",
			body:             []byte(`[{"measurement":"object-class","tags":{"t1": "1"}, "fields":{"f1":"1i", "message": "dump object message"}}]`),
			expectStatusCode: 400,
			expectBody:       ErrInvalidJSONPoint,
		},

		// global-host-tag
		{
			name: `with-global-host-tags`,
			globalHostTags: map[string]string{
				"host": "my-testing",
			},
			method:           "POST",
			url:              "/v1/write/object?echo_json=1&precision=n",
			contentType:      "application/json",
			body:             []byte(`[{"measurement":"object-class","tags":{"name": "1"}, "fields":{"f1":1, "message": "dump object message"}, "time": 123}]`),
			expectStatusCode: 200,
			expectBody: []*point.JSONPoint{
				{
					Measurement: "object-class",
					Tags: map[string]string{
						"name": "1",
					},
					Fields: map[string]interface{}{
						"f1": 1, "message": "dump object message",
					},
					Time: 123,
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
			url:              "/v1/write/object?echo_json=1&precision=n&ignore_global_host_tags=1&ignore_global_tags=1&global_election_tags=1", // global-host-tag ignored
			contentType:      "application/json",
			body:             []byte(`[{"measurement":"object-class","tags":{"name": "1"}, "fields":{"f1":1, "message": "dump object message"}, "time": 123}]`),
			expectStatusCode: 200,
			expectBody: []*point.JSONPoint{
				{
					Measurement: "object-class",
					Tags: map[string]string{
						"name": "1",
					},
					Fields: map[string]interface{}{
						"f1": 1, "message": "dump object message",
					},
					Time: 123,
				},
			},
		},
	}

	errEq := func(e1, e2 *uhttp.HttpError) bool {
		return e1.ErrCode == e2.ErrCode
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			datakit.ClearGlobalTags()

			if tc.globalHostTags != nil {
				for k, v := range tc.globalHostTags {
					datakit.SetGlobalHostTags(k, v)
				}
			}

			if tc.globalEnvTags != nil {
				for k, v := range tc.globalEnvTags {
					datakit.SetGlobalElectionTags(k, v)
				}
			}

			var resp *http.Response
			var err error
			switch tc.method {
			case "POST":

				resp, err = http.Post(fmt.Sprintf("%s%s", ts.URL, tc.url),
					tc.contentType,
					bytes.NewBuffer(tc.body))
				require.NoError(t, err)
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
				require.NoError(t, err)
			}

			t.Logf("body: %q", body)

			defer resp.Body.Close() //nolint:errcheck

			if tc.expectBody != nil {
				switch x := tc.expectBody.(type) {
				case *uhttp.HttpError:
					var e uhttp.HttpError
					if err := json.Unmarshal(body, &e); err != nil {
						t.Error(err)
					} else {
						assert.True(t, errEq(x, &e), "\n%+#v\nexpect to equal\n%+#v", x, &e)
					}

				default:
					t.Logf("get expect type: %s", reflect.TypeOf(tc.expectBody).String())

					switch x := tc.expectBody.(type) {
					case []byte:
						assert.Equal(t, string(x), string(body))
					default:
						j, err := json.Marshal(tc.expectBody)
						if err != nil {
							t.Errorf("json.Marshal: %s", err)
							return
						}
						assert.Equal(t, string(j), string(body))
					}
				}
			}

			assert.Equal(t, tc.expectStatusCode, resp.StatusCode)
		})
	}
}

// go test -v -timeout 30s -run ^Test_getTimeFromInt64$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi
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

func getTimeFromInt64(n int64, opt *lp.Option) time.Time {
	if opt != nil {
		switch opt.Precision {
		case "h":
			return time.Unix(n*3600, 0).UTC()
		case "m":
			return time.Unix(n*60, 0).UTC()
		case "s":
			return time.Unix(n, 0).UTC()
		case "ms":
			return time.Unix(0, n*1000).UTC()
		case "u":
			return time.Unix(0, n*1000000).UTC()
		default:
		}
	}

	// nanoseconds
	return time.Unix(0, n).UTC()
}

var (
	// prepare 2 sample point, each point got all-types fields
	samplePoints = func() []*point.Point {
		var kvs point.KVs
		kvs = kvs.Add("f_i", 123, false, false)
		kvs = kvs.Add("f_d", []byte(`hello`), false, false)
		kvs = kvs.Add("f_b", false, false, false)
		kvs = kvs.Add("f_f", 3.14, false, false)
		kvs = kvs.Add("f_s", "world", false, false)
		kvs = kvs.Add("f_u", uint(321), false, false)
		kvs = kvs.Add("f_a", point.MustNewAnyArray(1, 2, 3, 4, 5), false, false)
		kvs = kvs.Add("f_d_arr", point.MustNewAnyArray([]byte("hello"), []byte("world")), false, false)
		kvs = kvs.Add("t", "some-tag-val", true, false)

		return []*point.Point{
			point.NewPointV2("sample_point_1", kvs, point.WithTimestamp(int64(123))),
			point.NewPointV2("sample_point_2", kvs, point.WithTimestamp(int64(123))),
		}
	}()

	// prepare json/pbjson/line-protocol represent of the 2 points
	JSON, _ = json.Marshal(samplePoints)

	pbJSON = func() []byte {
		for _, pt := range samplePoints {
			pt.SetFlag(point.Ppb)
		}

		defer func() {
			for _, pt := range samplePoints {
				pt.ClearFlag(point.Ppb)
			}
		}()

		x, _ := json.Marshal(samplePoints)
		return x
	}()

	lineProtocol = func() []byte {
		enc := point.GetEncoder(point.WithEncEncoding(point.LineProtocol))
		defer point.PutEncoder(enc)
		arr, _ := enc.Encode(samplePoints)
		return arr[0]
	}()

	pb = func() []byte {
		enc := point.GetEncoder(point.WithEncEncoding(point.Protobuf))
		defer point.PutEncoder(enc)
		arr, _ := enc.Encode(samplePoints)
		return arr[0]
	}()
)

func TestAPIV1Write(t *T.T) {
	t.Logf("lineProtocol:\n%s", string(lineProtocol))
	t.Logf("JSON:\n%s", string(JSON))

	t.Run("decode-bytes-array-line-proto", func(t *T.T) {
		dec := point.GetDecoder(point.WithDecEncoding(point.LineProtocol))
		pts, err := dec.Decode(lineProtocol)
		assert.NoError(t, err)

		for _, pt := range pts {
			t.Logf("pt: %s", pt.Pretty())
		}
	})

	t.Run("basic", func(t *T.T) {
		req := httptest.NewRequest("POST", "/v1/write/metric", bytes.NewBuffer(lineProtocol))

		wr := GetAPIWriteResult()
		defer PutAPIWriteResult(wr)

		assert.NoError(t, wr.APIV1Write(req))

		assert.Equal(t, point.Metric, wr.Category)
		t.Logf("pt[0]: %s", wr.Points[0].Pretty())

		assert.Len(t, wr.Points, len(samplePoints))
	})

	// test precision=xxx
	precSpecs := map[string]int64{
		"u":       123000,
		"ms":      123000000,
		"s":       123000000000,
		"m":       123 * 60 * 1000000000,
		"h":       123 * 60 * 60 * 1000000000,
		"invalid": 123, // default as second
	}

	for k, v := range precSpecs {
		t.Run("precision-"+k, func(t *T.T) {
			req := httptest.NewRequest("POST", fmt.Sprintf("/v1/write/metric?precision=%s", k), bytes.NewBuffer(lineProtocol))

			wr := GetAPIWriteResult()
			defer PutAPIWriteResult(wr)

			assert.NoError(t, wr.APIV1Write(req))

			assert.Equal(t, point.Metric, wr.Category)

			assert.Len(t, wr.Points, len(samplePoints))
			t.Logf("pt[0]: %s", wr.Points[0].Pretty())
			assert.Equal(t, v, wr.Points[0].Time().UnixNano())
		})
	}

	inputSpec := map[string]string{
		"/v1/write/metric?input=demo": "demo",
		"/v1/write/metric":            "datakit-http",
		"/v1/write/security":          "scheck",
		"/v1/write/rum":               "rum",
		"/v1/write/custom_object":     "custom_object",
	}

	for k, v := range inputSpec {
		t.Run("input-name-"+k, func(t *T.T) {
			req := httptest.NewRequest("POST", k, bytes.NewBuffer(lineProtocol))

			wr := GetAPIWriteResult()
			defer PutAPIWriteResult(wr)

			assert.NoError(t, wr.APIV1Write(req))

			assert.Len(t, wr.Points, len(samplePoints))
			t.Logf("pt[0]: %s", wr.Points[0].Pretty())

			assert.Equal(t, v, wr.input)

			t.Logf("wr: %+#v", wr)
		})
	}

	catSpec := map[string]point.Category{
		"/v1/write/custom_object": point.CustomObject,
		"/v1/write/invalid":       point.UnknownCategory,
		"/v1/write/keyevent":      point.KeyEvent,
		"/v1/write/logging":       point.Logging,
		"/v1/write/metric":        point.Metric,
		"/v1/write/network":       point.Network,
		"/v1/write/object":        point.Object,
		"/v1/write/security":      point.Security,
		"/v1/write/tracing":       point.Tracing,

		// TODO: not support profile
		// "/v1/write/profile":  point.Profile,
	}

	for k, v := range catSpec {
		t.Run("category-"+k, func(t *T.T) {
			req := httptest.NewRequest("POST", k, bytes.NewBuffer(lineProtocol))

			wr := GetAPIWriteResult()
			defer PutAPIWriteResult(wr)

			err := wr.APIV1Write(req)
			if wr.Category != point.UnknownCategory {
				assert.NoError(t, err)
				assert.Len(t, wr.Points, len(samplePoints))
			} else {
				assert.Error(t, err)
				t.Logf("error: %s", err)
			}

			assert.Equal(t, v, wr.Category)
		})
	}

	t.Run("ignore-global-host-tags", func(t *T.T) {
		req := httptest.NewRequest("POST", "/v1/write/metric?ignore_global_host_tags=1", bytes.NewBuffer(lineProtocol))

		wr := GetAPIWriteResult()
		defer PutAPIWriteResult(wr)

		assert.NoError(t, wr.APIV1Write(req))
		assert.True(t, wr.ignoreGlobalTags)
	})

	t.Run("enable-global-election-tags", func(t *T.T) {
		req := httptest.NewRequest("POST", "/v1/write/metric?global_election_tags=1", bytes.NewBuffer(lineProtocol))
		wr := GetAPIWriteResult()
		defer PutAPIWriteResult(wr)

		assert.NoError(t, wr.APIV1Write(req))
		assert.True(t, wr.globalElectionTags)
	})

	t.Run("input-version", func(t *T.T) {
		req := httptest.NewRequest("POST", "/v1/write/metric?version=1.0", bytes.NewBuffer(lineProtocol))
		wr := GetAPIWriteResult()
		defer PutAPIWriteResult(wr)

		assert.NoError(t, wr.APIV1Write(req))
		assert.Equal(t, "1.0", wr.inputVersion)
	})

	t.Run("source", func(t *T.T) {
		req := httptest.NewRequest("POST", "/v1/write/metric?source=demo", bytes.NewBuffer(lineProtocol))
		wr := GetAPIWriteResult()
		defer PutAPIWriteResult(wr)

		assert.NoError(t, wr.APIV1Write(req))
		assert.Equal(t, "demo.p", wr.plName)
	})

	t.Run("strict-point", func(t *T.T) {
		req := httptest.NewRequest("POST", "/v1/write/logging?strict=1", bytes.NewBuffer(lineProtocol))
		wr := GetAPIWriteResult()
		defer PutAPIWriteResult(wr)

		err := wr.APIV1Write(req)
		assert.Error(t, err)
		t.Logf("error(expected): %s", err)
		assert.Len(t, wr.Points, 0)
	})

	t.Run("echo-pbjson", func(t *T.T) {
		req := httptest.NewRequest("POST", "/v1/write/network?echo=pbjson&precision=n", bytes.NewBuffer(lineProtocol))
		wr := GetAPIWriteResult()
		defer PutAPIWriteResult(wr)

		assert.NoError(t, wr.APIV1Write(req))

		var pts []*point.Point
		assert.NoError(t, json.Unmarshal(wr.RespBody, &pts))

		for i, pt := range pts {
			assert.True(t, pt.HasFlag(point.Ppb))
			equal, why := pt.EqualWithReason(samplePoints[i])
			assert.Truef(t, equal, "exp %s\ngot %s\nwhy %s\nPOST lineproto: %s\nresponse pbjson: %s",
				samplePoints[i].Pretty(), pt.Pretty(), why, string(lineProtocol), string(wr.RespBody))
		}

		t.Logf("echo: %s", string(wr.RespBody))
	})

	t.Run("echo-lp", func(t *T.T) {
		req := httptest.NewRequest("POST", "/v1/write/network?echo=lp&precision=n", bytes.NewBuffer(lineProtocol))
		wr := GetAPIWriteResult()
		defer PutAPIWriteResult(wr)

		assert.NoError(t, wr.APIV1Write(req))
		assert.Equalf(t, lineProtocol, wr.RespBody, "exp:\n%s\ngot\n%s", string(lineProtocol), string(wr.RespBody))
		t.Logf("echo: %s", string(wr.RespBody))
	})

	t.Run("echo-json", func(t *T.T) {
		req := httptest.NewRequest("POST", "/v1/write/network?echo=json&precision=n", bytes.NewBuffer(lineProtocol))
		wr := GetAPIWriteResult()
		defer PutAPIWriteResult(wr)

		assert.NoError(t, wr.APIV1Write(req))
		assert.Equal(t, JSON, wr.RespBody)
		t.Logf("echo: %s", string(wr.RespBody))
	})

	// HTTP header encoding cases
	encSpecs := []struct {
		name,
		header string
		body []byte
	}{
		{"pb-json", "application/json", pbJSON},
		{"simple-json", "application/json", JSON},
		{"default-as-line-protocol", "not-specified", lineProtocol},
		{"protobuf", point.Protobuf.HTTPContentType(), pb},
	}

	for _, s := range encSpecs {
		t.Run("enc-"+s.name, func(t *T.T) {
			req := httptest.NewRequest("POST", "/v1/write/network?precision=n", bytes.NewBuffer(s.body))
			req.Header.Add("Content-Type", s.header)

			wr := GetAPIWriteResult()
			defer PutAPIWriteResult(wr)

			assert.NoError(t, wr.APIV1Write(req))
			assert.Len(t, wr.Points, len(samplePoints))

			for i, pt := range wr.Points {
				t.Logf("pt: %s", pt.Pretty())

				if s.name != "simple-json" {
					eq, why := samplePoints[i].EqualWithReason(pt)
					assert.Truef(t, eq, "exp: %s\ngot: %s, why: %s", samplePoints[i].Pretty(), pt.Pretty(), why)
				} else {
					// NOTE: Simple json point do not equal with samplePoints:
					//  - all int cast to float
					//  - no []byte support: json will use encoded-base64 as common string
					assert.Equal(t, "aGVsbG8=" /* = base64("hello") */, pt.Get("f_d"))
					assert.Equal(t, float64(123), pt.Get("f_i"))
				}
			}
		})
	}

	// RUM get-client-ip
	for _, tc := range []struct {
		name,
		header,
		ip,
		ipStatus,
		geoStatus string
		ipq plval.IIPQuerier
	}{
		{"x-forward-for", "X-Forwarded-For", "1.1.1.1", plval.IPTypePublic, plval.LocateStatusGEOSuccess, &ipq{}},
		{"x-real-ip", "X-Real-IP", "1.2.3.4", plval.IPTypePublic, plval.LocateStatusGEOSuccess, &ipq{}},

		// all are private IP
		{"x-real-ip-local-10", "X-Real-IP", "10.1.1.10", plval.IPTypePrivate, plval.LocateStatusGEOSuccess, &ipq{}},
		{"x-forward-for-local-192.168", "X-Forwarded-For", "192.168.0.10", plval.IPTypePrivate, plval.LocateStatusGEOSuccess, &ipq{}},
		{"x-forward-for-local-172.16", "X-Forwarded-For", "172.16.0.10", plval.IPTypePrivate, plval.LocateStatusGEOSuccess, &ipq{}},

		{"not-specified", "", "", plval.IPTypeRemoteAddr, plval.LocateStatusGEOSuccess, &ipq{}},
		{"invalid-ip-query", "X-Forwarded-For", "1.2.3.4", plval.IPTypePublic, plval.LocateStatusGEOFailure, &invalidIPQ{}},
		{"always-nil-ip-query", "X-Forwarded-For", "1.2.3.4", plval.IPTypePublic, plval.LocateStatusGEONil, &nilIPQ{}},
	} {
		t.Run("rum-with-forward-ip-"+tc.name, func(t *T.T) {
			req := httptest.NewRequest("POST", "/v1/write/rum", bytes.NewBuffer(lineProtocol))
			if len(tc.header) > 0 {
				req.Header.Add(tc.header, tc.ip)
			}

			wr := GetAPIWriteResult()
			PutAPIWriteResult(wr)
			wr.IPQuerier = tc.ipq

			assert.NoError(t, wr.APIV1Write(req))
			if len(tc.header) > 0 {
				assert.Equalf(t, tc.ip, wr.SrcIP, "got %s", wr.SrcIP)
			} else {
				assert.Truef(t, len(wr.SrcIP) > 0, "got empty src-ip")
				assert.Equal(t, tc.ipStatus, wr.IPStatus, "SrcIP: %s", wr.SrcIP)
			}

			assert.Equal(t, tc.geoStatus, wr.LocateStatus)
			assert.Equal(t, tc.ipStatus, wr.IPStatus)

			if tc.geoStatus == plval.LocateStatusGEOSuccess {
				for _, pt := range wr.Points {
					assert.Equal(t, mockIPRecord.Country, pt.Get("country"), "get %s", pt.Pretty())
					assert.Equal(t, mockIPRecord.Region, pt.Get("province"), "get %s", pt.Pretty())
					assert.Equal(t, mockIPRecord.City, pt.Get("city"), "get %s", pt.Pretty())
					assert.Equal(t, mockIPRecord.Isp, pt.Get("isp"), "get %s", pt.Pretty())
					assert.Equal(t, wr.SrcIP, pt.Get("ip"), "get %s", pt.Pretty())

					t.Logf("pt: %s", pt.Pretty())
				}
			}
		})
	}

	for _, tc := range []struct {
		name        string
		contentType string
		body        []byte
	}{
		{"line-proto", "", lineProtocol},
		{"json", "application/json", JSON},
		{"pbjson", "application/json", pbJSON},
	} {
		t.Run("with-pt-callback-"+tc.name, func(t *T.T) {
			ptcb := func(pt *point.Point) (*point.Point, error) {
				pt.AddTag("added-by-callback", "some-value")
				if pt.Name() == "sample_point_1" {
					return nil, nil
				} else {
					return pt, nil
				}
			}

			wr := GetAPIWriteResult()
			wr.PointCallback = ptcb
			defer PutAPIWriteResult(wr)

			req := httptest.NewRequest("POST", "/v1/write/rum", bytes.NewBuffer(tc.body))

			if len(tc.contentType) > 0 {
				req.Header.Add("Content-Type", tc.contentType)
			}

			assert.NoError(t, wr.APIV1Write(req))
			assert.Len(t, wr.Points, 1)
			assert.Equal(t, "some-value", wr.Points[0].Get("added-by-callback"))
			assert.Equal(t, "sample_point_2", wr.Points[0].Name())
		})
	}
}

func BenchmarkAPIV1Write(x *T.B) {
	x.Run("basic", func(b *T.B) {
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("POST", "/v1/write/network", bytes.NewBuffer(lineProtocol))
			wr := GetAPIWriteResult()
			defer PutAPIWriteResult(wr)

			_ = wr.APIV1Write(req)
		}
	})

	x.Run("rum-with-ip-info", func(b *T.B) {
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("POST", "/v1/write/rum", bytes.NewBuffer(lineProtocol))
			wr := GetAPIWriteResult()
			PutAPIWriteResult(wr)
			wr.IPQuerier = &ipq{}

			_ = wr.APIV1Write(req)
		}
	})

	x.Run("with-echo-pbjson", func(b *T.B) {
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("POST", "/v1/write/rum?echo=pbjson", bytes.NewBuffer(lineProtocol))
			wr := GetAPIWriteResult()
			PutAPIWriteResult(wr)
			wr.IPQuerier = &ipq{}

			_ = wr.APIV1Write(req)
		}
	})

	x.Run("with-echo-json", func(b *T.B) {
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("POST", "/v1/write/rum?echo=json", bytes.NewBuffer(lineProtocol))
			wr := GetAPIWriteResult()
			PutAPIWriteResult(wr)
			wr.IPQuerier = &ipq{}

			_ = wr.APIV1Write(req)
		}
	})

	x.Run("with-echo-line-protocol", func(b *T.B) {
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("POST", "/v1/write/rum?echo=lp", bytes.NewBuffer(lineProtocol))
			wr := GetAPIWriteResult()
			PutAPIWriteResult(wr)
			wr.IPQuerier = &ipq{}

			_ = wr.APIV1Write(req)
		}
	})
}

type ipq struct{}

var mockIPRecord = ipdb.IPdbRecord{
	Country:   "USA",
	Region:    "some-region",
	City:      "some-city",
	Isp:       "some-ips",
	Latitude:  1.0,
	Longitude: 1.0,
	Timezone:  "",
	Areacode:  "some-areacode",
}

func (q *ipq) Query(_ string) (*ipdb.IPdbRecord, error) {
	return &mockIPRecord, nil
}

func (q *ipq) GetSourceIP(req *http.Request) (string, string) {
	return plval.RequestSourceIP(req, "")
}

type invalidIPQ struct{}

func (q *invalidIPQ) Query(_ string) (*ipdb.IPdbRecord, error) {
	return nil, fmt.Errorf("allways failed IP querier")
}

func (q *invalidIPQ) GetSourceIP(req *http.Request) (string, string) {
	return plval.RequestSourceIP(req, "")
}

type nilIPQ struct{}

func (q *nilIPQ) Query(_ string) (*ipdb.IPdbRecord, error) {
	return nil, nil
}

func (q *nilIPQ) GetSourceIP(req *http.Request) (string, string) {
	return plval.RequestSourceIP(req, "")
}
