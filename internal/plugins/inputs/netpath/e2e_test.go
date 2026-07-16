// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

import (
	"context"
	"errors"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunTCPE2ECountsRefusalAsResponse(t *testing.T) {
	previous := dialE2ETCP
	dialE2ETCP = func(_ context.Context, address string) (net.Conn, error) {
		assert.Equal(t, "203.0.113.10:443", address)
		return nil, syscall.ECONNREFUSED
	}
	t.Cleanup(func() { dialE2ETCP = previous })

	result := runTCPE2E(task{
		Target:     "203.0.113.10",
		Port:       443,
		Protocol:   protocolTCP,
		Timeout:    time.Second,
		E2EQueries: 2,
	})
	tags := map[string]string{}
	fields := map[string]interface{}{}
	mergeE2EResult(tags, fields, result, 2)

	assert.NoError(t, result.err)
	assert.Equal(t, "reached", tags["e2e_status"])
	assert.Equal(t, "203.0.113.10", fields["e2e_dest_ip"])
	assert.Equal(t, int64(2), fields["e2e_packets_received"])
	assert.Equal(t, int64(2), fields["e2e_tcp_connection_refused"])
	assert.Equal(t, float64(0), fields["e2e_probe_loss_percent"])
}

func TestMergeE2EResultUsesSendOrderForVariation(t *testing.T) {
	tags := map[string]string{}
	fields := map[string]interface{}{}
	mergeE2EResult(tags, fields, e2eResult{
		protocol: protocolTCP,
		destIP:   "203.0.113.10",
		samples: []e2eSample{
			{sequence: 3, sent: true, received: true, rtt: 4 * time.Millisecond},
			{sequence: 0, sent: true, received: true, rtt: time.Millisecond},
			{sequence: 2, sent: true},
			{sequence: 1, sent: true, received: true, rtt: 3 * time.Millisecond, refused: true},
		},
	}, 4)

	assert.Equal(t, "partial", tags["e2e_status"])
	assert.Equal(t, "203.0.113.10", tags["probe_dest_ip"])
	assert.Equal(t, "203.0.113.10", fields["e2e_dest_ip"])
	assert.Equal(t, int64(4), fields["e2e_packets_sent"])
	assert.Equal(t, int64(3), fields["e2e_packets_received"])
	assert.Equal(t, float64(25), fields["e2e_probe_loss_percent"])
	assert.Equal(t, int64(1), fields["e2e_tcp_connection_refused"])
	assert.Equal(t, float64(8000.0/3.0), fields["e2e_rtt_avg"])
	assert.Equal(t, int64(1), fields["e2e_rtt_variation_samples"])
	assert.Equal(t, float64(2000), fields["e2e_rtt_variation_avg"])
}

func TestMergeE2EResultDoesNotTreatUDPSilenceAsLoss(t *testing.T) {
	tags := map[string]string{}
	fields := map[string]interface{}{}
	mergeE2EResult(tags, fields, e2eResult{
		protocol: protocolUDP,
		samples: []e2eSample{
			{sequence: 0, sent: true, unknown: true},
			{sequence: 1, sent: true, unknown: true},
		},
	}, 2)

	assert.Equal(t, "unknown", tags["e2e_status"])
	assert.Equal(t, int64(2), fields["e2e_unknown"])
	assert.NotContains(t, fields, "e2e_probe_loss_percent")
}

func TestMergeE2EResultPreservesIndependentDestination(t *testing.T) {
	tags := map[string]string{"probe_dest_ip": "203.0.113.20"}
	fields := map[string]interface{}{}
	mergeE2EResult(tags, fields, e2eResult{
		protocol: protocolTCP,
		destIP:   "203.0.113.10",
	}, 1)

	assert.Equal(t, "203.0.113.20", tags["probe_dest_ip"])
	assert.Equal(t, "203.0.113.10", fields["e2e_dest_ip"])
}

func TestRunProbeKeepsE2EIndependentFromPathFailure(t *testing.T) {
	previous := runE2EProbe
	runE2EProbe = func(task) e2eResult {
		return e2eResult{
			protocol: protocolICMP,
			samples:  []e2eSample{{sequence: 0, sent: true, received: true, rtt: time.Millisecond}},
		}
	}
	t.Cleanup(func() { runE2EProbe = previous })

	result := runProbe(task{
		ID:         "task-1",
		Protocol:   "unsupported",
		E2EQueries: 1,
	})

	assert.Error(t, result.err)
	assert.Equal(t, "reached", result.tags["e2e_status"])
	assert.Equal(t, int64(1), result.fields["e2e_packets_received"])
	assert.NotContains(t, result.fields, "e2e_success")
}

func TestRunProbeFinalizesDurationAndPointTime(t *testing.T) {
	result := runProbe(task{ID: "task-duration", Protocol: "unsupported"})

	require.Error(t, result.err)
	assert.False(t, result.finishedAt.IsZero())
	assert.GreaterOrEqual(t, result.duration, time.Duration(0))
	assert.Equal(t, result.finishedAt.Sub(result.startedAt), result.duration)
	points := (&Input{}).pointsForResult(result)
	require.Len(t, points, 1)
	assert.True(t, result.finishedAt.Equal(points[0].Time()))
}

func TestMergeE2EResultReportsSetupFailure(t *testing.T) {
	tags := map[string]string{}
	fields := map[string]interface{}{}
	mergeE2EResult(tags, fields, e2eResult{
		protocol: protocolICMP,
		err:      errors.New("raw socket denied"),
	}, 3)

	assert.Equal(t, "failed", tags["e2e_status"])
	assert.Equal(t, "raw socket denied", fields["e2e_fail_reason"])
}
