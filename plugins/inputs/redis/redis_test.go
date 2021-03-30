package redis

import (
	"testing"
	// "github.com/stretchr/testify/assert"
)

func TestCollectInfoMetrics(t *testing.T) {
	r := &Redis{}
	r.MetricName= "test"

	r.Init()

	r.collectInfoMetrics()

	r.submit()
}