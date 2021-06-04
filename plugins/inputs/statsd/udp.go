package statsd

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func (x *input) setupUDPServer() error {
	addr, err := net.ResolveUDPAddr("udp", x.Bind)
	if err != nil {
		l.Error(err)
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		l.Error(err)
		return err
	}

	x.udpListener = conn
	x.wg.Add(1)

	go func() {
		defer x.wg.Done()
		if err := x.udpWait(); err != nil {
			l.Error(err)
		}
	}()

	return nil
}

func (x *input) udpWait() error {
	buf := make([]byte, udpMaxPktSize)
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("statsd udpListen exit.")
			return nil

		default:
			n, addr, err := x.udpListener.ReadFromUDP(buf)
			if err != nil {
				if !strings.Contains(err.Error(), "closed network") {
					l.Error("error reading: %s", err.Error())
					continue
				}
				return err
			}

			b, ok := x.bufpool.Get().(*bytes.Buffer)
			if !ok {
				return fmt.Errorf("unexpected bufpool")
			}

			b.Reset()
			if _, err := b.Write(buf[:n]); err != nil {
				l.Error(err)
				return err
			}

			select {
			case x.in <- &job{
				Buffer: b,
				Time:   time.Now(),
				Addr:   addr.IP.String()}:
			default:
				x.drops++
				l.Warnf("dropped %d messages", x.drops)
			}
		}
	}
}
