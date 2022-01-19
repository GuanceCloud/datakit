package build

import (
	"fmt"
	"path"
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestAddOSSFiles(t *testing.T) {
	cases := []struct {
		id            string
		ossPath       string
		files, expect map[string]string
	}{
		{
			ossPath: "abc",
			expect: map[string]string{
				"abc/1": "1",
				"abc/2": "2",
				"abc/3": "3",
			},
			files: map[string]string{
				"1": "1",
				"2": "2",
				"3": "3",
			},
		},

		{
			ossPath: "datakit/rc",
			expect: map[string]string{
				"datakit/rc/version":                               path.Join("pub", "release", "version"),
				"datakit/rc/datakit.yaml":                          "datakit.yaml",
				"datakit/rc/install.sh":                            "install.sh",
				"datakit/rc/install.ps1":                           "install.ps1",
				fmt.Sprintf("datakit/rc/datakit-%s.yaml", "1.1.3"): "datakit.yaml",
				fmt.Sprintf("datakit/rc/install-%s.sh", "1.1.3"):   "install.sh",
				fmt.Sprintf("datakit/rc/install-%s.ps1", "1.1.3"):  "install.ps1",
			},
			files: map[string]string{
				"version":                               path.Join("pub", "release", "version"),
				"datakit.yaml":                          "datakit.yaml",
				"install.sh":                            "install.sh",
				"install.ps1":                           "install.ps1",
				fmt.Sprintf("datakit-%s.yaml", "1.1.3"): "datakit.yaml",
				fmt.Sprintf("install-%s.sh", "1.1.3"):   "install.sh",
				fmt.Sprintf("install-%s.ps1", "1.1.3"):  "install.ps1",
			},
		},

		{
			ossPath: "datakit",
			expect: map[string]string{
				"datakit/version":      path.Join("pub", "release", "version"),
				"datakit/datakit.yaml": "datakit.yaml",
				"datakit/install.sh":   "install.sh",
				"datakit/install.ps1":  "install.ps1",
				fmt.Sprintf("datakit/datakit-%s.yaml", "1.1024.3"): "datakit.yaml",
				fmt.Sprintf("datakit/install-%s.sh", "1.1024.3"):   "install.sh",
				fmt.Sprintf("datakit/install-%s.ps1", "1.1024.3"):  "install.ps1",
			},
			files: map[string]string{
				"version":      path.Join("pub", "release", "version"),
				"datakit.yaml": "datakit.yaml",
				"install.sh":   "install.sh",
				"install.ps1":  "install.ps1",
				fmt.Sprintf("datakit-%s.yaml", "1.1024.3"): "datakit.yaml",
				fmt.Sprintf("install-%s.sh", "1.1024.3"):   "install.sh",
				fmt.Sprintf("install-%s.ps1", "1.1024.3"):  "install.ps1",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			files := addOSSFiles(tc.ossPath, tc.files)
			for k, v := range files {
				tu.Equals(t, tc.expect[k], v)
				t.Logf("%s -> %s", v, k)
			}
		})
	}
}
