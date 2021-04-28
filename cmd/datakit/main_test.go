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
			newVersion: &datakitVerInfo{VersionString: "v1.1.2", Commit: "12345"},
			curVersion: &datakitVerInfo{VersionString: "v1.1.1", Commit: "12345"},
			acceptRC:   false,
			isNewVer:   true,
		},

		{
			newVersion: &datakitVerInfo{VersionString: "v1.1.2-rc0", Commit: "12345"},
			curVersion: &datakitVerInfo{VersionString: "v1.1.1", Commit: "12345"},
			acceptRC:   false,
			isNewVer:   false,
		},

		{
			newVersion: &datakitVerInfo{VersionString: "v1.1.2-rc0", Commit: "12345"},
			curVersion: &datakitVerInfo{VersionString: "v1.1.1", Commit: "12345"},
			acceptRC:   true,
			isNewVer:   true,
		},

		{
			newVersion: &datakitVerInfo{VersionString: "v1.1.1", Commit: "12345"},
			curVersion: &datakitVerInfo{VersionString: "v1.1.1", Commit: "12345"},
			acceptRC:   true,
			isNewVer:   false,
		},

		{
			newVersion: &datakitVerInfo{VersionString: "v1.1.1-rc0", Commit: "12345"},
			curVersion: &datakitVerInfo{VersionString: "v1.1.1", Commit: "12345"},
			acceptRC:   true,
			isNewVer:   false,
		},

		{
			newVersion: &datakitVerInfo{VersionString: "1.1.5-rc1", Commit: "12345"},
			curVersion: &datakitVerInfo{VersionString: "1.1.5-rc0-1-g5d960738", Commit: "12345"},
			acceptRC:   true,
			isNewVer:   true,
		},
	}

	for _, tc := range cases {

		if err := tc.newVersion.parse(); err != nil {
			t.Error(err)
		}
		if err := tc.curVersion.parse(); err != nil {
			t.Error(err)
		}

		ok := isNewVersion(tc.newVersion, tc.curVersion, tc.acceptRC)
		if tc.isNewVer {
			tu.Assert(t, ok == true, "%s expect to be new version, current version is %s", tc.newVersion, tc.curVersion)
		} else {
			tu.Assert(t, ok == false, "%s expect to be not new version, current version is %s", tc.newVersion, tc.curVersion)
		}
	}
}
