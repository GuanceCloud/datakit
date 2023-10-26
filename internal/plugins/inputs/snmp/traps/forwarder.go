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

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmpmeasurement"
)

const trapsObject = "traps-object"

// TrapForwarder consumes from a trapsIn channel, format traps and send them as EventPlatformEvents
// The TrapForwarder is an intermediate step between the listener and the epforwarder in order to limit the processing of the listener
// to the minimum. The forwarder process payloads received by the listener via the trapsIn channel, formats them and finally
// give them to the epforwarder for sending it to Datadog.
type TrapForwarder struct {
	trapsIn   PacketsChannel
	formatter Formatter
	stopChan  chan struct{}
	election  bool
	inputTags map[string]string
	feeder    dkio.Feeder
	tagger    datakit.GlobalTagger
}

// NewTrapForwarder creates a simple TrapForwarder instance.
func NewTrapForwarder(formatter Formatter, packets PacketsChannel, opt *TrapsServerOpt) (*TrapForwarder, error) {
	trapForwarder := &TrapForwarder{
		trapsIn:   packets,
		formatter: formatter,
		stopChan:  make(chan struct{}),
		election:  opt.Election,
		inputTags: opt.InputTags,
		feeder:    opt.Feeder,
		tagger:    opt.Tagger,
	}

	return trapForwarder, nil
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
	fields := map[string]interface{}{
		"trap_payload": payload,
	}
	tn := time.Now()

	if tf.election {
		tags = inputs.MergeTagsWrapper(tags, tf.tagger.ElectionTags(), tf.inputTags, host)
	} else {
		tags = inputs.MergeTagsWrapper(tags, tf.tagger.HostTags(), tf.inputTags, host)
	}

	metric := &snmpmeasurement.SNMPObject{
		Name:   "traps",
		Tags:   tags,
		Fields: fields,
		TS:     tn,
	}

	if err := tf.feeder.Feed(trapsObject, point.Object,
		[]*point.Point{metric.Point()},
		&dkio.Option{CollectCost: time.Since(tn)}); err != nil {
		l.Errorf("Feed object err: %v", err)
		tf.feeder.FeedLastError(err.Error(),
			dkio.WithLastErrorInput(snmpmeasurement.InputName),
			dkio.WithLastErrorSource(trapsObject),
			dkio.WithLastErrorCategory(point.Object),
		)
	}
}
