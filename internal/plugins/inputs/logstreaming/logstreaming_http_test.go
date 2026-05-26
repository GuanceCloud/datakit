// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package logstreaming testing.
package logstreaming

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sync"
	"testing"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/stretchr/testify/assert"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
)

func BenchmarkScanner(b *testing.B) {
	data := []byte(`111111111111111111111111111111111111111111111111111111
222222222222222222222222222222222222222222222222222222
333333333333333333333333333333333333333333333333333333
444444444444444444444444444444444444444444444444444444
555555555555555555555555555555555555555555555555555555
666666666666666666666666666666666666666666666666666666`)
	buf := bytes.NewBuffer(data)

	scanBuf := make([]byte, 128)

	b.Run("set-scan-buffer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			scanner := bufio.NewScanner(buf)
			scanner.Buffer(scanBuf, 64)
			for scanner.Scan() {
				// do nothing
			}

			buf.Reset()
			buf.Write(data)
		}
	})

	b.Run("no-scan-buffer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			scanner := bufio.NewScanner(buf)
			for scanner.Scan() {
				// do nothing
			}

			buf.Reset()
			buf.Write(data)
		}
	})
}

// TestInput_processLogBody 测试 processLogBody 的三种协议类型
// 修改：body 从 *bytes.Buffer 改为 io.ReadCloser，需要用 io.NopCloser 包装
func TestInput_processLogBody(t *testing.T) {
	// ==================== influxdb 测试 ====================
	pt, _ := influxdb.NewPoint("test_logging", map[string]string{"host": "hostName"}, map[string]interface{}{"message": "this is message", "status": "unknown"})
	pt2, _ := influxdb.NewPoint("test_logging", map[string]string{"host": "hostName"}, map[string]interface{}{"message": "this is message01", "status": "unknown"})

	buf := &bytes.Buffer{}
	buf.WriteString(pt.PrecisionString("ns"))
	buf.WriteByte('\n')
	buf.WriteString(pt2.PrecisionString("ns"))
	buf.WriteByte('\n')

	influxdbCheckers := []inputs.PointCheckOption{
		inputs.WithExtraTags(map[string]string{"host": "hostName"}),
		inputs.WithMeasurementCheckIgnored(true),
		inputs.WithDoc(&logstreamingMeasurement{}),
	}

	// ==================== 其他类型（流式）测试 ====================
	obuf := &bytes.Buffer{}
	obuf.Write([]byte(`this is message
this is message
this is message
`))
	otherCheckers := []inputs.PointCheckOption{
		inputs.WithMeasurementCheckIgnored(true),
		inputs.WithOptionalFields("messsage"),
	}

	feeder := dkio.NewMockedFeeder()
	type fields struct {
		IgnoreURLTags    bool
		WPConfig         *workerpool.WorkerPoolConfig
		LocalCacheConfig *storage.StorageConfig
		feeder           dkio.Feeder
	}
	type args struct {
		param *parameters
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		checks  []inputs.PointCheckOption
	}{
		{
			name: "test_influxdb",
			fields: fields{
				feeder: feeder,
			},
			args: args{
				param: &parameters{
					ignoreURLTags: false,
					url:           &url.URL{Scheme: "http", Host: "127.0.0.1", Path: "/"},
					queryValues:   url.Values{"type": []string{"influxdb"}},
					body:          io.NopCloser(buf), // 用 io.NopCloser 包装
				},
			},
			wantErr: false,
			checks:  influxdbCheckers,
		},
		{
			name: "test_others",
			fields: fields{
				feeder: feeder,
			},
			args: args{
				param: &parameters{
					ignoreURLTags: false,
					url: &url.URL{
						Scheme: "http",
						Host:   "127.0.0.1",
						Path:   "/",
					},
					queryValues: url.Values{
						"type":     []string{"txtType"},
						"pipeline": []string{"log.p"},
						"source":   []string{"testSource"},
					},
					body: io.NopCloser(obuf), // 用 io.NopCloser 包装
				},
			},
			wantErr: false,
			checks:  otherCheckers,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := &Input{
				feeder:         tt.fields.feeder,
				scanbufPool:    &sync.Pool{},
				ScanBufferSize: 4 * 1024, // 添加：设置默认的 4KB buffer
			}

			if err := ipt.processLogBody(tt.args.param); (err != nil) != tt.wantErr {
				t.Errorf("processLogBody() error = %v, wantErr %v", err, tt.wantErr)
			}

			pts, err := feeder.AnyPoints(time.Second * 2)
			if err != nil {
				t.Errorf("feeder err = %v", err)
				return
			}

			for i, point := range pts {
				msgs := inputs.CheckPoint(point, tt.checks...)
				if len(msgs) != 0 {
					t.Errorf("check err = %v", msgs)
				}
				bts, _ := point.MarshalJSON()
				t.Logf("i:%d pt :%s", i, string(bts))
			}
		})
	}
}

