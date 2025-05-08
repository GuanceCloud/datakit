// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddOSSFiles(t *testing.T) {
	cases := []struct {
		name          string
		ossPath       string
		files, expect []ossFile
	}{
		{
			name:    `basic`,
			ossPath: "abc",
			expect: []ossFile{
				{"abc/1", "1"},
				{"abc/2", "2"},
				{"abc/3", "3"},
			},
			files: []ossFile{
				{"1", "1"},
				{"2", "2"},
				{"3", "3"},
			},
		},

		{
			name:    `path-rc`,
			ossPath: "datakit/rc",
			expect: []ossFile{
				{"datakit/rc/version", path.Join("pub", "release", "version")},
				{"datakit/rc/datakit.yaml", "datakit.yaml"},
				{"datakit/rc/datakit-elinker.yaml", "datakit-elinker.yaml"},
				{"datakit/rc/install.sh", "install.sh"},
				{"datakit/rc/install.ps1", "install.ps1"},
				{fmt.Sprintf("datakit/rc/datakit-%s.yaml", "1.1.3"), "datakit.yaml"},
				{fmt.Sprintf("datakit/rc/datakit-elinker-%s.yaml", "1.1.3"), "datakit-elinker.yaml"},
				{fmt.Sprintf("datakit/rc/install-%s.sh", "1.1.3"), "install.sh"},
				{fmt.Sprintf("datakit/rc/install-%s.ps1", "1.1.3"), "install.ps1"},
			},
			files: []ossFile{
				{"version", path.Join("pub", "release", "version")},
				{"datakit.yaml", "datakit.yaml"},
				{"datakit-elinker.yaml", "datakit-elinker.yaml"},
				{"install.sh", "install.sh"},
				{"install.ps1", "install.ps1"},
				{fmt.Sprintf("datakit-%s.yaml", "1.1.3"), "datakit.yaml"},
				{fmt.Sprintf("datakit-elinker-%s.yaml", "1.1.3"), "datakit.yaml"},
				{fmt.Sprintf("install-%s.sh", "1.1.3"), "install.sh"},
				{fmt.Sprintf("install-%s.ps1", "1.1.3"), "install.ps1"},
			},
		},

		{
			name:    "path-datakit",
			ossPath: "datakit",
			expect: []ossFile{
				{"datakit/version", path.Join("pub", "release", "version")},
				{"datakit/datakit.yaml", "datakit.yaml"},
				{"datakit/datakit-elinker.yaml", "datakit-elinker.yaml"},
				{"datakit/install.sh", "install.sh"},
				{"datakit/install.ps1", "install.ps1"},
				{fmt.Sprintf("datakit/datakit-%s.yaml", "1.1024.3"), "datakit.yaml"},
				{fmt.Sprintf("datakit/datakit-elinker-%s.yaml", "1.1024.3"), "datakit-elinker.yaml"},
				{fmt.Sprintf("datakit/install-%s.sh", "1.1024.3"), "install.sh"},
				{fmt.Sprintf("datakit/install-%s.ps1", "1.1024.3"), "install.ps1"},
			},
			files: []ossFile{
				{"version", path.Join("pub", "release", "version")},
				{"datakit.yaml", "datakit.yaml"},
				{"datakit-elinker.yaml", "datakit-elinker.yaml"},
				{"install.sh", "install.sh"},
				{"install.ps1", "install.ps1"},
				{fmt.Sprintf("datakit-%s.yaml", "1.1024.3"), "datakit.yaml"},
				{fmt.Sprintf("datakit-elinker-%s.yaml", "1.1024.3"), "datakit-elinker.yaml"},
				{fmt.Sprintf("install-%s.sh", "1.1024.3"), "install.sh"},
				{fmt.Sprintf("install-%s.ps1", "1.1024.3"), "install.ps1"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files := addOSSFiles(tc.ossPath, tc.files)
			for i, x := range files {
				assert.Equal(t, tc.expect[i].remote, x.remote)
				t.Logf("%s -> %s", x.local, x.remote)
			}
		})
	}
}
