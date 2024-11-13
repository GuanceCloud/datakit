package exporter

import (
	"context"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
)

func TestSampling(t *testing.T) {
	Init(context.Background())

	s := newSampling(context.Background(), &cfg{
		samplingRate: "0.1",
	})

	var pts []*point.Point
	for i := 1; i <= 199; i++ {
		pt, _ := point.NewPoint("a", nil, map[string]any{"id": i})
		pts = append(pts, pt)
	}

	newpts := s.sampling(point.Network.String(), pts)
	assert.Equal(t, 20, len(newpts))
}

func TestSamplingPtsLess1(t *testing.T) {
	Init(context.Background())

	ePtsVec.WithLabelValues("x", point.Network.String()).Add(1000)
	r, rmp := parseRate(
		"0.01", 1000)
	t.Logf("set samping rate %.2f pts/min", r)
	s := &sampling{
		rate:         map[string]float64{},
		ptsPerMinute: rmp,
	}
	lastStats, _ := getTotalPtsMetric()

	ePtsVec.WithLabelValues("y", point.Network.String()).Add(2000)

	_ = s.changeRate(lastStats)

	var pts []*point.Point
	for i := 1; i <= 11; i++ {
		pt, _ := point.NewPoint("a", nil, map[string]any{"id": i})
		pts = append(pts, pt)
	}

	newpts := s.sampling(point.Network.String(), pts)
	assert.Equal(t, 1, len(newpts))
}
