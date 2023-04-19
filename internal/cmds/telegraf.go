// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	dl "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/downloader"
)

const (
	dirName = "telegraf"
)

func installTelegraf(installDir string) error {
	url := "https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit/telegraf/" +
		fmt.Sprintf("telegraf-%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)

	if runtime.GOOS != datakit.OSWindows {
		installDir = "/"
	}

	cp.Infof("Start downloading Telegraf...\n")
	dl.CurDownloading = "telegraf"

	cli := getcli()

	if err := dl.Download(cli, url, installDir, false, false); err != nil {
		return err
	}

	if err := writeTelegrafSample(installDir); err != nil {
		return err
	}

	//nolint:lll
	cp.Infof("Install Telegraf successfully!\n")
	if runtime.GOOS == datakit.OSWindows {
		cp.Infof("Start telegraf by `cd %s`, `copy telegraf.conf.sample tg.conf`, and `telegraf.exe --config tg.conf`\n",
			filepath.Join(installDir, dirName))
	} else {
		cp.Infof("Start telegraf by `cd %s`, `cp telegraf.conf.sample tg.conf`, and `telegraf --config tg.conf`\n",
			filepath.Join(installDir, dirName))
	}

	cp.Infof("Vist https://www.influxdata.com/time-series-platform/telegraf/ for more information.\n")

	return nil
}

func writeTelegrafSample(installDir string) error {
	var filePath string
	if runtime.GOOS != "windows" {
		filePath = "/etc/telegraf/telegraf.conf.sample"
	} else {
		filePath = filepath.Join(installDir, dirName, "telegraf.conf.sample")
	}

	return ioutil.WriteFile(filePath, []byte(TelegrafConfTemplate), os.ModePerm)
}
