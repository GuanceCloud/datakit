// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package statsd

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (col *Collector) setupUnixServer() error {
	if col.UnixConn == nil {
		if col.opts.serviceUnixAddress == "" {
			return nil
		}
		conn, err := initUnixgramListener(col.opts.serviceUnixAddress)
		if err != nil {
			col.opts.l.Error(err)
			return err
		}
		col.opts.l.Info("Unix listening on %1", conn.LocalAddr().String())
		col.UnixConn = conn
	}

	g.Go(func(ctx context.Context) error {
		if err := col.unixListen(col.UnixConn); err != nil {
			col.opts.l.Warnf("unixListen: %s, ignored", err.Error())
		}
		return nil
	})

	return nil
}

func (col *Collector) unixListen(conn *net.UnixConn) error {
	col.opts.l.Debug("In unixListen.")
	defer func() {
		col.opts.l.Debug("Out unixListen.")
	}()

	if col.opts.readBufferSize > 0 {
		if err := col.UnixConn.SetReadBuffer(col.opts.readBufferSize); err != nil {
			return err
		}
	}

	buf := make([]byte, UDPMaxPacketSize)
	for {
		select {
		case <-col.done:
			return nil
		default:
			n, addr, err := conn.ReadFromUnix(buf)
			col.opts.l.Debugf("Get conn.ReadFromUDP(buf) bytes: %d %v", n, err)
			if err != nil {
				if !strings.Contains(err.Error(), "closed network") {
					col.opts.l.Errorf("Error reading: %s", err.Error())
					continue
				}
				return err
			}

			col.opts.l.Debugf("UDP: read %d bytes from %s", n, addr.String())

			httpGetBytesVec.WithLabelValues().Observe(float64(n))

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
				Addr:   addr.String(),
			}:
			default:
				col.dropsUnix++
				if col.dropsUnix == 1 || col.opts.allowedPendingMessages == 0 || col.dropsUnix%col.opts.allowedPendingMessages == 0 {
					col.opts.l.Errorf("Statsd message queue full. "+
						"We have dropped %d messages so far. "+
						"You may want to increase allowed_pending_messages in the config", col.dropsUnix)
				}
			}
		}
	}
}

func initUnixgramListener(udsPath string) (*net.UnixConn, error) {
	var listener *net.UnixConn

	if filepath.IsAbs(udsPath) {
		_ = os.MkdirAll(filepath.Dir(udsPath), 0o755) //nolint:gosec
		if fi, err := os.Stat(udsPath); err == nil {
			if fi.Mode()&os.ModeSocket == 0 {
				return nil, fmt.Errorf("reuse %s faild: file mode %s", udsPath,
					fi.Mode().String())
			}
			if err = os.Remove(udsPath); err != nil && !os.IsNotExist(err) {
				return nil, fmt.Errorf("remove %s: %w", udsPath, err)
			}
		}

		addr, err := net.ResolveUnixAddr("unixgram", udsPath)
		if err != nil {
			return nil, err
		}
		if listener, err = net.ListenUnixgram("unixgram", addr); err != nil {
			return nil, fmt.Errorf(`net.Listen("unix"): %w`, err)
		}
		if err := os.Chmod(udsPath, 0o722); err != nil { //nolint:gosec
			return nil, fmt.Errorf("setting socket permissions failed: %w", err)
		}

		return listener, nil
	} else {
		return nil, fmt.Errorf("uds path %s is not absolute", udsPath)
	}
}
