// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sensitive

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedactBugReportString(t *testing.T) {
	input := `dataway=https://openway.example.com?token=tkn_secret_123
password = "secret-pass"
args = '--password', 'secret-arg',
url = "http://user:secret@example.com"`

	output := RedactBugReportString(input, []string{"dataway", "password", "uri"})

	require.Contains(t, output, "token=******")
	require.Contains(t, output, `password = "******"`)
	require.Contains(t, output, "'--password', '******',")
	require.Contains(t, output, `http://user:******@example.com`)
	require.NotContains(t, output, "tkn_secret_123")
	require.NotContains(t, output, "secret-pass")
	require.NotContains(t, output, "secret-arg")
}

func TestIsSensitiveKey(t *testing.T) {
	require.True(t, IsSensitiveKey("ENV_DATAWAY_TOKEN"))
	require.True(t, IsSensitiveKey("OSS_ACCESS_KEY_SECRET"))
	require.False(t, IsSensitiveKey("ENV_HOSTNAME"))
}

func TestRedactLogString(t *testing.T) {
	input := `dataway=https://openway.example.com?token=secret-token kv=real-value OSS_ACCESS_KEY_ID=oss-key OSS_ACCESS_KEY_SECRET=oss-secret auth:'real-auth'`

	output := RedactLogString(input)

	require.Contains(t, output, "token=******")
	require.Contains(t, output, "kv=******")
	require.Contains(t, output, "OSS_ACCESS_KEY_ID=******")
	require.Contains(t, output, "OSS_ACCESS_KEY_SECRET=******")
	require.Contains(t, output, "auth:'******'")
	require.NotContains(t, output, "secret-token")
	require.NotContains(t, output, "real-value")
	require.NotContains(t, output, "oss-key")
	require.NotContains(t, output, "oss-secret")
	require.NotContains(t, output, "real-auth")
}

func TestRedactLogStringDatawayURLInError(t *testing.T) {
	input := `Do: Post "http://10.100.64.195:9528/v1/write/metric?token=tkn_forethoughtlocal0123456789abcdef": dial tcp 10.100.64.195:9528: connect: connection refused`

	output := RedactLogString(input)

	require.Contains(t, output, "token=******")
	require.NotContains(t, output, "tkn_forethoughtlocal0123456789abcdef")
}
