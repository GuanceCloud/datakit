package logger

import (
	"fmt"
	"net"
	"strings"
)

type remoteEndpoint struct {
	protocol string
	host     string

	tcpConn *net.TCPConn
	udpConn *net.UDPConn
}

func newRemoteSync(proto, iphost string) (*remoteEndpoint, error) {
	switch strings.ToLower(proto) {
	case SchemeTCP:
		addr, err := net.ResolveTCPAddr("tcp", iphost)
		if err != nil {
			return nil, err
		}

		conn, err := net.DialTCP("tcp", nil, addr)
		if err != nil {
			return nil, err
		}

		return &remoteEndpoint{
			protocol: proto,
			host:     iphost,
			tcpConn:  conn,
		}, nil

	case SchemeUDP:
		addr, err := net.ResolveUDPAddr("udp", iphost)
		if err != nil {
			return nil, err
		}
		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			return nil, err
		}

		return &remoteEndpoint{
			protocol: proto,
			host:     iphost,
			udpConn:  conn,
		}, nil
	default:
		return nil, fmt.Errorf("unknown remote protocol: %s", proto)
	}
}

func (te *remoteEndpoint) Write(data []byte) (int, error) {
	switch strings.ToLower(te.protocol) {
	case SchemeTCP:
		return te.tcpConn.Write(data)

	case SchemeUDP:
		return te.udpConn.Write(data)

	default:
		return -1, fmt.Errorf("unknown remote protocol: %s", te.protocol)
	}
}

func (te *remoteEndpoint) Sync() error { return nil } // do nothing
