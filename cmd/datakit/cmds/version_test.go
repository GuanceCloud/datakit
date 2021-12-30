package cmds

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestShowVersion(t *testing.T) {
	cases := []struct {
		verstr, id        string
		showTesting, fail bool
		expect            map[string]*newVersionInfo
	}{
		{
			id:          `base`,
			verstr:      `1.1.1`,
			showTesting: true,
			expect: map[string]*newVersionInfo{
				versionTypeOnline: {
					versionType: versionTypeOnline,
					install:     true,
				},
				versionTypeTesting: {
					versionType: versionTypeTesting,
					install:     true,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			vis, err := checkNewVersion(tc.verstr, tc.showTesting)
			if tc.fail {
				tu.NotOk(t, err, "")
				t.Logf("expect fail: %s", err)
			} else {
				tu.Ok(t, err)
			}

			for k, vi := range vis {
				t.Logf("\n%s\n%s", vi.String(), tc.expect[k].String())
				tu.Equals(t, tc.expect[k].String(), vi.String())
			}
		})
	}
}
