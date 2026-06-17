// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
)

func TestWithAPIConfigCopiesRuntimeHTTPFields(t *testing.T) {
	c := config.DefaultAPIConfig()
	c.RequestRateLimit = 42
	c.RequestRateLimitTTL = 7 * time.Second
	c.RequestRateLimitBurst = 11
	c.ReadTimeout = 13 * time.Second
	c.IdleTimeout = 17 * time.Second
	c.ReadHeaderTimeout = 19 * time.Second
	c.WriteTimeout = 23 * time.Second

	hs := defaultHTTPServerConf()
	WithAPIConfig(c)(hs)

	if hs.apiConfig.RequestRateLimit != c.RequestRateLimit {
		t.Fatalf("RequestRateLimit = %v, want %v", hs.apiConfig.RequestRateLimit, c.RequestRateLimit)
	}
	if hs.apiConfig.RequestRateLimitTTL != c.RequestRateLimitTTL {
		t.Fatalf("RequestRateLimitTTL = %v, want %v", hs.apiConfig.RequestRateLimitTTL, c.RequestRateLimitTTL)
	}
	if hs.apiConfig.RequestRateLimitBurst != c.RequestRateLimitBurst {
		t.Fatalf("RequestRateLimitBurst = %v, want %v", hs.apiConfig.RequestRateLimitBurst, c.RequestRateLimitBurst)
	}
	if hs.apiConfig.ReadTimeout != c.ReadTimeout {
		t.Fatalf("ReadTimeout = %v, want %v", hs.apiConfig.ReadTimeout, c.ReadTimeout)
	}
	if hs.apiConfig.IdleTimeout != c.IdleTimeout {
		t.Fatalf("IdleTimeout = %v, want %v", hs.apiConfig.IdleTimeout, c.IdleTimeout)
	}
	if hs.apiConfig.ReadHeaderTimeout != c.ReadHeaderTimeout {
		t.Fatalf("ReadHeaderTimeout = %v, want %v", hs.apiConfig.ReadHeaderTimeout, c.ReadHeaderTimeout)
	}
	if hs.apiConfig.WriteTimeout != c.WriteTimeout {
		t.Fatalf("WriteTimeout = %v, want %v", hs.apiConfig.WriteTimeout, c.WriteTimeout)
	}
}

func TestSetupRequestLimiter(t *testing.T) {
	hs := defaultHTTPServerConf()
	hs.apiConfig.RequestRateLimit = 42
	hs.apiConfig.RequestRateLimitTTL = time.Second
	hs.apiConfig.RequestRateLimitBurst = 11

	setupRequestLimiter(hs)

	if hs.reqLimiter == nil {
		t.Fatal("reqLimiter is nil")
	}
}
