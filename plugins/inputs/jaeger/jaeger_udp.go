// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jaeger

import (
	"context"
	"net"
	"time"

	"github.com/uber/jaeger-client-go/thrift"
	"github.com/uber/jaeger-client-go/thrift-gen/agent"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

func StartUDPAgent(addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	log.Debugf("%s(UDP): listen on path: %s", inputName, addr)

	// receiving loop
	buf := make([]byte, 65535)
	for {
		select {
		case <-datakit.Exit.Wait():
			if err := udpConn.Close(); err != nil {
				log.Warnf("Close: %s", err)
			}
			log.Infof("jaeger udp agent closed")

			return nil
		default:
		}

		err := udpConn.SetDeadline(time.Now().Add(3 * time.Second))
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

		param := &itrace.TraceParameters{
			Meta: &itrace.TraceMeta{
				Protocol: "udp",
				Buf:      buf[:n],
			},
		}

		if storage == nil {
			if err = parseJaegerTraceUDP(param); err != nil {
				log.Error(err.Error())
			}
		} else {
			if err = storage.Send(param); err != nil {
				log.Error(err.Error())
			}
		}
	}
}

func parseJaegerTraceUDP(param *itrace.TraceParameters) error {
	thriftBuffer := thrift.NewTMemoryBufferLen(len(param.Meta.Buf))
	_, err := thriftBuffer.Write(param.Meta.Buf)
	if err != nil {
		return err
	}

	var (
		protocolFactory = thrift.NewTCompactProtocolFactoryConf(&thrift.TConfiguration{})
		thriftProtocol  = protocolFactory.GetProtocol(thriftBuffer)
	)
	if _, _, _, err = thriftProtocol.ReadMessageBegin(context.TODO()); err != nil { //nolint:dogsled
		return err
	}

	batch := agent.AgentEmitBatchArgs{}
	if err = batch.Read(context.TODO(), thriftProtocol); err != nil {
		return err
	}

	if err = thriftProtocol.ReadMessageEnd(context.TODO()); err != nil {
		log.Error("read message end failed :%s,", err.Error())
	}

	if dktrace := batchToDkTrace(batch.Batch); len(dktrace) != 0 && afterGatherRun != nil {
		afterGatherRun.Run(inputName, itrace.DatakitTraces{dktrace}, false)
	}

	return nil
}
