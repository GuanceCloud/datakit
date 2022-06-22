package net

import (
	"context"
	"net"
	"time"

	dnscache "go.mercari.io/go-dnscache"
	"go.uber.org/zap"
)

type DialFunc func(ctx context.Context, network, addr string) (net.Conn, error)

func GetDNSCacheDialContext(freq time.Duration, lookupTimeout time.Duration) (DialFunc, error) {
	// resolver, err := dnscache.New(3*time.Second, 5*time.Second, zap.NewNop())
	resolver, err := dnscache.New(freq, lookupTimeout, zap.NewNop())
	if err != nil {
		return nil, err
	}

	// You can create a HTTP client which selects an IP from dnscache
	// randomly and dials it.
	// rand.Seed(time.Now().UTC().UnixNano()) // You MUST run in once in your application

	return DialFunc(dnscache.DialFunc(resolver, nil)), nil
}
