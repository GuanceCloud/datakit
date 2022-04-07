package cmds

import (
	"fmt"
	"path/filepath"
	"runtime"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dl "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/downloader"
)

func InstallDatakitEbpf() error {
	url := "https://" + filepath.Join(datakit.DownloadAddr, fmt.Sprintf(
		"datakit-ebpf-%s-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH, datakit.Version))

	if runtime.GOOS != datakit.OSLinux || runtime.GOARCH != "amd64" {
		return fmt.Errorf("datakit-ebpf only supports linux/amd64")
	}

	infof("install datakit-ebpf...\n")
	dl.CurDownloading = "datakit-ebpf"
	cli := getcli()

	if err := dl.Download(cli, url, filepath.Join(datakit.InstallDir, "externals"), false, false); err != nil {
		return err
	}
	infof("install success")
	return nil
}
