// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

// Package flowaggregator contains flow aggregator.
package flowaggregator

import (
	"encoding/json"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/prometheus/client_golang/prometheus"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/common"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/metrics"
	"go.uber.org/atomic"
)

const (
	flushFlowsToSendInterval = 10 * time.Second
	hostTagKey               = "host"
	unknownHost              = "unknown"
	OK                       = "OK"
)

// FlowAggregator is used for space and time aggregation of NetFlow flows.
type FlowAggregator struct {
	flowIn                       chan *common.Flow
	flushFlowsToSendInterval     time.Duration // interval for checking flows to flush and send them to EP Forwarder
	rollupTrackerRefreshInterval time.Duration
	flowAcc                      *flowAccumulator
	stopChan                     chan struct{}
	flushLoopDone                chan struct{}
	runDone                      chan struct{}
	receivedFlowCount            *atomic.Uint64
	flushedFlowCount             *atomic.Uint64
	hostname                     string
	goflowPrometheusGatherer     prometheus.Gatherer
	TimeNowFunction              func() time.Time // Allows to mock time in tests

	lastSequencePerExporter map[SequenceDeltaKey]uint32
	combinedTags            map[string]string
	feeder                  dkio.Feeder
	source                  string
}

type SequenceDeltaKey struct {
	Namespace  string
	ExporterIP string
	FlowType   common.FlowType
}

type SequenceDeltaValue struct {
	Delta        int64
	LastSequence uint32
	Reset        bool
}

// NewFlowAggregator returns a new FlowAggregator.
//
//nolint:lll
func NewFlowAggregator(config *config.NetflowConfig, combinedTags map[string]string, feeder dkio.Feeder, source string) *FlowAggregator {
	flushInterval := time.Duration(config.AggregatorFlushInterval) * time.Second
	flowContextTTL := time.Duration(config.AggregatorFlowContextTTL) * time.Second
	rollupTrackerRefreshInterval := time.Duration(config.AggregatorRollupTrackerRefreshInterval) * time.Second

	hostname, ok := combinedTags[hostTagKey]
	if !ok {
		hostname = unknownHost
	}

	return &FlowAggregator{
		flowIn:                       make(chan *common.Flow, config.AggregatorBufferSize),
		flowAcc:                      newFlowAccumulator(flushInterval, flowContextTTL, config.AggregatorPortRollupThreshold, config.AggregatorPortRollupDisabled),
		flushFlowsToSendInterval:     flushFlowsToSendInterval,
		rollupTrackerRefreshInterval: rollupTrackerRefreshInterval,
		stopChan:                     make(chan struct{}),
		runDone:                      make(chan struct{}),
		flushLoopDone:                make(chan struct{}),
		receivedFlowCount:            atomic.NewUint64(0),
		flushedFlowCount:             atomic.NewUint64(0),
		hostname:                     hostname,
		goflowPrometheusGatherer:     prometheus.DefaultGatherer,
		TimeNowFunction:              time.Now,
		lastSequencePerExporter:      make(map[SequenceDeltaKey]uint32),

		combinedTags: combinedTags,
		feeder:       feeder,
		source:       source,
	}
}

// Start will start the FlowAggregator worker.
func (agg *FlowAggregator) Start() {
	l.Info("Flow Aggregator started")
	go agg.run()
	agg.flushLoop() // blocking call
}

// Stop will stop running FlowAggregator.
func (agg *FlowAggregator) Stop() {
	close(agg.stopChan)
	<-agg.flushLoopDone
	<-agg.runDone
}

// GetFlowInChan returns flow input chan.
func (agg *FlowAggregator) GetFlowInChan() chan *common.Flow {
	return agg.flowIn
}

func (agg *FlowAggregator) run() {
	for {
		select {
		case <-agg.stopChan:
			l.Info("Stopping aggregator")
			agg.runDone <- struct{}{}
			return
		case flow := <-agg.flowIn:
			agg.receivedFlowCount.Inc()
			agg.flowAcc.add(flow)
		}
	}
}

