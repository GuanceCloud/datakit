package cmds

import (
	"fmt"
	"path/filepath"
	"runtime"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dl "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/downloader"
)

func InstallEbpf() error {
	url := "https://" + filepath.Join(datakit.DownloadAddr, fmt.Sprintf(
		"datakit-ebpf-%s-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH, datakit.Version))

	if runtime.GOOS != datakit.OSLinux || (runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64") {
		return fmt.Errorf("DataKit eBPF plugin only supports linux/amd64 and linux/arm64")
	}

	infof("install DataKit eBPF plugin...\n")
	dl.CurDownloading = "datakit-ebpf"
	cli := getcli()

	if err := dl.Download(cli, url, filepath.Join(datakit.InstallDir, "externals"), false, false); err != nil {
		return err
	}
	infof("install success\n")
	return nil
}
