// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

const (
	defaultReverseDNSEnrichmentBudget = 2 * time.Second
	defaultReverseDNSWorkers          = 8
)

type reverseDNSEnricher struct {
	enabled bool
	timeout time.Duration
	ttl     time.Duration
	maxSize int

	lookupFn    func(context.Context, string) ([]string, error)
	lookupSlots chan struct{}

	mu    sync.Mutex
	cache map[string]reverseDNSCacheEntry
}

type reverseDNSCacheEntry struct {
	name    string
	ok      bool
	expires time.Time
}

type reverseDNSLookupResult struct {
	ip   string
	name string
}

func normalizeReverseDNSConfig(cfg *ReverseDNSConfig) ReverseDNSConfig {
	out := ReverseDNSConfig{}
	if cfg != nil {
		out = *cfg
	}
	if out.Timeout == nil || out.Timeout.Duration <= 0 {
		out.Timeout = &datakit.Duration{Duration: defaultReverseDNSTimeout}
	}
	if out.CacheTTL == nil || out.CacheTTL.Duration <= 0 {
		out.CacheTTL = &datakit.Duration{Duration: defaultReverseDNSCacheTTL}
	}
	if out.CacheSize <= 0 {
		out.CacheSize = defaultReverseDNSCacheSize
	} else if out.CacheSize > maxReverseDNSCacheSize {
		out.CacheSize = maxReverseDNSCacheSize
	}
	return out
}

func newReverseDNSEnricher(cfg ReverseDNSConfig) *reverseDNSEnricher {
	return &reverseDNSEnricher{
		enabled:     cfg.Enabled,
		timeout:     cfg.Timeout.Duration,
		ttl:         cfg.CacheTTL.Duration,
		maxSize:     cfg.CacheSize,
		lookupFn:    net.DefaultResolver.LookupAddr,
		lookupSlots: make(chan struct{}, defaultReverseDNSWorkers),
		cache:       map[string]reverseDNSCacheEntry{},
	}
}

func (e *reverseDNSEnricher) lookupContext(ctx context.Context, ip string) (string, bool) {
	if e == nil || !e.enabled {
		return "", false
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Err() != nil {
		return "", false
	}
	ip = strings.TrimSpace(ip)
	if ip == "" || ip == "*" || net.ParseIP(ip) == nil {
		return "", false
	}

	now := time.Now()
	e.mu.Lock()
	if entry, ok := e.cache[ip]; ok && now.Before(entry.expires) {
		e.mu.Unlock()
		return entry.name, entry.ok
	}
	e.mu.Unlock()
	if e.lookupSlots != nil {
		select {
		case e.lookupSlots <- struct{}{}:
			defer func() { <-e.lookupSlots }()
		case <-ctx.Done():
			return "", false
		}

		// A lookup that held a slot may have populated the cache while this
		// caller was waiting.
		now = time.Now()
		e.mu.Lock()
		if entry, ok := e.cache[ip]; ok && now.Before(entry.expires) {
			e.mu.Unlock()
			return entry.name, entry.ok
		}
		e.mu.Unlock()
	}

	lookupCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	names, err := e.lookupFn(lookupCtx, ip)
	if ctx.Err() != nil {
		return "", false
	}
	name := ""
	ok := false
	if err == nil && len(names) > 0 {
		name = strings.TrimSuffix(strings.TrimSpace(names[0]), ".")
		ok = name != ""
	}

	e.mu.Lock()
	if len(e.cache) >= e.maxSize {
		e.cache = map[string]reverseDNSCacheEntry{}
	}
	e.cache[ip] = reverseDNSCacheEntry{
		name:    name,
		ok:      ok,
		expires: time.Now().Add(e.ttl),
	}
	e.mu.Unlock()

	return name, ok
}

func newReverseDNSEnrichmentContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, defaultReverseDNSEnrichmentBudget)
}

func (e *reverseDNSEnricher) lookupManyContext(ctx context.Context, ips []string) map[string]string {
	names := make(map[string]string)
	if e == nil || !e.enabled || len(ips) == 0 {
		return names
	}
	if ctx == nil {
		ctx = context.Background()
	}

	unique := make([]string, 0, len(ips))
	seen := make(map[string]struct{}, len(ips))
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip == "" || ip == "*" || net.ParseIP(ip) == nil {
			continue
		}
		if _, ok := seen[ip]; ok {
			continue
		}
		seen[ip] = struct{}{}
		unique = append(unique, ip)
	}
	if len(unique) == 0 {
		return names
	}

	jobs := make(chan string, len(unique))
	results := make(chan reverseDNSLookupResult, len(unique))
	for _, ip := range unique {
		jobs <- ip
	}
	close(jobs)

	workers := defaultReverseDNSWorkers
	if workers > len(unique) {
		workers = len(unique)
	}
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for {
				if ctx.Err() != nil {
					return
				}
				select {
				case <-ctx.Done():
					return
				case ip, ok := <-jobs:
					if !ok {
						return
					}
					if name, ok := e.lookupContext(ctx, ip); ok {
						results <- reverseDNSLookupResult{ip: ip, name: name}
					}
				}
			}
		}()
	}
	wg.Wait()
	close(results)

	for result := range results {
		names[result.ip] = result.name
	}
	return names
}
