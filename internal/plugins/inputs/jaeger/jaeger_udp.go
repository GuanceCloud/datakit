// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jaeger

import (
	"bytes"
	"context"
	"net"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/uber/jaeger-client-go/thrift"
	"github.com/uber/jaeger-client-go/thrift-gen/agent"
	"github.com/uber/jaeger-client-go/utils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

func StartUDPAgent(udpConn *net.UDPConn, addr string, semStop *cliutils.Sem) error {
	if udpConn == nil {
		udpAddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			return err
		}

		udpConn, err = net.ListenUDP("udp", udpAddr)
		if err != nil {
			return err
		}
	}

	log.Debugf("%s(UDP): listen on path: %s", inputName, addr)

	// receiving loop
	buf := make([]byte, utils.UDPPacketMaxLength)
	for {
		select {
		case <-datakit.Exit.Wait():
			udpExit(udpConn)
			log.Infof("jaeger udp agent exited")
			return nil

		case <-semStop.Wait():
			udpExit(udpConn)
			log.Infof("jaeger udp agent returned")
			return nil

		default:
		}

		err := udpConn.SetDeadline(time.Now().Add(10 * time.Second))
		if err != nil {
			log.Error(err.Error())
			continue
		}

		n, addr, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			log.Debug(err.Error())
			continue
		}
		if n <= 0 {
			continue
		}
		log.Debugf("### read from udp server:%s %d bytes", addr, n)

		param := &itrace.TraceParameters{Body: bytes.NewBuffer(buf[:n])}
		if err = parseJaegerTraceUDP(param); err != nil {
			log.Errorf("### parse jaeger trace from UDP failed: %s", err.Error())
		}
	}
}

func udpExit(udpConn *net.UDPConn) {
	if err := udpConn.Close(); err != nil {
		log.Warnf("UDP Close failed: %v", err)
	}
}

func parseJaegerTraceUDP(param *itrace.TraceParameters) error {
	tmbuf := thrift.NewTMemoryBufferLen(param.Body.Len())
	_, err := tmbuf.Write(param.Body.Bytes())
	if err != nil {
		return err
	}

	var (
		tprot = thrift.NewTCompactProtocolFactoryConf(&thrift.TConfiguration{}).GetProtocol(tmbuf)
		ctx   = context.Background()
	)
	if _, _, _, err = tprot.ReadMessageBegin(ctx); err != nil { //nolint:dogsled
		return err
	}
	defer func() {
		if err := tprot.ReadMessageEnd(ctx); err != nil {
			log.Error("read message end failed :%s,", err.Error())
		}
	}()

	batch := &agent.AgentEmitBatchArgs{}
	if err = batch.Read(ctx, tprot); err != nil {
		return err
	}

	if dktrace := batchToDkTrace(batch.Batch); len(dktrace) != 0 && afterGatherRun != nil {
		afterGatherRun.Run(inputName, itrace.DatakitTraces{dktrace})
	}

	return nil
}