// ==================== 新增测试：测试流式读取优化 ====================

// TestStreamingReadSmallBody 测试小 body 流式读取
func TestStreamingReadSmallBody(t *testing.T) {
	logLines := `line 1
line 2
line 3
`

	buf := bytes.NewBufferString(logLines)

	feeder := dkio.NewMockedFeeder()
	ipt := &Input{
		IgnoreURLTags:  false,
		feeder:         feeder,
		scanbufPool:    &sync.Pool{},
		ScanBufferSize: 4 * 1024,
	}

	param := &parameters{
		ignoreURLTags: false,
		url:           &url.URL{Scheme: "http", Host: "127.0.0.1", Path: "/"},
		queryValues:   url.Values{"type": []string{"txt"}, "source": []string{"test"}},
		body:          io.NopCloser(buf),
	}

	err := ipt.processLogBody(param)
	if err != nil {
		t.Errorf("processLogBody() error = %v", err)
		return
	}

	pts, err := feeder.AnyPoints(time.Second * 2)
	if err != nil {
		t.Errorf("feeder err = %v", err)
		return
	}

	if len(pts) != 3 {
		t.Errorf("expected 3 points, got %d", len(pts))
	}

	t.Logf("TestStreamingReadSmallBody: processed %d lines successfully", len(pts))
}

// TestStreamingReadLargeBody 测试大 body 流式读取（核心优化验证）
func TestStreamingReadLargeBody(t *testing.T) {
	// 生成 1MB 的日志数据
	var buf bytes.Buffer
	numLines := 10000
	for i := 0; i < numLines; i++ {
		fmt.Fprintf(&buf, "log line %d: some log content here for streaming test\n", i)
	}
	logData := buf.Bytes()

	bodyReader := bytes.NewReader(logData)

	feeder := dkio.NewMockedFeeder()
	ipt := &Input{
		IgnoreURLTags:  false,
		feeder:         feeder,
		scanbufPool:    &sync.Pool{},
		ScanBufferSize: 4 * 1024,
	}

	param := &parameters{
		ignoreURLTags: false,
		url:           &url.URL{Scheme: "http", Host: "127.0.0.1", Path: "/"},
		queryValues:   url.Values{"type": []string{"txt"}, "source": []string{"test"}},
		body:          io.NopCloser(bodyReader),
	}

	err := ipt.processLogBody(param)
	if err != nil {
		t.Errorf("processLogBody() error = %v", err)
		return
	}

	pts, err := feeder.AnyPoints(time.Second * 5)
	if err != nil {
		t.Errorf("feeder err = %v", err)
		return
	}

	if len(pts) != numLines {
		t.Errorf("expected %d points, got %d", numLines, len(pts))
	}

	t.Logf("TestStreamingReadLargeBody: processed %d lines from %.2f MB (streaming mode)",
		len(pts), float64(len(logData))/1024/1024)
}

// TestInfluxdbStillReadsFullBody 验证 influxdb 类型仍然一次性读取
func TestInfluxdbStillReadsFullBody(t *testing.T) {
	// Influx Line Protocol 格式
	body := `weather,location=us-midwest temperature=82 1465839830100400200
weather,location=us-east temperature=75 1465839830100400200`

	buf := bytes.NewBufferString(body)

	feeder := dkio.NewMockedFeeder()
	ipt := &Input{
		IgnoreURLTags:  false,
		feeder:         feeder,
		scanbufPool:    &sync.Pool{},
		ScanBufferSize: 4 * 1024,
	}

	param := &parameters{
		ignoreURLTags: false,
		url:           &url.URL{Scheme: "http", Host: "127.0.0.1", Path: "/"},
		queryValues:   url.Values{"type": []string{"influxdb"}, "source": []string{"test"}},
		body:          io.NopCloser(buf),
	}

	err := ipt.processLogBody(param)
	if err != nil {
		t.Errorf("processLogBody() error = %v", err)
		return
	}

	pts, err := feeder.AnyPoints(time.Second * 2)
	if err != nil {
		t.Errorf("feeder err = %v", err)
		return
	}

	if len(pts) != 2 {
		t.Errorf("expected 2 points for influxdb, got %d", len(pts))
	}

	t.Logf("TestInfluxdbStillReadsFullBody: processed %d influxdb points successfully", len(pts))
}

