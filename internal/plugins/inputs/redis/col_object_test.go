// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseInfoObject(t *testing.T) {
	info := `
# Server
redis_version:7.2.4
uptime_in_seconds:123
# Stats
instantaneous_ops_per_sec:42
`

	version, uptime, qps, err := parseInfoObject(info)
	require.NoError(t, err)
	assert.Equal(t, "7.2.4", version)
	assert.Equal(t, int64(123), uptime)
	assert.Equal(t, float64(42), qps)
}

func TestSanitizeObjectSettings(t *testing.T) {
	settings := sanitizeObjectSettings(map[string]string{
		"maxmemory":                "1048576",
		"requirepass":              "secret",
		"masterauth":               "",
		"tls-key-file-pass":        "tls-secret",
		"tls-client-key-file-pass": "client-secret",
	})

	assert.Equal(t, "1048576", settings["maxmemory"])
	assert.Equal(t, CREDENTIALSTR, settings["requirepass"])
	assert.Equal(t, notSet, settings["masterauth"])
	assert.Equal(t, CREDENTIALSTR, settings["tls-key-file-pass"])
	assert.Equal(t, CREDENTIALSTR, settings["tls-client-key-file-pass"])
}
