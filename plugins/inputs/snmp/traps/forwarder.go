// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package traps

import (
	"context"
	"net"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp/snmpmeasurement"
)

// TrapForwarder consumes from a trapsIn channel, format traps and send them as EventPlatformEvents
// The TrapForwarder is an intermediate step between the listener and the epforwarder in order to limit the processing of the listener
// to the minimum. The forwarder process payloads received by the listener via the trapsIn channel, formats them and finally
// give them to the epforwarder for sending it to Datadog.
type TrapForwarder struct {
	trapsIn   PacketsChannel
	formatter Formatter
	stopChan  chan struct{}
	Election  bool
}

// NewTrapForwarder creates a simple TrapForwarder instance.
func NewTrapForwarder(formatter Formatter, packets PacketsChannel, election bool) (*TrapForwarder, error) {
	return &TrapForwarder{
		trapsIn:   packets,
		formatter: formatter,
		stopChan:  make(chan struct{}),
		Election:  election,
	}, nil
}

// Start the TrapForwarder instance. Need to Stop it manually.
func (tf *TrapForwarder) Start() {
	l.Info("Starting TrapForwarder")
	g.Go(func(ctx context.Context) error {
		tf.run()
		return nil
	}) // Go
}

// Stop the TrapForwarder instance.
func (tf *TrapForwarder) Stop() {
	tf.stopChan <- struct{}{}
}

func (tf *TrapForwarder) run() {
	for {
		select {
		case <-tf.stopChan:
			l.Info("Stopped TrapForwarder")
			return
		case packet := <-tf.trapsIn:
			tf.sendTrap(packet)
		}
	}
}

func (tf *TrapForwarder) sendTrap(packet *SnmpPacket) {
	data, err := tf.formatter.FormatPacket(packet)
	if err != nil {
		l.Errorf("failed to format packet: %v", err)
		return
	}
	payload := string(data)
	l.Debugf("send trap payload: %s", payload)

	host, _, err := net.SplitHostPort(packet.Addr.String())
	if err != nil {
		l.Errorf("net.SplitHostPort failed: %v", err)
		return
	}
	if len(host) == 0 {
		l.Warn("host is empty")
		return
	}

	tags := map[string]string{
		"host": host,
	}
	Fields := map[string]interface{}{
		"trap_payload": payload,
	}
	tn := time.Now().UTC()
	var measurements []inputs.Measurement
	measurements = append(measurements, &snmpmeasurement.SNMPObject{
		Name:     "traps",
		Tags:     tags,
		Fields:   Fields,
		TS:       tn,
		Election: tf.Election,
	})

	if err := inputs.FeedMeasurement("traps-object",
		datakit.Object,
		measurements,
		&io.Option{CollectCost: time.Since(tn)}); err != nil {
		l.Errorf("FeedMeasurement object err: %v", err)
	}
}
