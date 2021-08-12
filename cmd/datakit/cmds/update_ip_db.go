package cmds

import (
	"fmt"
	nhttp "net/http"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	dataUrl = "https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit/data.tar.gz"
)

func updateIPDB(dkhost, addr string) error {
	if addr == "" {
		addr = dataUrl
	}

	fmt.Printf("Start downloading data.tar.gz...\n")

	curDownloading = "ipdb"
	if err := download(addr, datakit.InstallDir, true, false); err != nil {
		return err
	}

	fmt.Printf("Download and decompress data.tar.gz successfully.\n")

	_, err := nhttp.Get(fmt.Sprintf("http://%s/reload", dkhost))
	if err != nil {
		fmt.Printf("Reload datakit fail!\n")
		fmt.Printf("You need restart datakit by `datakit --restart` to make effect.\n")
		return nil
	} else {
		fmt.Printf("Update successfully.\n")
	}

	return nil
}
