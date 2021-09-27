package traceJaeger

import (
	"context"
	"net"
	"time"

	"github.com/uber/jaeger-client-go/thrift"
	"github.com/uber/jaeger-client-go/thrift-gen/agent"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

func StartUdpAgent(addr string) error {
	data := make([]byte, 65535)

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	log.Infof("jaeger udp agent %v start", addr)

	// 循环读取消息
	for {
		select {
		case <-datakit.Exit.Wait():
			udpConn.Close()
			log.Infof("jaeger udp agent closed")
			return nil

		default:
		}

		err := udpConn.SetDeadline(time.Now().Add(time.Second))
		if err != nil {
			log.Errorf("SetDeadline failed: %v", err)
			continue
		}

		n, addr, err := udpConn.ReadFromUDP(data[:])
		if err != nil {
			continue
		} else {
			log.Debugf("Read from udp server:%s %d bytes", addr, n)
		}

		if n <= 0 {
			continue
		}

		groups, err := parseJaegerUdp(data[:n])
		if err != nil {
			continue
		}
		if len(groups) != 0 {
			trace.MkLineProto(groups, inputName)
		} else {
			log.Debug("empty batch")
		}
	}
}

func parseJaegerUdp(data []byte) ([]*trace.TraceAdapter, error) {
	thriftBuffer := thrift.NewTMemoryBufferLen(len(data))
	if _, err := thriftBuffer.Write(data[:]); err != nil {
		log.Error("buffer write failed :%v,", err)

		return nil, err
	}

	protocolFactory := thrift.NewTCompactProtocolFactoryConf(&thrift.TConfiguration{})
	thriftProtocol := protocolFactory.GetProtocol(thriftBuffer)
	_, _, _, err := thriftProtocol.ReadMessageBegin(context.TODO()) //nolint:dogsled
	if err != nil {
		log.Error("read message begin failed :%v,", err)

		return nil, err
	}

	batch := agent.AgentEmitBatchArgs{}
	err = batch.Read(context.TODO(), thriftProtocol)
	if err != nil {
		log.Error("read batch failed :%v,", err)

		return nil, err
	}

	var filters []batchFilter
	if len(sampleConfs) != 0 {
		filters = append(filters, sample)
	}

	groups, err := batchToAdapters(batch.Batch, filters...)
	if err != nil {
		log.Error("process batch failed :%v,", err)

		return nil, err
	}

	err = thriftProtocol.ReadMessageEnd(context.TODO())
	if err != nil {
		log.Error("read message end failed :%v,", err)

		return nil, err
	}

	return groups, nil
}
