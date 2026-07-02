// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package net

import (
	"context"
	"errors"
	"net"
	"sync"
	T "testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateIPFamilyPolicy(t *T.T) {
	for _, policy := range []string{
		IPFamilyPolicyAuto,
		IPFamilyPolicyPreferIPv6,
		IPFamilyPolicyIPv6Only,
		IPFamilyPolicyIPv4Only,
	} {
		assert.NoError(t, ValidateIPFamilyPolicy(policy))
	}
	assert.Error(t, ValidateIPFamilyPolicy("invalid"))
}

func TestDNSCacheDialerPreferIPv6(t *T.T) {
	t.Run("IPv6 wins without fallback", func(t *T.T) {
		var fallbackCount int
		d := newTestDNSCacheDialer(func(_ context.Context, network, addr string) (net.Conn, error) {
			return newStubConn(network, addr), nil
		})
		d.fallbackDelay = time.Second
		d.fallbackStarted = func() { fallbackCount++ }

		conn, err := d.DialContext(context.Background(), "tcp", "dataway.example.com:9528")
		require.NoError(t, err)
		assert.Equal(t, "tcp6/[2001:db8::10]:9528", conn.RemoteAddr().String())
		assert.Equal(t, 0, fallbackCount)
		require.NoError(t, conn.Close())
	})

	t.Run("immediate IPv6 failure starts IPv4 immediately", func(t *T.T) {
		var fallbackCount int
		d := newTestDNSCacheDialer(func(_ context.Context, network, addr string) (net.Conn, error) {
			if network == "tcp6" {
				return nil, errors.New("network unreachable")
			}
			return newStubConn(network, addr), nil
		})
		d.fallbackDelay = time.Second
		d.fallbackStarted = func() { fallbackCount++ }

		start := time.Now()
		conn, err := d.DialContext(context.Background(), "tcp", "dataway.example.com:9528")
		require.NoError(t, err)
		assert.Less(t, time.Since(start), 500*time.Millisecond)
		assert.Equal(t, "tcp4/192.0.2.10:9528", conn.RemoteAddr().String())
		assert.Equal(t, 1, fallbackCount)
		require.NoError(t, conn.Close())
	})

	t.Run("hanging IPv6 races IPv4 after delay", func(t *T.T) {
		const fallbackDelay = 30 * time.Millisecond
		d := newTestDNSCacheDialer(func(ctx context.Context, network, addr string) (net.Conn, error) {
			if network == "tcp6" {
				<-ctx.Done()
				return nil, ctx.Err()
			}
			return newStubConn(network, addr), nil
		})
		d.fallbackDelay = fallbackDelay

		start := time.Now()
		conn, err := d.DialContext(context.Background(), "tcp", "dataway.example.com:9528")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, time.Since(start), fallbackDelay)
		assert.Equal(t, "tcp4/192.0.2.10:9528", conn.RemoteAddr().String())
		require.NoError(t, conn.Close())
	})
}

func TestDNSCacheDialerOnlyPolicies(t *T.T) {
	tests := []struct {
		name           string
		policy         string
		expectedNet    string
		expectedRemote string
	}{
		{
			name:           "IPv6 only",
			policy:         IPFamilyPolicyIPv6Only,
			expectedNet:    "tcp6",
			expectedRemote: "tcp6/[2001:db8::10]:9528",
		},
		{
			name:           "IPv4 only",
			policy:         IPFamilyPolicyIPv4Only,
			expectedNet:    "tcp4",
			expectedRemote: "tcp4/192.0.2.10:9528",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *T.T) {
			var gotNetwork string
			d := newTestDNSCacheDialer(func(_ context.Context, network, addr string) (net.Conn, error) {
				gotNetwork = network
				return newStubConn(network, addr), nil
			})
			d.ipFamilyPolicy = tc.policy

			conn, err := d.DialContext(context.Background(), "tcp", "dataway.example.com:9528")
			require.NoError(t, err)
			assert.Equal(t, tc.expectedNet, gotNetwork)
			assert.Equal(t, tc.expectedRemote, conn.RemoteAddr().String())
			require.NoError(t, conn.Close())
		})
	}
}

func TestDNSCacheDialerReturnsBothFamilyErrors(t *T.T) {
	d := newTestDNSCacheDialer(func(_ context.Context, network, _ string) (net.Conn, error) {
		if network == "tcp6" {
			return nil, errors.New("IPv6 failed")
		}
		return nil, errors.New("IPv4 failed")
	})

	_, err := d.DialContext(context.Background(), "tcp", "dataway.example.com:9528")
	require.Error(t, err)
	assert.ErrorContains(t, err, "IPv6 failed")
	assert.ErrorContains(t, err, "IPv4 failed")
}

func TestDNSCacheDialerConnectionObserver(t *T.T) {
	var opened, closed int
	d := newTestDNSCacheDialer(func(_ context.Context, network, addr string) (net.Conn, error) {
		return newStubConn(network, addr), nil
	})
	d.connectionOpened = func(family, remote string) {
		opened++
		assert.Equal(t, "ipv6", family)
		assert.Equal(t, "tcp6/[2001:db8::10]:9528", remote)
	}
	d.connectionClosed = func(family string) {
		closed++
		assert.Equal(t, "ipv6", family)
	}

	conn, err := d.DialContext(context.Background(), "tcp", "dataway.example.com:9528")
	require.NoError(t, err)
	assert.Equal(t, 1, opened)
	require.NoError(t, conn.Close())
	require.NoError(t, conn.Close())
	assert.Equal(t, 1, closed)
}

func newTestDNSCacheDialer(dial DialFunc) *dnsCacheDialer {
	return &dnsCacheDialer{
		fetch: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{
				net.ParseIP("192.0.2.10"),
				net.ParseIP("2001:db8::10"),
			}, nil
		},
		dial:           dial,
		fallbackDelay:  DefaultIPv4FallbackDelay,
		ipFamilyPolicy: IPFamilyPolicyPreferIPv6,
	}
}

type stubAddr string

func (a stubAddr) Network() string { return "tcp" }
func (a stubAddr) String() string  { return string(a) }

type stubConn struct {
	remote stubAddr
	once   sync.Once
}

func newStubConn(network, addr string) *stubConn {
	return &stubConn{remote: stubAddr(network + "/" + addr)}
}

func (c *stubConn) Read([]byte) (int, error)         { return 0, errors.New("not implemented") }
func (c *stubConn) Write(p []byte) (int, error)      { return len(p), nil }
func (c *stubConn) LocalAddr() net.Addr              { return stubAddr("local") }
func (c *stubConn) RemoteAddr() net.Addr             { return c.remote }
func (c *stubConn) SetDeadline(time.Time) error      { return nil }
func (c *stubConn) SetReadDeadline(time.Time) error  { return nil }
func (c *stubConn) SetWriteDeadline(time.Time) error { return nil }
func (c *stubConn) Close() error {
	c.once.Do(func() {})
	return nil
}
