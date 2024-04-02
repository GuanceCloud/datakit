// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ploffload

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

func BenchmarkFmt(b *testing.B) {
	a := 123

	b.Run("strconv123", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			strconv.FormatInt(int64(a), 10)
		}
	})

	b.Run("sprintf123", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = fmt.Sprintf("%d", a)
		}
	})

	a = 123456789

	b.Run("strconv123456789", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			strconv.FormatInt(int64(a), 10)
		}
	})

	b.Run("sprintf123456789", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = fmt.Sprintf("%d", a)
		}
	})
}

func TestHTTPHandle(t *testing.T) {
	feeder := io.NewMockedFeeder()
	ipt := &Input{
		semStop: cliutils.NewSem(),
		feeder:  feeder,
	}

	testSvc := httptest.NewServer(http.HandlerFunc(ipt.handlePlOffload))

	defer testSvc.Close()

	batchSize := 10
	enc := point.GetEncoder(point.WithEncEncoding(point.Protobuf))

	defer point.PutEncoder(enc)

	cases := []struct {
		name  string
		count int
	}{
		{
			"test1",
			batchSize + 1,
		},
	}

	for _, ca := range cases {
		t.Run(ca.name, func(t *testing.T) {
			pts := make([]*point.Point, 0, ca.count)
			for i := 0; i < ca.count; i++ {
				kvs := point.NewKVs(map[string]interface{}{"a": 1})
				kvs.AddTag("b", "2")
				pt := point.NewPointV2("test", kvs)
				pts = append(pts, pt)
			}

			buf, err := enc.Encode(pts)
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", testSvc.URL+"/v1/write/ploffload/logging", bytes.NewBuffer(buf[0]))

			req.Header.Add("Content-Type", point.Protobuf.HTTPContentType())

			assert.NoError(t, err)

			resp, err := testSvc.Client().Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, 200, resp.StatusCode)
		})
	}
}
