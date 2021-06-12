package statsd

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"
)

func (s *input) setupUDPServer() {
	address, err := net.ResolveUDPAddr(s.Protocol, s.ServiceAddress)
	if err != nil {
		l.Error(err)
		return
	}

	conn, err := net.ListenUDP(s.Protocol, address)
	if err != nil {
		l.Error(err)
		return
	}

	l.Infof("UDP listening on %q", conn.LocalAddr().String())
	s.UDPlistener = conn

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.udpListen(conn); err != nil {
			l.Errorf("udpListen: %s", err.Error())
		}
	}()
	return
}

// udpListen starts listening for udp packets on the configured port.
func (s *input) udpListen(conn *net.UDPConn) error {
	if s.ReadBufferSize > 0 {
		if err := s.UDPlistener.SetReadBuffer(s.ReadBufferSize); err != nil {
			return err
		}
	}

	buf := make([]byte, UDPMaxPacketSize)
	for {
		select {
		case <-s.done:
			return nil
		default:
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				if !strings.Contains(err.Error(), "closed network") {
					l.Errorf("Error reading: %s", err.Error())
					continue
				}
				return err
			}

			l.Debugf("UDP: read %d bytes from %s", n, addr.IP.String())

			b, ok := s.bufPool.Get().(*bytes.Buffer)
			if !ok {
				return fmt.Errorf("bufPool is not a bytes buffer")
			}
			b.Reset()
			if _, err := b.Write(buf[:n]); err != nil {
				return err
			}
			select {
			case s.in <- job{
				Buffer: b,
				Time:   time.Now(),
				Addr:   addr.IP.String()}:
			default:
				s.drops++
				if s.drops == 1 || s.AllowedPendingMessages == 0 || s.drops%s.AllowedPendingMessages == 0 {
					l.Errorf("Statsd message queue full. "+
						"We have dropped %d messages so far. "+
						"You may want to increase allowed_pending_messages in the config", s.drops)
				}
			}
		}
	}
}
