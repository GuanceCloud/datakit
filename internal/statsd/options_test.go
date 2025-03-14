package statsd

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/stretchr/testify/assert"
)

func TestWithProtocol(t *testing.T) {
	opt := &option{}
	WithProtocol("udp")(opt)
	assert.Equal(t, "udp", opt.protocol)
}

func TestWithServiceAddress(t *testing.T) {
	opt := &option{}
	WithServiceAddress("localhost:8125")(opt)
	assert.Equal(t, "localhost:8125", opt.serviceAddress)
}

func TestWithStatsdSourceKey(t *testing.T) {
	opt := &option{}
	WithStatsdSourceKey("source")(opt)
	assert.Equal(t, "source", opt.statsdSourceKey)
}

func TestWithStatsdHostKey(t *testing.T) {
	opt := &option{}
	WithStatsdHostKey("host")(opt)
	assert.Equal(t, "host", opt.statsdHostKey)
}

func TestWithSaveAboveKey(t *testing.T) {
	opt := &option{}
	WithSaveAboveKey(true)(opt)
	assert.True(t, opt.saveAboveKey)
}

func TestWithAllowedPendingMessages(t *testing.T) {
	opt := &option{}
	WithAllowedPendingMessages(100)(opt)
	assert.Equal(t, 100, opt.allowedPendingMessages)
}

func TestWithPercentiles(t *testing.T) {
	opt := &option{}
	percentiles := []float64{90.0, 95.0, 99.0}
	WithPercentiles(percentiles)(opt)
	assert.Equal(t, percentiles, opt.percentiles)
}

func TestWithPercentileLimit(t *testing.T) {
	opt := &option{}
	WithPercentileLimit(1000)(opt)
	assert.Equal(t, 1000, opt.percentileLimit)
}

func TestWithDeleteGauges(t *testing.T) {
	opt := &option{}
	WithDeleteGauges(true)(opt)
	assert.True(t, opt.deleteGauges)
}

func TestWithDeleteCounters(t *testing.T) {
	opt := &option{}
	WithDeleteCounters(true)(opt)
	assert.True(t, opt.deleteCounters)
}

func TestWithSetCounterInt(t *testing.T) {
	opt := &option{}
	WithSetCounterInt(true)(opt)
	assert.True(t, opt.setCounterInt)
}

func TestWithDeleteSets(t *testing.T) {
	opt := &option{}
	WithDeleteSets(true)(opt)
	assert.True(t, opt.deleteSets)
}

func TestWithDeleteTimings(t *testing.T) {
	opt := &option{}
	WithDeleteTimings(true)(opt)
	assert.True(t, opt.deleteTimings)
}

func TestWithConvertNames(t *testing.T) {
	opt := &option{}
	WithConvertNames(true)(opt)
	assert.True(t, opt.convertNames)
}

func TestWithMetricSeparator(t *testing.T) {
	opt := &option{}
	WithMetricSeparator(".")(opt)
	assert.Equal(t, ".", opt.metricSeparator)
}

func TestWithDataDogExtensions(t *testing.T) {
	opt := &option{}
	WithDataDogExtensions(true)(opt)
	assert.True(t, opt.dataDogExtensions)
}

func TestWithDataDogDistributions(t *testing.T) {
	opt := &option{}
	WithDataDogDistributions(true)(opt)
	assert.True(t, opt.dataDogDistributions)
}

func TestWithUDPPacketSize(t *testing.T) {
	opt := &option{}
	WithUDPPacketSize(1500)(opt)
	assert.Equal(t, 1500, opt.udpPacketSize)
}

func TestWithReadBufferSize(t *testing.T) {
	opt := &option{}
	WithReadBufferSize(65535)(opt)
	assert.Equal(t, 65535, opt.readBufferSize)
}

func TestWithDropTags(t *testing.T) {
	opt := &option{}
	tags := []string{"tag1", "tag2"}
	WithDropTags(tags)(opt)
	assert.Equal(t, tags, opt.dropTags)
}

func TestWithMetricMapping(t *testing.T) {
	opt := &option{}
	mappings := []string{"mapping1", "mapping2"}
	WithMetricMapping(mappings)(opt)
	assert.Equal(t, mappings, opt.metricMapping)
}

func TestWithTags(t *testing.T) {
	opt := &option{}
	tags := map[string]string{"key1": "value1", "key2": "value2"}
	WithTags(tags)(opt)
	assert.Equal(t, tags, opt.tags)
}

func TestWithMaxTCPConnections(t *testing.T) {
	opt := &option{}
	WithMaxTCPConnections(250)(opt)
	assert.Equal(t, 250, opt.maxTCPConnections)
}

func TestWithTCPKeepAlive(t *testing.T) {
	opt := &option{}
	WithTCPKeepAlive(true)(opt)
	assert.True(t, opt.tcpKeepAlive)
}

func TestWithMaxTTL(t *testing.T) {
	opt := &option{}
	ttl := 30 * time.Second
	WithMaxTTL(ttl)(opt)
	assert.Equal(t, ttl, opt.maxTTL)
}

func TestWithLogger(t *testing.T) {
	opt := &option{}
	l := logger.DefaultSLogger("test")
	WithLogger(l)(opt)
	assert.Equal(t, l, opt.l)
}
