package main

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestNewVersionAvailable(t *testing.T) {
	cases := []struct {
		newVersion *datakitVerInfo
		curVersion *datakitVerInfo
		isNewVer   bool
		acceptRC   bool
	}{
		{
			newVersion: &datakitVerInfo{Version: "v1.1.2", Commit: "12345"},
			curVersion: &datakitVerInfo{Version: "v1.1.1", Commit: "12345"},
			acceptRC:   false,
			isNewVer:   true,
		},

		{
			newVersion: &datakitVerInfo{Version: "v1.1.2-rc0", Commit: "12345"},
			curVersion: &datakitVerInfo{Version: "v1.1.1", Commit: "12345"},
			acceptRC:   false,
			isNewVer:   false,
		},

		{
			newVersion: &datakitVerInfo{Version: "v1.1.2-rc0", Commit: "12345"},
			curVersion: &datakitVerInfo{Version: "v1.1.1", Commit: "12345"},
			acceptRC:   true,
			isNewVer:   true,
		},

		{
			newVersion: &datakitVerInfo{Version: "v1.1.1", Commit: "12345"},
			curVersion: &datakitVerInfo{Version: "v1.1.1", Commit: "12345"},
			acceptRC:   true,
			isNewVer:   false,
		},

		{
			newVersion: &datakitVerInfo{Version: "v1.1.1-rc0", Commit: "12345"},
			curVersion: &datakitVerInfo{Version: "v1.1.1", Commit: "12345"},
			acceptRC:   true,
			isNewVer:   false,
		},
	}

	for _, tc := range cases {
		ok := newVersionAvailable(tc.newVersion, tc.curVersion, tc.acceptRC)
		if tc.isNewVer {
			tu.Assert(t, ok == true, "%s expect to be new version, current version is %s", tc.newVersion, tc.curVersion)
		} else {
			tu.Assert(t, ok == false, "%s expect to be not new version, current version is %s", tc.newVersion, tc.curVersion)
		}
	}
}
