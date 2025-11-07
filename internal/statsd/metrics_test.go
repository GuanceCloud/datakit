package statsd

import (
	"testing"

	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/stretchr/testify/assert"
)

func TestMetricsSetup(t *testing.T) {
	// Unregister existing metrics first
	metrics.Unregister(collectPointsTotalVec)
	metrics.Unregister(httpGetBytesVec)

	// Reset metric vectors
	collectPointsTotalVec = nil
	httpGetBytesVec = nil

	// Call metricsSetup
	metricsSetup()

	// Verify metrics were initialized
	assert.NotNil(t, collectPointsTotalVec, "collectPointsTotalVec should be initialized")
	assert.NotNil(t, httpGetBytesVec, "httpGetBytesVec should be initialized")

	// Test that metrics can accept values without error
	err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = r.(error)
			}
		}()
		collectPointsTotalVec.WithLabelValues().Observe(1.0)
		httpGetBytesVec.WithLabelValues().Observe(100.0)
		return nil
	}()
	assert.NoError(t, err, "metrics should accept values without error")
}
