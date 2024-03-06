// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package statsd

import (
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

func TestInput_Collect_BatchLength(t *testing.T) {
	tests := []struct {
		name     string
		ptLength int
		want     []int
	}{
		{
			name:     "0 pt",
			ptLength: 0,
			want:     []int{},
		},
		{
			name:     "1 pt",
			ptLength: 1,
			want:     []int{1},
		},
		{
			name:     "1024 pt",
			ptLength: 1024,
			want:     []int{1024},
		},
		{
			name:     "1025 pt",
			ptLength: 1025,
			want:     []int{1024, 1},
		},

		{
			name:     "2049 pt",
			ptLength: 2049,
			want:     []int{1024, 1024, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := &Input{}

			feeder := NewPTLenMockedFeeder()
			ipt.Feeder = feeder

			ipt.feedBatch(getPoints(tt.ptLength))

			assert.Equal(t, feeder.GetBatchLength(), tt.want)
		})
	}
}

// mock data

func getPoints(pointLength int) []*point.Point {
	metricName := "mockMetric"
	measurs := []*point.Point{}
	var opts []point.Option

	for i := 0; i < pointLength; i++ {
		fields := map[string]interface{}{"val": i}
		pt := point.NewPointV2(metricName,
			point.NewKVs(fields),
			opts...)

		measurs = append(measurs, pt)
	}

	return measurs
}

// ------ pt len mock feeder ------

type PTLenMockedFeeder struct {
	batchLength []int
}

func NewPTLenMockedFeeder() *PTLenMockedFeeder {
	return &PTLenMockedFeeder{}
}

func (m *PTLenMockedFeeder) Feed(name string, category point.Category, pts []*point.Point, opts ...*dkio.Option) error {
	m.batchLength = append(m.batchLength, len(pts))
	return nil
}

func (m *PTLenMockedFeeder) FeedV2(category point.Category, pts []*point.Point, opts ...dkio.FeedOption) error {
	m.batchLength = append(m.batchLength, len(pts))
	return nil
}

func (m *PTLenMockedFeeder) FeedLastError(err string, opts ...dkio.LastErrorOption) {}

func (m *PTLenMockedFeeder) GetBatchLength() []int {
	if m.batchLength == nil {
		return []int{}
	}
	return m.batchLength
}

func (m *PTLenMockedFeeder) Clear() {
	m.batchLength = make([]int, 0)
}
