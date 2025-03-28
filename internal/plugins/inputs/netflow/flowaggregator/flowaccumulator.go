// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package flowaggregator

import (
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/common"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/portrollup"
	"go.uber.org/atomic"
)

var timeNow = time.Now

// flowContext contains flow information and additional flush related data.
type flowContext struct {
	flow                *common.Flow
	nextFlush           time.Time
	lastSuccessfulFlush time.Time
}

// flowAccumulator is used to accumulate aggregated flows.
type flowAccumulator struct {
	flows map[uint64]flowContext
	// mutex is needed to protect `flows` since `flowAccumulator.add()` and  `flowAccumulator.flush()`
	// are called by different routines.
	flowsMutex sync.Mutex

	flowFlushInterval time.Duration
	flowContextTTL    time.Duration

	portRollup          *portrollup.EndpointPairPortRollupStore
	portRollupThreshold int
	portRollupDisabled  bool

	hashCollisionFlowCount *atomic.Uint64
}

func newFlowContext(flow *common.Flow) flowContext {
	now := timeNow()
	return flowContext{
		flow:      flow,
		nextFlush: now,
	}
}

//nolint:lll
func newFlowAccumulator(aggregatorFlushInterval time.Duration, aggregatorFlowContextTTL time.Duration, portRollupThreshold int, portRollupDisabled bool) *flowAccumulator {
	return &flowAccumulator{
		flows:                  make(map[uint64]flowContext),
		flowFlushInterval:      aggregatorFlushInterval,
		flowContextTTL:         aggregatorFlowContextTTL,
		portRollup:             portrollup.NewEndpointPairPortRollupStore(portRollupThreshold),
		portRollupThreshold:    portRollupThreshold,
		portRollupDisabled:     portRollupDisabled,
		hashCollisionFlowCount: atomic.NewUint64(0),
	}
}

// flush will flush specific flow context (distinct hash) if nextFlush is reached
// once a flow context is flushed nextFlush will be updated to the next flush time
//
// flowContextTTL:
// flowContextTTL defines the duration we should keep a specific flowContext in `flowAccumulator.flows`
// after `lastSuccessfulFlush`. // Flow context in `flowAccumulator.flows` map will be deleted if `flowContextTTL`
// is reached to avoid keeping flow context that are not seen anymore.
// We need to keep flowContext (contains `nextFlush` and `lastSuccessfulFlush`) after flush
// to be able to flush at regular interval (`flowFlushInterval`).
// Example, after a flush, flowContext will have a new nextFlush, that will be the next flush time for new flows being added.
//
//nolint:lll
func (f *flowAccumulator) flush() []*common.Flow {
	f.flowsMutex.Lock()
	defer f.flowsMutex.Unlock()

	var flowsToFlush []*common.Flow
	for key, flowCtx := range f.flows {
		now := timeNow()
		if flowCtx.flow == nil && (flowCtx.lastSuccessfulFlush.Add(f.flowContextTTL).Before(now)) {
			l.Debugf("Delete flow context (key=%d, lastSuccessfulFlush=%s, nextFlush=%s)", key, flowCtx.lastSuccessfulFlush.String(), flowCtx.nextFlush.String())
			// delete flowCtx wrapper if there is no successful flushes since `flowContextTTL`
			delete(f.flows, key)
			continue
		}
		if flowCtx.nextFlush.After(now) {
			continue
		}
		if flowCtx.flow != nil {
			flowsToFlush = append(flowsToFlush, flowCtx.flow)
			flowCtx.lastSuccessfulFlush = now
			flowCtx.flow = nil
		}
		flowCtx.nextFlush = flowCtx.nextFlush.Add(f.flowFlushInterval)
		f.flows[key] = flowCtx
	}
	return flowsToFlush
}

func (f *flowAccumulator) add(flowToAdd *common.Flow) {
	l.Debugf("Add new flow: %+v", flowToAdd)

	if !f.portRollupDisabled {
		// Handle port rollup
		f.portRollup.Add(flowToAdd.SrcAddr, flowToAdd.DstAddr, uint16(flowToAdd.SrcPort), uint16(flowToAdd.DstPort))
		ephemeralStatus := f.portRollup.IsEphemeral(flowToAdd.SrcAddr, flowToAdd.DstAddr, uint16(flowToAdd.SrcPort), uint16(flowToAdd.DstPort))
		//nolint:exhaustive
		switch ephemeralStatus {
		case portrollup.IsEphemeralSourcePort:
			flowToAdd.SrcPort = portrollup.EphemeralPort
		case portrollup.IsEphemeralDestPort:
			flowToAdd.DstPort = portrollup.EphemeralPort
		}
	}

	f.flowsMutex.Lock()
	defer f.flowsMutex.Unlock()

	aggHash := flowToAdd.AggregationHash()
	aggFlow, ok := f.flows[aggHash]
	if !ok {
		f.flows[aggHash] = newFlowContext(flowToAdd)
		return
	}
	if aggFlow.flow == nil {
		aggFlow.flow = flowToAdd
	} else {
		// use go routine for has collision detection to avoid blocking critical path
		go f.detectHashCollision(aggHash, *aggFlow.flow, *flowToAdd)

		// accumulate flowToAdd with existing flow(s) with same hash
		aggFlow.flow.Bytes += flowToAdd.Bytes
		aggFlow.flow.Packets += flowToAdd.Packets
		aggFlow.flow.StartTimestamp = common.MinUint64(aggFlow.flow.StartTimestamp, flowToAdd.StartTimestamp)
		aggFlow.flow.EndTimestamp = common.MaxUint64(aggFlow.flow.EndTimestamp, flowToAdd.EndTimestamp)
		aggFlow.flow.SequenceNum = common.MaxUint32(aggFlow.flow.SequenceNum, flowToAdd.SequenceNum)
		aggFlow.flow.TCPFlags |= flowToAdd.TCPFlags
	}
	f.flows[aggHash] = aggFlow
}

func (f *flowAccumulator) getFlowContextCount() int {
	f.flowsMutex.Lock()
	defer f.flowsMutex.Unlock()

	return len(f.flows)
}

func (f *flowAccumulator) detectHashCollision(hash uint64, existingFlow common.Flow, flowToAdd common.Flow) {
	if !common.IsEqualFlowContext(existingFlow, flowToAdd) {
		l.Warnf("Hash collision for flows with hash `%d`: existingFlow=`%+v` flowToAdd=`%+v`", hash, existingFlow, flowToAdd)
		f.hashCollisionFlowCount.Inc()
	}
}
