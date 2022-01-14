package jaeger

import (
	"context"
	"net"
	"time"

	"github.com/uber/jaeger-client-go/thrift"
	"github.com/uber/jaeger-client-go/thrift-gen/agent"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
)

func StartUDPAgent(addr string) error {
	data := make([]byte, 65535)

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	log.Infof("Jaeger UDP agent listening on %s", addr)

	// receiving loop
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

		err := udpConn.SetDeadline(time.Now().Add(time.Second))
		if err != nil {
			log.Errorf("SetDeadline failed: %v", err)
			continue
		}

		n, addr, err := udpConn.ReadFromUDP(data)
		if err != nil {
			log.Debug(err.Error())
			continue
		}
		log.Debugf("Read from udp server:%s %d bytes", addr, n)

		if n <= 0 {
			continue
		}

		groups, err := parseJaegerUDP(data[:n])
		if err != nil {
			continue
		}
		if len(groups) != 0 {
			itrace.MkLineProto(groups, inputName)
		} else {
			log.Debug("empty batch")
		}
	}
}

func parseJaegerUDP(data []byte) ([]*itrace.DatakitSpan, error) {
	thriftBuffer := thrift.NewTMemoryBufferLen(len(data))
	if _, err := thriftBuffer.Write(data); err != nil {
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

	groups, err := batchToAdapters(batch.Batch)
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
