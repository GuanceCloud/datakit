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

		err := udpConn.SetDeadline(time.Now().Add(time.Second))
		if err != nil {
			log.Errorf("SetDeadline failed: %s", err.Error())
			continue
		}

		n, addr, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			log.Debug(err.Error())
			continue
		}
		log.Debugf("Read from udp server:%s %d bytes", addr, n)

		if n <= 0 {
			continue
		}

		dktrace, err := parseJaegerUDP(buf[:n])
		if err != nil {
			log.Error(err.Error())
			continue
		}
		if len(dktrace) == 0 {
			log.Warn("empty datakit trace")
		} else {
			afterGather.Run(inputName, dktrace, false)
		}
	}
}

func parseJaegerUDP(data []byte) (itrace.DatakitTrace, error) {
	thriftBuffer := thrift.NewTMemoryBufferLen(len(data))
	_, err := thriftBuffer.Write(data)
	if err != nil {
		log.Error("buffer write failed :%s", err.Error())

		return nil, err
	}

	var (
		protocolFactory = thrift.NewTCompactProtocolFactoryConf(&thrift.TConfiguration{})
		thriftProtocol  = protocolFactory.GetProtocol(thriftBuffer)
	)
	if _, _, _, err = thriftProtocol.ReadMessageBegin(context.TODO()); err != nil { //nolint:dogsled
		log.Error("read message begin failed :%s,", err.Error())

		return nil, err
	}

	batch := agent.AgentEmitBatchArgs{}
	err = batch.Read(context.TODO(), thriftProtocol)
	if err != nil {
		log.Error("read batch failed :%s,", err.Error())

		return nil, err
	}

	dktrace := batchToDkTrace(batch.Batch)

	if err = thriftProtocol.ReadMessageEnd(context.TODO()); err != nil {
		log.Error("read message end failed :%s,", err.Error())
	}

	return dktrace, err
}
