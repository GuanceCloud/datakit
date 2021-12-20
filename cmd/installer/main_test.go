package main

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestCheckUpgradeVersion(t *testing.T) {
	cases := []struct {
		id, s string
		fail  bool
	}{
		{
			id: "normal",
			s:  "1.2.3",
		},
		{
			id: "zero-minor-version",
			s:  "1.0.3",
		},

		{
			id: "large minor version",
			s:  "1.1024.3",
		},
		{
			id:   `too-large-minor-version`,
			s:    "1.1026.3",
			fail: true,
		},
		{
			id:   `unstable-version`,
			s:    "1.3.3",
			fail: true,
		},

		{
			id:   `unstable-rc-version`,
			s:    "1.1.9-rc1",
			fail: true,
		},

		{
			id:   `unstable-rc-testing-version`,
			s:    "1.1.7-rc1-125-g40c4860c",
			fail: true,
		},

		{
			id:   `unstable-rc-hotfix-version`,
			s:    "1.1.7-rc7.1",
			fail: true,
		},

		{
			id:   `invalid-version-string`,
			s:    "2.1.7.0-rc1-126-g40c4860c",
			fail: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			err := checkUpgradeVersion(tc.s)
			if tc.fail {
				tu.NotOk(t, err, "")
				t.Logf("expect error: %s -> %s", tc.s, err)
			} else {
				tu.Ok(t, err)
			}
		})
	}
}
