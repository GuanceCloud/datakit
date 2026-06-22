package trace

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func Test_metricsSetup(t *testing.T) {
	// Reset any existing metrics
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	// Call metricsSetup
	metricsSetup()

	// Verify TracingProcessCount metric was created with correct parameters
	assert.NotNil(t, TracingProcessCount)
	assert.Equal(t, "datakit", TracingProcessCount.Desc().Namespace)
	assert.Equal(t, "input", TracingProcessCount.Desc().Subsystem)
	assert.Equal(t, "tracing_total", TracingProcessCount.Desc().Name)
	assert.Equal(t, "The total links number of Trace processed by the trace module", TracingProcessCount.Desc().Help)
	assert.Equal(t, []string{"input", "service"}, TracingProcessCount.Desc().VariableLabels)

	// Verify tracingSamplerCount metric was created with correct parameters
	assert.NotNil(t, tracingSamplerCount)
	assert.Equal(t, "datakit", tracingSamplerCount.Desc().Namespace)
	assert.Equal(t, "input", tracingSamplerCount.Desc().Subsystem)
	assert.Equal(t, "sampler_total", tracingSamplerCount.Desc().Name)
	assert.Equal(t, "The sampler number of Trace processed by the trace module", tracingSamplerCount.Desc().Help)
	assert.Equal(t, []string{"input", "service"}, tracingSamplerCount.Desc().VariableLabels)
}
