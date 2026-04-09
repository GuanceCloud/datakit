package aggregate

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

type TailSamplingBuiltinMetric interface {
	Name() string
	OnIngest(packet *DataPacket) []DerivedMetricRecord
	OnPreDecision(packet *DataPacket) []DerivedMetricRecord
	OnDecision(packet *DataPacket, decision DerivedMetricDecision) []DerivedMetricRecord
}

type TailSamplingBuiltinMetrics []TailSamplingBuiltinMetric

func (ms TailSamplingBuiltinMetrics) OnIngest(packet *DataPacket) []DerivedMetricRecord {
	var records []DerivedMetricRecord
	for _, metric := range ms {
		records = append(records, metric.OnIngest(packet)...)
	}
	return records
}

func (ms TailSamplingBuiltinMetrics) OnDecision(packet *DataPacket, decision DerivedMetricDecision) []DerivedMetricRecord {
	var records []DerivedMetricRecord
	for _, metric := range ms {
		records = append(records, metric.OnDecision(packet, decision)...)
	}
	return records
}

func (ms TailSamplingBuiltinMetrics) OnPreDecision(packet *DataPacket) []DerivedMetricRecord {
	var records []DerivedMetricRecord
	for _, metric := range ms {
		records = append(records, metric.OnPreDecision(packet)...)
	}
	return records
}

type TailSamplingProcessor struct {
	sampler   *GlobalSampler
	collector *DerivedMetricCollector
	metrics   TailSamplingBuiltinMetrics
}

func NewDefaultTailSamplingProcessor(shardCount int, waitTime time.Duration) *TailSamplingProcessor {
	return &TailSamplingProcessor{
		sampler:   NewGlobalSampler(shardCount, waitTime),
		collector: NewDerivedMetricCollector(DefaultDerivedMetricFlushWindow),
		metrics:   DefaultTailSamplingBuiltinMetrics(),
	}
}

func NewTailSamplingProcessor(
	sampler *GlobalSampler,
	collector *DerivedMetricCollector,
	metrics TailSamplingBuiltinMetrics,
) *TailSamplingProcessor {
	return &TailSamplingProcessor{
		sampler:   sampler,
		collector: collector,
		metrics:   metrics,
	}
}

func (r *TailSamplingProcessor) Sampler() *GlobalSampler {
	if r == nil {
		return nil
	}
	return r.sampler
}

func (r *TailSamplingProcessor) Collector() *DerivedMetricCollector {
	if r == nil {
		return nil
	}
	return r.collector
}

func (r *TailSamplingProcessor) BuiltinMetrics() TailSamplingBuiltinMetrics {
	if r == nil {
		return nil
	}
	return r.metrics
}

func (r *TailSamplingProcessor) UpdateConfig(token string, cfg *TailSamplingConfigs) error {
	if r == nil || r.sampler == nil || cfg == nil {
		return nil
	}

	return r.sampler.UpdateConfig(token, cfg)
}

func (r *TailSamplingProcessor) IngestPacket(packet *DataPacket) {
	if r == nil || packet == nil {
		return
	}

	if r.collector != nil && len(r.metrics) > 0 {
		r.collector.Add(r.filterBuiltinRecords(packet, r.metrics.OnIngest(packet)))
	}

	if r.sampler != nil {
		r.sampler.Ingest(packet)
	}
}

func (r *TailSamplingProcessor) AdvanceTime() map[uint64]*DataGroup {
	if r == nil || r.sampler == nil {
		return nil
	}

	return r.sampler.AdvanceTime()
}

func (r *TailSamplingProcessor) TailSamplingData(dataGroups map[uint64]*DataGroup) map[uint64]*DataPacket {
	if r == nil || r.sampler == nil {
		return nil
	}

	if r.collector != nil && len(r.metrics) > 0 {
		for _, dg := range dataGroups {
			if dg == nil || dg.packet == nil {
				continue
			}
			r.collector.Add(r.filterBuiltinRecords(dg.packet, r.metrics.OnPreDecision(dg.packet)))
		}
	}

	outcomes := r.sampler.TailSamplingOutcomes(dataGroups)
	keptPackets := make(map[uint64]*DataPacket)

	if r.collector == nil || len(r.metrics) == 0 {
		for key, outcome := range outcomes {
			if outcome == nil || outcome.Packet == nil {
				continue
			}
			keptPackets[key] = outcome.Packet
		}
		return keptPackets
	}

	for key, outcome := range outcomes {
		if outcome == nil {
			continue
		}

		packetForMetrics := outcome.SourcePacket
		if packetForMetrics != nil {
			r.collector.Add(r.filterBuiltinRecords(packetForMetrics, r.metrics.OnDecision(packetForMetrics, outcome.Decision)))
		}

		if outcome.Packet != nil {
			keptPackets[key] = outcome.Packet
		}
	}

	return keptPackets
}

func (r *TailSamplingProcessor) RecordDecision(packet *DataPacket, decision DerivedMetricDecision) {
	if r == nil || packet == nil || r.collector == nil || len(r.metrics) == 0 {
		return
	}

	r.collector.Add(r.filterBuiltinRecords(packet, r.metrics.OnDecision(packet, decision)))
}

func (r *TailSamplingProcessor) FlushDerivedMetrics(now time.Time) []*DerivedMetricPoints {
	if r == nil || r.collector == nil {
		return nil
	}

	return r.collector.Flush(now)
}

func (r *TailSamplingProcessor) filterBuiltinRecords(packet *DataPacket, records []DerivedMetricRecord) []DerivedMetricRecord {
	if len(records) == 0 || packet == nil {
		return records
	}

	filtered := make([]DerivedMetricRecord, 0, len(records))
	for _, record := range records {
		if r.isBuiltinMetricEnabled(packet.Token, packet.DataType, record.MetricName) {
			filtered = append(filtered, record)
		}
	}

	return filtered
}

func (r *TailSamplingProcessor) isBuiltinMetricEnabled(token, dataType, metricName string) bool {
	if r == nil || r.sampler == nil {
		return true
	}

	var cfgs []*BuiltinMetricCfg

	switch dataType {
	case point.STracing:
		traceCfg := r.sampler.GetTraceConfig(token)
		if traceCfg == nil {
			return true
		}
		cfgs = traceCfg.BuiltinMetrics
	case point.SLogging:
		loggingCfg := r.sampler.GetLoggingConfig(token)
		if loggingCfg == nil {
			return true
		}
		cfgs = loggingCfg.BuiltinMetrics
	case point.SRUM:
		rumCfg := r.sampler.GetRUMConfig(token)
		if rumCfg == nil {
			return true
		}
		cfgs = rumCfg.BuiltinMetrics
	default:
		return true
	}

	if len(cfgs) == 0 {
		return true
	}

	for _, cfg := range cfgs {
		if cfg == nil {
			continue
		}
		if cfg.Name == metricName {
			return cfg.Enabled
		}
	}

	return true
}
