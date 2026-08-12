// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package w32time

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestDefaultInput(t *testing.T) {
	ipt := newDefaultInput()
	require.NotNil(t, ipt)
	assert.Equal(t, time.Minute, ipt.Interval)
	assert.NotNil(t, ipt.Tags)
	assert.Equal(t, []string{datakit.OSLabelWindows}, ipt.AvailableArchs())
	assert.Len(t, ipt.SampleMeasurement(), 1)
}

func TestNormalizeCounterValues(t *testing.T) {
	raw := map[string]float64{
		fieldComputedTimeOffset:   1_500_000,
		fieldNTPRoundtripDelay:    250_000,
		fieldNTPClientSourceCount: 2,
	}

	fields := normalizeCounterValues(raw)
	assert.Equal(t, 1.5, fields[fieldComputedTimeOffset])
	assert.Equal(t, 0.25, fields[fieldNTPRoundtripDelay])
	assert.Equal(t, int64(2), fields[fieldNTPClientSourceCount])
}

func TestNormalizeCounterValuesKeepsPartialData(t *testing.T) {
	fields := normalizeCounterValues(map[string]float64{
		fieldComputedTimeOffset: 500_000,
	})

	assert.Equal(t, map[string]interface{}{
		fieldComputedTimeOffset: 0.5,
	}, fields)
}

func TestCollectPoint(t *testing.T) {
	partialErr := errors.New("one counter unavailable")
	serviceErr := errors.New("access denied")

	tests := []struct {
		name             string
		running          bool
		serviceErr       error
		raw              map[string]float64
		counterErr       error
		wantPoint        bool
		wantErr          string
		wantCounterCalls int
		wantFields       map[string]interface{}
	}{
		{
			name:       "service query fails",
			serviceErr: serviceErr,
			wantErr:    "query W32Time service: access denied",
		},
		{
			name:      "service stopped",
			wantPoint: true,
			wantFields: map[string]interface{}{
				fieldServiceRunning: int64(0),
			},
		},
		{
			name:             "service running",
			running:          true,
			wantPoint:        true,
			wantCounterCalls: 1,
			raw: map[string]float64{
				fieldComputedTimeOffset:   1_500_000,
				fieldNTPRoundtripDelay:    250_000,
				fieldNTPClientSourceCount: 2,
			},
			wantFields: map[string]interface{}{
				fieldServiceRunning:       int64(1),
				fieldComputedTimeOffset:   1.5,
				fieldNTPRoundtripDelay:    0.25,
				fieldNTPClientSourceCount: int64(2),
			},
		},
		{
			name:             "partial counter data",
			running:          true,
			wantPoint:        true,
			wantErr:          partialErr.Error(),
			wantCounterCalls: 1,
			counterErr:       partialErr,
			raw: map[string]float64{
				fieldComputedTimeOffset: 500_000,
			},
			wantFields: map[string]interface{}{
				fieldServiceRunning:     int64(1),
				fieldComputedTimeOffset: 0.5,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ipt := newDefaultInput()
			ipt.mergedTags = map[string]string{
				"host": "windows-host",
				"env":  "test",
			}

			counterCalls := 0
			pt, err := ipt.collectPoint(123,
				func() (bool, error) { return tc.running, tc.serviceErr },
				func() (map[string]float64, error) {
					counterCalls++
					return tc.raw, tc.counterErr
				},
			)

			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.wantErr)
			}
			assert.Equal(t, tc.wantCounterCalls, counterCalls)

			if !tc.wantPoint {
				assert.Nil(t, pt)
				return
			}

			require.NotNil(t, pt)
			assert.Equal(t, "windows-host", pt.GetTag("host"))
			assert.Equal(t, "test", pt.GetTag("env"))
			for field, want := range tc.wantFields {
				assert.Equal(t, want, pt.Get(field))
			}
		})
	}
}

func TestTerminate(t *testing.T) {
	ipt := newDefaultInput()
	ipt.Terminate()

	select {
	case <-ipt.semStop.Wait():
	case <-time.After(time.Second):
		t.Fatal("Terminate did not close the stop semaphore")
	}
}
