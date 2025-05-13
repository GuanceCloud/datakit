// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/stretchr/testify/assert"
)

func TestMakeServer(t *testing.T) {
	cases := []struct {
		in   string
		fail bool
	}{
		// we use random port(:0) here, see https://stackoverflow.com/a/43425461/342348
		{
			in: "tcp://127.0.0.1:0",
		},
		{
			in: "udp://127.0.0.1:0",
		},
		{
			in:   "udp1://127.0.0.1:0",
			fail: true,
		},
		{
			in:   "udp127.0.0.1:0",
			fail: true,
		},
	}

	for _, tc := range cases {
		sk := &SocketLogger{
			opt: &option{
				source:  "testing",
				sockets: []string{tc.in},
			},
		}
		sk.log = logger.SLogger("socketLog/" + sk.opt.source)

		err := sk.makeServer()
		if tc.fail && assert.Error(t, err) {
			continue
		} else {
			assert.NoError(t, err)
		}

		assert.Equal(t, len(sk.opt.sockets), len(sk.servers))
		assert.NotNil(t, sk.servers[0])
	}
}

func TestForwardMessage(t *testing.T) {
	inMultilineMatch := []string{`^\d+-\d+-\d+ \d+:\d+:\d+(,\d+)?`}

	cases := []struct {
		inScheme string
		inData   [][]byte
		out      [][]byte
	}{
		{
			inScheme: "tcp",
			inData: [][]byte{
				[]byte("2021-07-08 05:08:19,214 INFO Testing output-01.\n"),
				[]byte("unexpected output\n"),
				[]byte("123456\n"),
				[]byte("2021-07-08 05:08:19,214 INFO Testing output-02.\n"),
			},
			out: [][]byte{
				[]byte("2021-07-08 05:08:19,214 INFO Testing output-01.\nunexpected output\n123456"),
				// []byte("2021-07-08 05:08:19,214 INFO Testing output-02.\n"), // in multiline cache
			},
		},
		{
			inScheme: "udp",
			inData: [][]byte{
				[]byte("2021-07-08 05:08:19,214 INFO Testing output-01."),
				[]byte("unexpected output\n123456"),
				[]byte("2021-07-08 05:08:19,214 INFO Testing output-02."),
			},
			out: [][]byte{
				[]byte("2021-07-08 05:08:19,214 INFO Testing output-01."),
				[]byte("unexpected output"),
				[]byte("123456"),
				[]byte("2021-07-08 05:08:19,214 INFO Testing output-02."),
			},
		},
	}

	for _, tc := range cases {
		var (
			srv server
			err error

			opt = &option{
				multilinePatterns: inMultilineMatch,
			}
			address = "127.0.0.1:0" // make server port random
		)

		switch tc.inScheme {
		case "tcp":
			srv, err = newTCPServer(tc.inScheme, address, opt)
			address = srv.(*tcpServer).listener.Addr().(*net.TCPAddr).String()
			t.Logf("TCP server address: %s", address)
		case "udp":
			srv, err = newUDPServer(tc.inScheme, address)
			address = srv.(*udpServer).conn.LocalAddr().String()
			t.Logf("UDP server address: %s", address)
		default:
			t.Error("invalid scheme")
		}

		assert.NoError(t, err)
		assert.NotNil(t, srv)

		res := make(chan [][]byte, len(tc.out))
		feed := func(pending [][]byte) {
			res <- pending
		}

		ctx, cancel := context.WithCancel(context.Background())
		listenerReady := make(chan struct{})
		go func() {
			close(listenerReady)
			err := srv.forwardMessage(ctx, feed)
			assert.ErrorIs(t, err, net.ErrClosed)
		}()

		<-listenerReady
		err = send(tc.inScheme, address, tc.inData)
		assert.NoError(t, err)

		pending := [][]byte{}
		count := &atomic.Int64{}

		for {
			if count.Load() == int64(len(tc.out)) {
				break
			}
			x, ok := <-res
			if !ok {
				break
			}
			pending = append(pending, x...)
			count.Add(int64(len(x)))
		}

		cancel()
		srv.close()

		assert.Equal(t, tc.out, pending)
	}
}

func send(scheme, address string, data [][]byte) error {
	conn, err := net.DialTimeout(scheme, address, time.Second)
	if err != nil {
		return err
	}
	defer conn.Close() //nolint:errcheck

	for _, b := range data {
		if _, err = conn.Write(b); err != nil {
			return err
		}
	}
	return nil
}
