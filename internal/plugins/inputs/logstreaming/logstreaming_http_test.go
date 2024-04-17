// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package logstreaming testing.
package logstreaming

import (
	"bufio"
	"bytes"
	"net/url"
	"sync"
	"testing"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
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

func TestInput_processLogBody(t *testing.T) {
	// influxdb
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

	// others
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
					body:          buf,
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
					body: obuf,
				},
			},
			wantErr: false,
			checks:  otherCheckers,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := &Input{
				feeder:      tt.fields.feeder,
				scanbufPool: &sync.Pool{},
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
