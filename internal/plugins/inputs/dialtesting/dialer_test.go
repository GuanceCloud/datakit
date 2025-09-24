// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCronTicker(t *testing.T) {
	tests := []struct {
		name      string
		crontab   string
		wantErr   bool
		errString string
	}{
		{
			name:    "valid standard crontab",
			crontab: "*/1 * * * *",
			wantErr: false,
		},
		{
			name:    "valid every minute",
			crontab: "0 * * * *",
			wantErr: false,
		},
		{
			name:    "valid every 5 minutes",
			crontab: "*/5 * * * *",
			wantErr: false,
		},
		{
			name:      "invalid crontab format",
			crontab:   "invalid",
			wantErr:   true,
			errString: "expected exactly 5 fields",
		},
		{
			name:      "empty crontab",
			crontab:   "",
			wantErr:   true,
			errString: "empty spec string",
		},
		{
			name:    "valid hourly",
			crontab: "0 0 * * *",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct, err := newCronTicker(tt.crontab)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, ct)
			assert.NotNil(t, ct.C)
			assert.NotNil(t, ct.cron)

			// Clean up
			ct.Stop()
		})
	}
}

func TestCronTickerStop(t *testing.T) {
	ct, err := newCronTicker("*/1 * * * *")
	assert.NoError(t, err)

	// Stop the ticker
	ct.Stop()

	// After stop, channel should be closed
	_, ok := <-ct.C
	assert.False(t, ok, "channel should be closed after Stop")
}
