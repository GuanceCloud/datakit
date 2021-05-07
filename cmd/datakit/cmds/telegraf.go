package cmds

import (
	"fmt"
	"path/filepath"
	"runtime"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/installer/install"
)

func InstallTelegraf(installDir string) error {
	url := "https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit/telegraf/" + fmt.Sprintf("telegraf-%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)

	fmt.Printf("Start downloading Telegraf...\n")
	if err := install.Download(url, installDir, false, false); err != nil {
		return err
	}

	fmt.Printf("Install Telegraf successfully!\n")
	if runtime.GOOS == "windows" {
		fmt.Printf("Start telegraf by `cd %v`, and `telegraf.exe --config <file>`\n", filepath.Join(installDir, "telegraf"))
	} else {
		fmt.Printf("Start telegraf by `cd %v`, and ` ./usr/bin/telegraf --config <file>`\n", filepath.Join(installDir, "telegraf"))
	}

	fmt.Printf("Vist https://www.influxdata.com/time-series-platform/telegraf/ for more infomation.\n")

	return nil
}
