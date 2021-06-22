package statsd

import (
	"bufio"
	"bytes"
	"net"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
)

func (s *input) setupTCPServer() {

	address, err := net.ResolveTCPAddr("tcp", s.ServiceAddress)
	if err != nil {
		l.Error(err)
		return
	}
	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		l.Error(err)
		return
	}

	l.Infof("TCP listening on %q", listener.Addr().String())
	s.TCPlistener = listener

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.tcpListen(listener); err != nil {
			l.Errorf("tcpListen: %s", err.Error())
		}
	}()
}

// tcpListen() starts listening for udp packets on the configured port.
func (s *input) tcpListen(listener *net.TCPListener) error {
	for {
		select {
		case <-s.done:
			return nil
		default:
			// Accept connection:
			conn, err := listener.AcceptTCP()
			if err != nil {
				return err
			}

			if s.TCPKeepAlive {
				if err = conn.SetKeepAlive(true); err != nil {
					return err
				}

				if s.TCPKeepAlivePeriod != nil {
					if err = conn.SetKeepAlivePeriod(s.TCPKeepAlivePeriod.Duration); err != nil {
						return err
					}
				}
			}

			select {
			case <-s.accept:
				// not over connection limit, handle the connection properly.
				s.wg.Add(1)
				// generate a random id for this TCPConn
				id := cliutils.XID("tcp_")
				s.remember(id, conn)
				go s.handler(conn, id)
			default:
				// We are over the connection limit, refuse & close.
				s.refuser(conn)
			}
		}
	}
}

// handler handles a single TCP Connection
func (s *input) handler(conn *net.TCPConn, id string) {
	// connection cleanup function
	defer func() {
		s.wg.Done()

		// Ignore the returned error as we cannot do anything about it anyway
		//nolint:errcheck,revive
		conn.Close()

		// Add one connection potential back to channel when this one closes
		s.accept <- true
		s.forget(id)
	}()

	var remoteIP string
	if addr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		remoteIP = addr.IP.String()
	}

	var n int
	scanner := bufio.NewScanner(conn)
	for {
		select {
		case <-s.done:
			return
		default:
			if !scanner.Scan() {
				return
			}
			n = len(scanner.Bytes())
			if n == 0 {
				continue
			}

			b := s.bufPool.Get().(*bytes.Buffer)
			b.Reset()
			// Writes to a bytes buffer always succeed, so do not check the errors here
			//nolint:errcheck,revive
			b.Write(scanner.Bytes())
			//nolint:errcheck,revive
			b.WriteByte('\n')

			select {
			case s.in <- job{Buffer: b, Time: time.Now(), Addr: remoteIP}:
			default:
				s.drops++
				if s.drops == 1 || s.drops%s.AllowedPendingMessages == 0 {
					l.Errorf("Statsd message queue full. "+
						"We have dropped %d messages so far. "+
						"You may want to increase allowed_pending_messages in the config", s.drops)
				}
			}
		}
	}
}

// refuser refuses a TCP connection
func (s *input) refuser(conn *net.TCPConn) {
	// Ignore the returned error as we cannot do anything about it anyway
	//nolint:errcheck,revive
	conn.Close()
	l.Infof("Refused TCP Connection from %s", conn.RemoteAddr())
	l.Warn("Maximum TCP Connections reached, you may want to adjust max_tcp_connections")
}

// forget a TCP connection
func (s *input) forget(id string) {
	s.cleanup.Lock()
	defer s.cleanup.Unlock()
	delete(s.conns, id)
}

// remember a TCP connection
func (s *input) remember(id string, conn *net.TCPConn) {
	s.cleanup.Lock()
	defer s.cleanup.Unlock()
	s.conns[id] = conn
}
