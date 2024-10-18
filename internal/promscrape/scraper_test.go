// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promscrape

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
)

const (
	mockHeader = `
# HELP datakit_http_worker_number The number of the worker
# TYPE datakit_http_worker_number gauge
`
	mockBody = `
datakit_http_worker_number{category="metric",domain="dataway.testing.com",status="%d", } 11.0
datakit_http_worker_number{category="metric",domain="dataway.testing.com",status="%d", } 12.2
datakit_http_worker_number{category="metric",domain="dataway.testing.com",status="%d", } 13.0
datakit_http_worker_number{category="metric",domain="dataway.testing.com",status="%d", } 14.2
datakit_http_worker_number{category="metric",domain="dataway.testing.com",status="%d", } 15.0
`
)

func TestParseStream(t *testing.T) {
	count := 0
	run := func() {
		var buf bytes.Buffer
		buf.WriteString(mockHeader)
		for i := 0; i < 10000; i++ {
			buf.WriteString(fmt.Sprintf(mockBody, i, i, i, i, i))
		}
		p := &PromScraper{
			opt: &option{
				measurement: "testing-meas",
				extraTags:   map[string]string{"key-01": "value-01"},
				callback: func(pts []*point.Point) error {
					count += len(pts)
					return nil
				},
			},
		}
		err := p.ParserStream(&buf)
		assert.NoError(t, err)
	}

	run()
	t.Logf("count: %d\n", count)
}

func BenchmarkParseStream(b *testing.B) {
	var buf bytes.Buffer
	buf.WriteString(mockHeader)
	for i := 0; i < 10000; i++ {
		buf.WriteString(fmt.Sprintf(mockBody, i, i, i, i, i))
	}

	p := &PromScraper{
		opt: &option{
			measurement: "testing-meas",
			extraTags:   map[string]string{"key-01": "value-01"},
			callback: func(pts []*point.Point) error {
				return nil
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := p.ParserStream(&buf)
		assert.NoError(b, err)
	}
}
