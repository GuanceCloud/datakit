// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package socket

import (
	"fmt"
	"net"
	"net/netip"
	"net/url"
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

type tagger struct {
	election, host map[string]string
}

func (t *tagger) ElectionTags() map[string]string {
	return t.election
}

func (t *tagger) HostTags() map[string]string {
	return t.host
}

type feeder struct {
	errs    []*io.LastError
	errStrs []string
}

func (f *feeder) Feed(name string,
	category point.Category,
	pts []*point.Point,
	opt ...*io.Option,
) error {
	return nil
}

func (f *feeder) FeedLastError(err string, opts ...io.LastErrorOption) {
	le := &io.LastError{}
	for _, opt := range opts {
		opt(le)
	}

	f.errs = append(f.errs, le)
	f.errStrs = append(f.errStrs, err)
}

func listenUDP(urlStr string) (net.Conn, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "udp":

		x, err := netip.ParseAddrPort(u.Host)
		if err != nil {
			return nil, err
		}

		addr := net.UDPAddr{
			IP:   net.ParseIP(x.Addr().String()),
			Port: int(x.Port()),
		}

		return net.ListenUDP("udp", &addr)

	default:
		return nil, fmt.Errorf("not UDP, got %q", u.Scheme)
	}
}

func listenTCP(t *T.T, urlStr string) (net.Listener, error) {
	t.Helper()

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "tcp":
		t.Logf("listen to %q", u.Host)
		listener, err := net.Listen("tcp", u.Host)
		if err != nil {
			return nil, err
		}

		go func() {
			for {
				conn, _ := listener.Accept() // block here

				if conn != nil {
					t.Logf("accept %s", conn.RemoteAddr().String())
				}
			}
		}()

		return listener, nil
	default:
		return nil, fmt.Errorf("not TCP, got %q", u.Scheme)
	}
}

