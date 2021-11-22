package version

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

//nolint:funlen
func TestCompare(t *testing.T) {
	cases := []struct {
		v1, v2      string
		newVersion  bool
		sameVersion bool
		fail        bool
	}{
		{
			v1:         "1.1.7-rc7.1_foo-bar", // with tags
			v2:         "1.1.7-rc7",
			newVersion: true,
		},

		{
			v1:         "1.1.7-rc7.1",
			v2:         "1.1.7-rc7",
			newVersion: true,
		},

		{
			v1:         "1.1.7-rc7",
			v2:         "1.1.7-rc0",
			newVersion: true,
		},

		{
			v1:         "1.1.7-rc6.1-125-g40c4860c",
			v2:         "1.1.7-rc6-125-g40c4860c",
			newVersion: true,
		},
		{
			v1: "1.1.7-rc1-125-g40c4860c",
			v2: "1.1.7-rc2-125-g40c4860c",
		},
		{
			v1:         "1.1.7-rc1-126-g40c4860c",
			v2:         "1.1.7-rc1-125-g40c4860c",
			newVersion: true,
		},

		{
			v1:          "1.1.7-rc1-126-g40c4860c",
			v2:          "1.1.7-rc1-126-g40c4860c",
			sameVersion: true,
		},

		{ // version tag not comapred
			v1:          "1.1.7-rc1-126-g40c4860c_bar",
			v2:          "1.1.7-rc1-126-g40c4860c_foo",
			sameVersion: true,
		},

		{
			v1:         "1.1.7-rc1-126-g40c4860c__", // invalid tag, ignored
			v2:         "1.1.7-rc1",
			newVersion: true,
		},

		{
			v1:         "1.1.7-rc1-126-g40c4860c_tag_with_unerscore",
			v2:         "1.1.7-rc1",
			newVersion: true,
		},

		{
			v1:         "1.1.7-rc1-126-g40c4860c",
			v2:         "1.1.7-rc1",
			newVersion: true,
		},

		{
			v1: "1.1.7-rc1-126-g40c4860c",
			v2: "1.1.7-rc2",
		},

		{
			v1:         "2.1.7-rc1-126-g40c4860c",
			v2:         "1.1.7-rc2",
			newVersion: true,
		},

		{
			v1:          "1.2.7-rc1-126-g40c4860c",
			v2:          "1.2.7-rc1-126-g40c4860c",
			sameVersion: true,
		},

		{
			v1:   "2.1.7.0-rc1-126-g40c4860c", // invalid version string
			fail: true,
		},

		{
			v1:   "2.1.7_rc1-126-g40c4860c_foo-bar", // invalid version string
			fail: true,
		},

		{
			v1:         "2.1.0-rc1-126-g40c4860c",
			v2:         "1.1.7-rc2",
			newVersion: true,
		},

		{
			v1:         "2.1.0-rc1-126-g40c4860c",
			v2:         "1.1.999-rc2",
			newVersion: true,
		},

		{
			v1:         "1.1.0-rc1-126-g40c4860c",
			v2:         "1.0.1024-rc2",
			newVersion: true,
		},

		{
			v1:         "2.20.0-rc1-126-g40c4860c",
			v2:         "2.17.1024-rc2",
			newVersion: true,
		},

		{
			v1:         "10.1024.0-rc1-126-g40c4860c",
			v2:         "9.1024.1024-rc2",
			newVersion: true,
		},

		{
			v1:   "2.17.1025-rc2",
			fail: true,
		},

		{
			v1:   "2.17.-1-rc2",
			fail: true,
		},
	}

	var err error
	for _, tc := range cases {
		v1 := &VerInfo{VersionString: tc.v1}
		v2 := &VerInfo{VersionString: tc.v2}

		err = v1.Parse()
		if tc.fail {
			tu.NotOk(t, err, "")
			t.Log(err)
			continue
		} else {
			tu.Ok(t, err)
		}

		err = v2.Parse()
		if tc.fail {
			tu.NotOk(t, err, "")
			t.Log(err)
			continue
		} else {
			tu.Ok(t, err)
		}

		if tc.newVersion {
			tu.Assert(t, v1.Compare(v2) > 0, "%s should larger than %s", tc.v1, tc.v2)
		} else {
			if tc.sameVersion {
				tu.Assert(t, v1.Compare(v2) == 0, "")
			} else {
				tu.Assert(t, v1.Compare(v2) < 0, "")
			}
		}
	}
}