// TestFirelensStillReadsFullBody 验证 firelens 类型仍然一次性读取
func TestFirelensStillReadsFullBody(t *testing.T) {
	// CloudWatch Logs JSON 格式
	body := `[
  {
    "log": "2021-10-01 12:34:56 INFO request completed",
    "date": 1633078496000,
    "source": "cwlogs"
  },
  {
    "log": "2021-10-01 12:34:57 ERROR connection failed",
    "date": 1633078497000,
    "source": "cwlogs"
  }
]`

	buf := bytes.NewBufferString(body)

	feeder := dkio.NewMockedFeeder()
	ipt := &Input{
		IgnoreURLTags:  false,
		feeder:         feeder,
		scanbufPool:    &sync.Pool{},
		ScanBufferSize: 4 * 1024,
	}

	param := &parameters{
		ignoreURLTags: false,
		url:           &url.URL{Scheme: "http", Host: "127.0.0.1", Path: "/"},
		queryValues:   url.Values{"type": []string{"firelens"}, "source": []string{"test"}},
		body:          io.NopCloser(buf),
	}

	err := ipt.processLogBody(param)
	if err != nil {
		t.Errorf("processLogBody() error = %v", err)
		return
	}

	pts, err := feeder.AnyPoints(time.Second * 2)
	if err != nil {
		t.Errorf("feeder err = %v", err)
		return
	}

	if len(pts) != 2 {
		t.Errorf("expected 2 points for firelens, got %d", len(pts))
	}

	t.Logf("TestFirelensStillReadsFullBody: processed %d firelens points successfully", len(pts))
}

func TestFirelensKeepsNestedMapAndListFields(t *testing.T) {
	body := `[
  {
    "log": {"msg": "request completed", "code": 200},
    "date": 1633078496000,
    "source": "cwlogs",
    "attrs": {"cluster": "prod", "pod": "api-1"},
    "items": [1, "two", {"nested": true}]
  }
]`

	feeder := dkio.NewMockedFeeder()
	ipt := &Input{
		IgnoreURLTags:  false,
		feeder:         feeder,
		scanbufPool:    &sync.Pool{},
		ScanBufferSize: 4 * 1024,
	}

	param := &parameters{
		ignoreURLTags: false,
		url:           &url.URL{Scheme: "http", Host: "127.0.0.1", Path: "/"},
		queryValues:   url.Values{"type": []string{"firelens"}, "source": []string{"test"}},
		body:          io.NopCloser(bytes.NewBufferString(body)),
	}

	err := ipt.processLogBody(param)
	assert.NoError(t, err)

	pts, err := feeder.AnyPoints(time.Second * 2)
	assert.NoError(t, err)
	if assert.Len(t, pts, 1) {
		pt := pts[0]

		assert.Equal(t, `{"cluster":"prod","pod":"api-1"}`, pt.Get("attrs"))
		assert.Equal(t, `[1,"two",{"nested":true}]`, pt.Get("items"))
		assert.Equal(t, `{"code":200,"msg":"request completed"}`, pt.Get("message"))
		assert.Equal(t, "cwlogs", pt.Get("firelens_source"))
	}
}

func TestFirehoseBody(t *testing.T) {
	rd := requestData{
		RequestID: "xxxxx-001",
		Timestamp: time.Now().UnixMilli(),
		Records: []*Record{
			{
				Data: []byte("hello data 1111111111"),
			},
			{
				Data: []byte("hello data 2222222222"),
			},
		},
	}
	bts, err := json.Marshal(rd)
	assert.Nil(t, err)
	assert.NotNil(t, bts)

	param := &parameters{
		ignoreURLTags: false,
		url:           &url.URL{Scheme: "http", Host: "127.0.0.1", Path: "/"},
		queryValues:   url.Values{"type": []string{"firehose"}, "source": []string{"test"}},
		body:          io.NopCloser(bytes.NewBuffer(bts)),
		remoteIP:      "127.0.0.1",
		headers: map[string]string{
			"arn":                       "arn",
			"request_id":                "xxxxx-0001",
			"firehose_protocol_version": "1.0.0",
		},
	}

	feeder := dkio.NewMockedFeeder()
	ipt := &Input{
		IgnoreURLTags:  false,
		feeder:         feeder,
		scanbufPool:    &sync.Pool{},
		ScanBufferSize: 4 * 1024,
	}

	err = ipt.processLogBody(param)
	assert.Nil(t, err)

	pts, err := feeder.AnyPoints(time.Second * 2)
	if err != nil {
		t.Errorf("feeder err = %v", err)
		return
	}

	if len(pts) != 2 {
		t.Errorf("expected 2 points for firelens, got %d", len(pts))
	}

	for _, pt := range pts {
		t.Logf("firehose log : %s", pt.LineProto())
	}
}
