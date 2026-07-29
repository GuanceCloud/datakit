// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

import (
	"context"
	"errors"
	"net"
	"runtime"
	"testing"
	"time"

	dt "github.com/GuanceCloud/cliutils/dialtesting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeDialTraceroute(t *testing.T) {
	for _, fields := range []map[string]interface{}{
		{},
		{"traceroute": ""},
		{"traceroute": "[]"},
	} {
		normalizeDialTraceroute(fields)
		assert.JSONEq(t,
			`{"runs":[],"hop_count":{"avg":0,"min":0,"max":0}}`,
			fields["traceroute"].(string))
	}

	const populated = `{"runs":[{"run_id":"1","hops":[]}],"hop_count":{"avg":0,"min":0,"max":0}}`
	fields := map[string]interface{}{"traceroute": populated}
	normalizeDialTraceroute(fields)
	assert.Equal(t, populated, fields["traceroute"])
}

func TestValidateDialPlatform(t *testing.T) {
	// Unsupported OS must be rejected regardless of protocol.
	for _, goos := range []string{"windows", "unknown", ""} {
		err := validateDialPlatformFor(goos, "icmp")
		assert.Error(t, err, "goos=%q", goos)
		assert.Contains(t, err.Error(), "unsupported")
	}

	// On a supported OS, ICMP must not be rejected by the platform gate.
	if netpathSupportedOS(runtime.GOOS) {
		assert.NoError(t, validateDialPlatformFor(runtime.GOOS, "icmp"))
	}

	for _, protocol := range []string{"tcp", " TCP ", "uDp"} {
		err := validateDialPlatformFor("darwin", protocol)
		assert.Error(t, err, "protocol=%q", protocol)
		assert.Contains(t, err.Error(), "unsupported")
	}
	assert.NoError(t, validateDialPlatformFor("darwin", " ICMP "))
}

func TestRunDialProbeCancelledContext(t *testing.T) {
	// When the platform is unsupported RunDialProbe returns a fast
	// unsupported_platform error without touching the probe runner.
	if !netpathSupportedOS(runtime.GOOS) {
		res := RunDialProbe(context.Background(), dt.NetPathProbeConfig{
			Host: "netpath.invalid", Protocol: "icmp",
			Timeout: time.Second, MaxTTL: 3,
			TracerouteQueries: 1, E2EQueries: 1,
		}, nil)
		assert.Error(t, res.Err)
		assert.Equal(t, "unsupported_platform", res.FailType)
		return
	}

	// On supported platforms a pre-canceled context must short-circuit the
	// probe instead of running until the internal ~5m deadline. A non-IP
	// host keeps enrichProbeGatewayWithLookup from doing a gateway lookup.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := dt.NetPathProbeConfig{
		Host: "netpath.invalid", Protocol: "icmp",
		Timeout: time.Second, MaxTTL: 3,
		TracerouteQueries: 1, E2EQueries: 1,
	}
	start := time.Now()
	res := RunDialProbe(ctx, cfg, nil)
	elapsed := time.Since(start)

	assert.Less(t, elapsed, time.Second)
	assert.Error(t, res.Err)
	assert.ErrorIs(t, res.Err, context.Canceled)
}

func TestValidateResolvedTaskIPUsesDialtestingValidator(t *testing.T) {
	called := 0
	expected := errors.New("internal destination")
	probeTask := task{
		resolvedIPValidator: func(ip net.IP) error {
			called++
			assert.Equal(t, "10.0.0.1", ip.String())
			return expected
		},
	}

	err := validateResolvedTaskIP(probeTask, "example.com", net.ParseIP("10.0.0.1"))
	require.Error(t, err)
	assert.ErrorIs(t, err, expected)
	assert.Equal(t, 1, called)
}
