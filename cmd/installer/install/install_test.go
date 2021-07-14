package install

import (
	"runtime"
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

func TestPreEnableHostobjectInput(t *testing.T) {
	for _, cloud := range []string{
		"aws", "tencent", "aliyun", "unknown", "",
	} {
		conf, err := preEnableHostobjectInput(cloud)
		if err != nil {
			t.Fatal()
		}

		t.Logf("conf:\n%s", string(conf))
	}
}

func TestUpgradeMainConfigure(t *testing.T) {
	cases := []struct {
		c, expect *config.Config
		os        string
		arch      string
	}{
		{
			c:      &config.Config{Log: "a/b/c", GinLog: "/d/e/f"},
			expect: &config.Config{Log: "/var/log/datakit/log", GinLog: "/var/log/datakit/gin.log"},
		},

		{
			c: &config.Config{Log: "a/b/c", GinLog: "/d/e/f"},
			expect: &config.Config{Log: `C:\Program Files\datakit\log`,
				GinLog: `C:\Program Files\datakit\gin.log`},
			os: "windows",
		},

		{
			c: &config.Config{Log: "a/b/c", GinLog: "/d/e/f"},
			expect: &config.Config{Log: `C:\Program Files (x86)\datakit\log`,
				GinLog: `C:\Program Files (x86)\datakit\gin.log`},
			os:   "windows",
			arch: "386",
		},
	}

	for _, tc := range cases {
		if tc.os == "" || runtime.GOOS == tc.os {
			c, _ := upgradeMainConfig(tc.c)
			tu.Equals(t, c.String(), tc.expect.String())
		}
	}
}