func TestCollect(t *T.T) {
	tagger := &tagger{
		election: map[string]string{
			"et1": "xx",
			"et2": "yy",
		},
		host: map[string]string{
			"t1": "xxx",
			"t2": "yyy",
		},
	}

	t.Run("udp", func(t *T.T) {
		udpConn, port, err := testutils.RandPortUDP()
		assert.NoError(t, err)
		assert.NoError(t, udpConn.Close())

		urls := []string{
			fmt.Sprintf("udp://127.0.0.1:%d", port),
		}

		t.Logf("urls: %q", urls)

		feeder := &feeder{}

		conn, err := listenUDP(urls[0])
		assert.NoError(t, err)

		t.Cleanup(func() {
			assert.NoError(t, conn.Close())
		})

		i := &input{
			DestURL:    urls,
			UDPTimeOut: datakit.Duration{Duration: time.Second * 10},
			// TCPTimeOut: datakit.Duration{Duration: time.Second * 10},

			Election: true,

			feeder: feeder,
			tagger: tagger,
		}

		i.setup()
		i.Collect()
		assert.Len(t, i.collectCache, 1)

		pt := i.collectCache[0]
		t.Log("point line =", pt.LineProto())
		assert.Equal(t, "127.0.0.1", pt.Get("dest_host").(string))
		assert.Equal(t, fmt.Sprintf("%d", port), pt.Get("dest_port").(string))
		assert.Equal(t, "udp", pt.Get("proto").(string))
		assert.Equal(t, int64(1), pt.Get("success").(int64))
		assert.Equal(t, "xx", pt.Get("et1").(string))
		assert.Equal(t, "yy", pt.Get("et2").(string))

		for _, pt := range i.collectCache {
			t.Log(pt.Pretty())
		}
	})

	t.Run("tcp", func(t *T.T) {
		port := testutils.RandPort("tcp")
		urls := []string{
			fmt.Sprintf("tcp://127.0.0.1:%d", port),
		}

		t.Logf("urls: %q", urls)

		listener, err := listenTCP(t, urls[0])
		assert.NoError(t, err)

		defer listener.Close()

		time.Sleep(time.Second) // wait TCP server OK

		feeder := &feeder{}

		i := &input{
			DestURL:    urls,
			TCPTimeOut: datakit.Duration{Duration: time.Second * 10},

			Election: true,

			feeder: feeder,
			tagger: tagger,
		}

		i.setup()
		i.Collect()
		assert.Len(t, i.collectCache, 1)

		pt := i.collectCache[0]
		assert.Equal(t, "127.0.0.1", pt.Get("dest_host").(string))
		assert.Equal(t, fmt.Sprintf("%d", port), pt.Get("dest_port").(string))
		assert.Equal(t, "tcp", pt.Get("proto").(string))
		assert.Equal(t, int64(1), pt.Get("success").(int64))
		assert.True(t, pt.Get("response_time_with_dns").(int64) > 0)
		assert.True(t, pt.Get("response_time").(int64) > 0)
		assert.Equal(t, "xx", pt.Get("et1").(string))
		assert.Equal(t, "yy", pt.Get("et2").(string))

		for _, pt := range i.collectCache {
			t.Log(pt.Pretty())
		}
	})

	t.Run("fail-tcp", func(t *T.T) {
		urls := []string{
			"tcp://127.0.0.1:0",
		}

		listener, err := listenTCP(t, urls[0])
		assert.NoError(t, err)

		defer listener.Close()

		time.Sleep(time.Second) // wait TCP server OK

		feeder := &feeder{}

		i := &input{
			DestURL:    urls,
			TCPTimeOut: datakit.Duration{Duration: time.Second * 10},

			Election: true,

			feeder: feeder,
			tagger: tagger,
		}

		i.setup()
		i.Collect()
		assert.Len(t, i.collectCache, 1)

		pt := i.collectCache[0]
		assert.Equal(t, "127.0.0.1", pt.Get("dest_host").(string))
		assert.Equal(t, "0", pt.Get("dest_port").(string))
		assert.Equal(t, "tcp", pt.Get("proto").(string))
		assert.Equal(t, int64(-1), pt.Get("success").(int64))
		assert.True(t, pt.Get("response_time_with_dns").(int64) == 0)
		assert.True(t, pt.Get("response_time").(int64) == 0)
		assert.Equal(t, "xx", pt.Get("et1").(string))
		assert.Equal(t, "yy", pt.Get("et2").(string))

		for _, pt := range i.collectCache {
			t.Log(pt.Pretty())
		}

		assert.Len(t, feeder.errStrs, 1)
		t.Logf("le: %q", feeder.errs[0])
	})

	t.Run("fail-udp", func(t *T.T) {
		urls := []string{
			"udp://127.0.0.1:0",
		}

		feeder := &feeder{}

		conn, err := listenUDP(urls[0])
		assert.NoError(t, err)

		t.Cleanup(func() {
			assert.NoError(t, conn.Close())
		})

		i := &input{
			DestURL:    urls,
			UDPTimeOut: datakit.Duration{Duration: time.Second * 10},
			// TCPTimeOut: datakit.Duration{Duration: time.Second * 10},

			Election: true,

			feeder: feeder,
			tagger: tagger,
		}

		i.setup()
		i.Collect()
		assert.Len(t, i.collectCache, 1)

		pt := i.collectCache[0]
		assert.Equal(t, "127.0.0.1", pt.Get("dest_host").(string))
		assert.Equal(t, "0", pt.Get("dest_port").(string))
		assert.Equal(t, "udp", pt.Get("proto").(string))
		assert.Equal(t, int64(-1), pt.Get("success").(int64))
		assert.Equal(t, "xx", pt.Get("et1").(string))
		assert.Equal(t, "yy", pt.Get("et2").(string))
		for _, pt := range i.collectCache {
			t.Log(pt.Pretty())
		}

		t.Logf("last err strings: %q", feeder.errStrs)
	})

	t.Run("invalid-schema", func(t *T.T) {
		urls := []string{
			"http://127.0.0.1:0",
		}

		feeder := &feeder{}

		i := &input{
			DestURL: urls,

			Election: true,

			feeder: feeder,
			tagger: tagger,
		}

		i.setup()
		i.Collect()
		assert.Len(t, i.collectCache, 0)

		t.Logf("last err strings: %q", feeder.errStrs)
	})
}
