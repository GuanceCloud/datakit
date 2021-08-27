package cmds

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dl "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/downloader"
)

const (
	dataUrl = "https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit/data.tar.gz"
)

func updateIPDB(addr string) error {
	if addr == "" {
		addr = dataUrl
	}

	fmt.Printf("Start downloading data.tar.gz...\n")

	cli := getcli()

	dl.CurDownloading = "ipdb"
	if err := dl.Download(cli, addr, datakit.InstallDir, true, false); err != nil {
		return err
	}

	fmt.Printf("Download and decompress data.tar.gz successfully. Please restart datakit to load new IPDB\n")

	return nil
}
