package externals

import (
	"fmt"
	"path"
	"runtime"
)

func InstallTelegraf(url string) error {
	if url == "" {
		url = "https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit/" + path.Join("telegraf", fmt.Sprintf("agent-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH))
	}

	fmt.Println(url)
	return nil
}
