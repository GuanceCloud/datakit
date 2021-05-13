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

	ispDir := filepath.Join(datakit.InstallDir, "data", "ip2isp")
	ispDirBack := filepath.Join(datakit.InstallDir, "data", "ip2isp_backup")

	if err := os.Rename(ispDir, ispDirBack); err != nil {
		l.Errorf("rename %s to %s failed: %v", ispDir, ispDirBack, err)
	}

	fmt.Printf("Start downloading data.tar.gz...\n")

	if err := install.Download(addr, datakit.InstallDir, true, false); err != nil {
		if e := os.Rename(ispDirBack, ispDir); e != nil {
			l.Errorf("rename %s to %s failed: %v", ispDirBack, ispDir, e)
		}
		return err
	} else {
		if e := os.RemoveAll(ispDirBack); e != nil {
			l.Errorf("remove %s failed: %v", ispDirBack, e)
		}
	}

	fmt.Printf("Download and decompress data.tar.gz successfully.\n")

	_, err := nhttp.Get(fmt.Sprintf("http://127.0.0.1:%d/reload", port))
	if err != nil {
		fmt.Printf("Reload datakit fail!\n")
		fmt.Printf("You need restart datakit by `datakit --restart` to make effect.\n")
		return nil
	} else {
		fmt.Printf("Update successfully.\n")
	}

	return nil
}
