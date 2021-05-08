package cmds

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/installer/install"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

const (
	DIR_NAME = "telegraf"
)

func InstallTelegraf(installDir string) error {
	url := "https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit/telegraf/" + fmt.Sprintf("telegraf-%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)

	fmt.Printf("Start downloading Telegraf...\n")
	if err := install.Download(url, installDir, false, false); err != nil {
		return err
	}

	if err := writeTelegrafSample(installDir); err != nil {
		return err
	}

	fmt.Printf("Install Telegraf successfully!\n")
	if runtime.GOOS == "windows" {
		fmt.Printf("Start telegraf by `cd %v`, `copy telegraf.conf.sample telegraf.conf`, and `telegraf.exe --config <file>`\n", filepath.Join(installDir, DIR_NAME))
	} else {
		fmt.Printf("Start telegraf by `cd %v`, `cp telegraf.conf.sample telegraf.conf`, and `./usr/bin/telegraf --config telegraf.conf`\n", filepath.Join(installDir, DIR_NAME))
	}

	fmt.Printf("Vist https://www.influxdata.com/time-series-platform/telegraf/ for more infomation.\n")

	return nil
}

func writeTelegrafSample(installDir string) error {
	if err := config.LoadCfg(datakit.Cfg, datakit.MainConfPath); err != nil {
		return err
	}

	file := filepath.Join(installDir, DIR_NAME, "telegraf.conf.sample")
	return ioutil.WriteFile(file, []byte(TelegrafConfTemplate), 0x666)
}