func (agg *FlowAggregator) sendFlows(flows []*common.Flow, flushTime time.Time) {
	for _, flow := range flows {
		flowPayload := buildPayload(flow, agg.hostname, flushTime)
		payloadBytes, err := json.Marshal(flowPayload)
		if err != nil {
			l.Errorf("Error marshaling device metadata: %s", err)
			continue
		}

		// l.Debugf("flushed flow: %s", string(payloadBytes))

		logging := &metrics.NetflowMeasurement{
			Name: agg.source,
			Tags: agg.combinedTags,
			TS:   flushTime,
		}
		if logging.Fields == nil {
			logging.Fields = make(map[string]interface{})
		}
		logging.Fields[pipeline.FieldMessage] = string(payloadBytes)
		logging.Fields[pipeline.FieldStatus] = OK
		logging.Fields["bytes"] = flowPayload.Bytes
		logging.Fields["dest_ip"] = flowPayload.Destination.IP
		logging.Fields["dest_port"] = flowPayload.Destination.Port
		logging.Fields["device_ip"] = flowPayload.Exporter.IP
		logging.Fields["ip_protocol"] = flowPayload.IPProtocol
		logging.Fields["source_ip"] = flowPayload.Source.IP
		logging.Fields["source_port"] = flowPayload.Source.Port
		logging.Fields["type"] = flowPayload.FlowType

		if err := agg.feeder.Feed(common.InputName+"/"+agg.source, point.Logging,
			[]*point.Point{logging.Point()}, &dkio.Option{CollectCost: time.Since(flushTime)},
		); err != nil {
			l.Errorf("Feed failed: %v", err)
		}
	}
}

func (agg *FlowAggregator) flushLoop() {
	var flushFlowsToSendTicker <-chan time.Time

	if agg.flushFlowsToSendInterval > 0 {
		flushFlowsToSendTicker = time.NewTicker(agg.flushFlowsToSendInterval).C
	} else {
		l.Debug("flushFlowsToSendInterval set to 0: will never flush automatically")
	}

	rollupTrackersRefresh := time.NewTicker(agg.rollupTrackerRefreshInterval).C
	// TODO: move rollup tracker refresh to a separate loop (separate PR) to avoid rollup tracker and flush flows impacting each other

	for {
		select {
		// stop sequence
		case <-agg.stopChan:
			agg.flushLoopDone <- struct{}{}
			return

		// automatic flush sequence
		case <-flushFlowsToSendTicker:
			agg.flush()

		// refresh rollup trackers
		case <-rollupTrackersRefresh:
			agg.rollupTrackersRefresh()
		}
	}
}

// Flush flushes the aggregator.
//
//nolint:lll
func (agg *FlowAggregator) flush() int {
	flowsContexts := agg.flowAcc.getFlowContextCount()
	flushTime := agg.TimeNowFunction()
	flowsToFlush := agg.flowAcc.flush()
	l.Debugf("Flushing %d flows to the forwarder (flush_duration=%d, flow_contexts_before_flush=%d)", len(flowsToFlush), time.Since(flushTime).Milliseconds(), flowsContexts)

	// TODO: Add flush stats to agent telemetry e.g. aggregator newFlushCountStats()
	if len(flowsToFlush) > 0 {
		agg.sendFlows(flowsToFlush, flushTime)
	}

	flushCount := len(flowsToFlush)

	// We increase `flushedFlowCount` at the end to be sure that the metrics are submitted before hand.
	// Tests will wait for `flushedFlowCount` to be increased before asserting the metrics.
	agg.flushedFlowCount.Add(uint64(flushCount))
	return len(flowsToFlush)
}

func (agg *FlowAggregator) rollupTrackersRefresh() {
	l.Debugf("Rollup tracker refresh: use new store as current store")
	agg.flowAcc.portRollup.UseNewStoreAsCurrentStore()
}

var l = logger.DefaultSLogger(common.InputName)

func SetLogger(log *logger.Logger) {
	l = log
}
