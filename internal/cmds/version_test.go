// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	T "testing"

	"github.com/GuanceCloud/cliutils/testutil"
	"github.com/stretchr/testify/assert"
)

func TestCanonicalInstallBaseUrl(t *T.T) {
	type testCase struct {
		name     string
		param    string
		expected string
	}

	testCases := []testCase{
		{
			name:     "no-slash-no-datakit",
			param:    "https://static.guance.com",
			expected: "https://static.guance.com/datakit/",
		},
		{
			name:     "with-slash-no-datakit",
			param:    "https://static.guance.com/",
			expected: "https://static.guance.com/datakit/",
		},

		{
			name:     "with-datakit-no-slash",
			param:    "https://static.guance.com/datakit",
			expected: "https://static.guance.com/datakit/",
		},
		{
			name:     "with-datakit-with-slash",
			param:    "https://static.guance.com/datakit/",
			expected: "https://static.guance.com/datakit/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *T.T) {
			testutil.Equals(t, tc.expected, CanonicalInstallBaseURL(tc.param))
		})
	}
}

func TestUpgradeCommand(t *T.T) {
	t.Run("win-upgrade", func(t *T.T) {
		s := getUpgradeCommand("windows", "https://static.guance.com/datakit", "")

		t.Logf("\n%s", s)

		expect := `    Remove-Item -ErrorAction SilentlyContinue Env:DK_*;
    $env:DK_UPGRADE="1";
    Set-ExecutionPolicy Bypass -scope Process -Force;
    Import-Module bitstransfer;
    start-bitstransfer  -source https://static.guance.com/datakit/install.ps1 -destination .install.ps1;
    powershell .install.ps1;`

		assert.Equal(t, expect, s)
	})

	t.Run("win-upgrade-with-proxy", func(t *T.T) {
		s := getUpgradeCommand("windows", "https://static.guance.com/datakit", "1.2.3.4:80")

		t.Logf("\n%s", s)

		expect := `    Remove-Item -ErrorAction SilentlyContinue Env:DK_*;
    $env:DK_UPGRADE="1";
    $env:HTTPS_PROXY="1.2.3.4:80";
    Set-ExecutionPolicy Bypass -scope Process -Force;
    Import-Module bitstransfer;
    start-bitstransfer -ProxyUsage Override -ProxyList $env:HTTPS_PROXY -source https://static.guance.com/datakit/install.ps1 -destination .install.ps1;
    powershell .install.ps1;`

		assert.Equal(t, expect, s)
	})

	t.Run("unix-upgrade", func(t *T.T) {
		s := getUpgradeCommand("linux", "https://static.guance.com/datakit", "")
		t.Logf("\n%s", s)
		expect := `    DK_UPGRADE=1 bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"`
		assert.Equal(t, expect, s)
	})

	t.Run("unix-upgrade-with-proxy", func(t *T.T) {
		s := getUpgradeCommand("linux", "https://static.guance.com/datakit", "1.2.3.4:80")
		t.Logf("\n%s", s)
		expect := `    DK_UPGRADE=1 HTTPS_PROXY=1.2.3.4:80 bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"`
		assert.Equal(t, expect, s)
	})
}
