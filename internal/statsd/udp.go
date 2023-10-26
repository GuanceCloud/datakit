// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package statsd

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

func (col *Collector) setupUDPServer() error {
	if col.UDPlistener == nil {
		address, err := net.ResolveUDPAddr(col.opts.protocol, col.opts.serviceAddress)
		if err != nil {
			col.opts.l.Error(err)
			return err
		}

		conn, err := net.ListenUDP(col.opts.protocol, address)
		if err != nil {
			col.opts.l.Error(err)
			return err
		}

		col.opts.l.Infof("UDP listening on %q", conn.LocalAddr().String())
		col.UDPlistener = conn
	}

	g.Go(func(ctx context.Context) error {
		if err := col.udpListen(col.UDPlistener); err != nil {
			col.opts.l.Warnf("udpListen: %s, ignored", err.Error())
		}
		return nil
	})

	return nil
}

// udpListen starts listening for udp packets on the configured port.
func (col *Collector) udpListen(conn *net.UDPConn) error {
	col.opts.l.Debug("In udpListen.")
	defer func() {
		col.opts.l.Debug("Out udpListen.")
	}()

	if col.opts.readBufferSize > 0 {
		if err := col.UDPlistener.SetReadBuffer(col.opts.readBufferSize); err != nil {
			return err
		}
	}

	buf := make([]byte, UDPMaxPacketSize)
	for {
		select {
		case <-col.done:
			return nil
		default:
			n, addr, err := conn.ReadFromUDP(buf)
			col.opts.l.Debugf("Get conn.ReadFromUDP(buf) bytes: %d %v", n, err)
			if err != nil {
				if !strings.Contains(err.Error(), "closed network") {
					col.opts.l.Errorf("Error reading: %s", err.Error())
					continue
				}
				return err
			}

			col.opts.l.Debugf("UDP: read %d bytes from %s", n, addr.IP.String())

			b, ok := col.bufPool.Get().(*bytes.Buffer)
			if !ok {
				return fmt.Errorf("bufPool is not a bytes buffer")
			}
			b.Reset()
			if _, err := b.Write(buf[:n]); err != nil {
				return err
			}
			select {
			case col.in <- job{
				Buffer: b,
				Time:   time.Now(),
				Addr:   addr.IP.String(),
			}:
			default:
				col.drops++
				if col.drops == 1 || col.opts.allowedPendingMessages == 0 || col.drops%col.opts.allowedPendingMessages == 0 {
					col.opts.l.Errorf("Statsd message queue full. "+
						"We have dropped %d messages so far. "+
						"You may want to increase allowed_pending_messages in the config", col.drops)
				}
			}
		}
	}
}
