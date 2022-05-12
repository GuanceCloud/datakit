// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dl "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/downloader"
)

const (
	dataURL = "https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit/data.tar.gz"
)

func updateIPDB(addr string) error {
	if addr == "" {
		addr = dataURL
	}

	fmt.Printf("Start downloading data.tar.gz...\n")

	cli := getcli()

	dl.CurDownloading = "ipdb"
	if err := dl.Download(cli, addr, datakit.InstallDir, true, false); err != nil {
		return err
	}

	return nil
}
