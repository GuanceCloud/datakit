package traceJaeger

import (
	"net"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"

	"github.com/uber/jaeger-client-go/thrift"
	"github.com/uber/jaeger-client-go/thrift-gen/agent"
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
	defer udpConn.Close()

	// 循环读取消息
	for {
		select {
		case <-datakit.Exit.Wait():
			log.Infof("jaeger udp agent closed")
			return nil
		default:

		}
		n, addr, err := udpConn.ReadFromUDP(data[:])
		if err != nil {
			log.Error(err)
			continue
		} else {
			log.Infof("Read from udp server:%s %d bytes", addr, n)
		}

		if n <= 0 {
			continue
		}

		thriftBuffer := thrift.NewTMemoryBufferLen(n)
		if _, err = thriftBuffer.Write(data[:n]); err != nil {
			log.Error("buffer write failed :%v,", err)
			continue
		}

		protocolFactory := thrift.NewTCompactProtocolFactory()
		thriftProtocol := protocolFactory.GetProtocol(thriftBuffer)
		_, _, _, err = thriftProtocol.ReadMessageBegin()
		if err != nil {
			log.Error("read message begin failed :%v,", err)
			continue
		}

		batch := agent.AgentEmitBatchArgs{}
		err = batch.Read(thriftProtocol)
		if err != nil {
			log.Error("read batch failed :%v,", err)
			continue
		}

		processBatch(batch.Batch)
		if err != nil {
			log.Error("process batch failed :%v,", err)
			continue
		}

		err = thriftProtocol.ReadMessageEnd()
		if err != nil {
			log.Error("read message end failed :%v,", err)
			continue
		}
	}
}
