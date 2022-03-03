package version

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestParse(t *testing.T) {
	cases := []struct {
		v    string
		fail bool
	}{
		{
			v: "1.2.3",
		},

		{
			v: "1.2.3-123-g123abcde",
		},
	}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			vi := VerInfo{VersionString: tc.v}
			err := vi.Parse()
			if tc.fail {
				tu.NotOk(t, err, "")
				return
			} else {
				tu.Ok(t, err)
			}

			t.Logf(vi.String())
		})
	}
}

//nolint:funlen
func TestCompare(t *testing.T) {
	cases := []struct {
		id          string
		v1, v2      string
		newVersion  bool
		sameVersion bool
		fail        bool
	}{
		{
			id:         "version-with-tag",
			v1:         "1.1.7-rc7.1_foo-bar",
			v2:         "1.1.7-rc7",
			newVersion: true,
		},

		{
			id:         "hotfix-version",
			v1:         "1.1.7-rc7.1",
			v2:         "1.1.7-rc7",
			newVersion: true,
		},

		{
			id:         "among-rc-versions",
			v1:         "1.1.7-rc7",
			v2:         "1.1.7-rc0",
			newVersion: true,
		},

		{
			id:         "among-rc-testing-versions",
			v1:         "1.1.7-rc6.1-125-g40c4860c",
			v2:         "1.1.7-rc6-125-g40c4860c",
			newVersion: true,
		},
		{
			id: "same-testing-version",
			v1: "1.1.7-rc1-125-g40c4860c",
			v2: "1.1.7-rc2-125-g40c4860c",
		},
		{
			id:         "different-testing-version",
			v1:         "1.1.7-rc1-126-g40c4860c",
			v2:         "1.1.7-rc1-125-g40c4860c",
			newVersion: true,
		},

		{
			id:          "same-version-with-diff-tag",
			v1:          "1.1.7-rc1-126-g40c4860c_bar",
			v2:          "1.1.7-rc1-126-g40c4860c_foo",
			sameVersion: true,
		},

		{
			id:         "underscore-tag-name",
			v1:         "1.1.7-rc1-126-g40c4860c__",
			v2:         "1.1.7-rc1",
			newVersion: true,
		},

		{
			id:         "long-tag-name-with-multi-underscore",
			v1:         "1.1.7-rc1-126-g40c4860c_tag_with_unerscore",
			v2:         "1.1.7-rc1",
			newVersion: true,
		},

		{
			id:         "testing-RC-version",
			v1:         "1.1.7-rc1-126-g40c4860c",
			v2:         "1.1.7-rc1",
			newVersion: true,
		},

		{
			id: "among-rc-versions",
			v1: "1.1.7-rc1-126-g40c4860c",
			v2: "1.1.7-rc2",
		},

		{
			id:         "among-master-versions",
			v1:         "2.1.7-rc1-126-g40c4860c",
			v2:         "1.1.7-rc2",
			newVersion: true,
		},

		{
			id:   "invalid-version",
			v1:   "2.1.7.0-rc1-126-g40c4860c", // invalid version string
			fail: true,
		},

		{
			id:          "seems-like-invalid-version-but-ok",
			v1:          "2.1.7_rc1-126-g40c4860c_foo-bar",
			v2:          "2.1.7__rc1-126-g40c4860c_foo-bar",
			sameVersion: true,
		},

		{
			id:         "among-master-versions",
			v1:         "2.1.0-rc1-126-g40c4860c",
			v2:         "1.1.7-rc2",
			newVersion: true,
		},

		{
			id:         "among-master-versions",
			v1:         "2.1.0-rc1-126-g40c4860c",
			v2:         "1.1.999-rc2",
			newVersion: true,
		},

		{
			id:         "among-biggest-minor-versions",
			v1:         "1.1.0-rc1-126-g40c4860c",
			v2:         "1.0.1024-rc2",
			newVersion: true,
		},

		{
			id:         "among-biggest-minimal-versions",
			v1:         "2.20.0-rc1-126-g40c4860c",
			v2:         "2.17.1024-rc2",
			newVersion: true,
		},

		{
			id:         "among-biggest-minor-minimal-versions",
			v1:         "10.1024.0-rc1-126-g40c4860c",
			v2:         "9.1024.1024-rc2",
			newVersion: true,
		},

		{
			id:   "invalid-minor-version",
			v1:   "2.17.1025-rc2",
			fail: true,
		},

		{
			id:   "invalid-version",
			v1:   "2.17.-1-rc2",
			fail: true,
		},
	}

	var err error
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			v1 := &VerInfo{VersionString: tc.v1}
			v2 := &VerInfo{VersionString: tc.v2}

			err = v1.Parse()
			if tc.fail {
				tu.NotOk(t, err, "expect error, but got %+#v", v1)
				t.Logf("expect error: %s", err)
				return
			} else {
				tu.Ok(t, err)
			}

			err = v2.Parse()
			if tc.fail {
				tu.NotOk(t, err, "")
				t.Log(err)
				return
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
		})
	}
}
