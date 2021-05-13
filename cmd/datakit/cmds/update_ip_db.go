package cmds

import (
	"fmt"
	nhttp "net/http"
	"os"
	"path/filepath"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/installer/install"
)

const (
	dataUrl = "https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit/data.tar.gz"
)

func UpdateIpDB(port int, addr string) error {
	if addr == "" {
		addr = dataUrl
	}

	os.RemoveAll(filepath.Join(datakit.InstallDir, "data", "ip2isp"))

	fmt.Printf("Start downloading data.tar.gz...\n")

	if err := install.Download(addr, datakit.InstallDir, true, false); err != nil {
		return err
	}

	fmt.Printf("Download and decompress data.tar.gz successfully.\n")

	nhttp.Get(fmt.Sprintf("http://127.0.0.1:%d/reload", port))

	fmt.Printf("Update successfully.\n")

	return nil
}
