// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dl "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/downloader"
)

const (
	baseURL = "https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit/ipdb/"
)

var allDB = []string{
	"iploc.tar.gz",
	"geolite2.tar.gz",
}

var dbDir = []string{
	"iploc",
	"geolite2",
}

func updateIPDB() error {
	installDir := datakit.DataDir + "/ipdb/"
	for idx, c := range allDB {
		fmt.Printf("Start downloading %s...\n", c)
		c = baseURL + c
		cli := getcli()

		dl.CurDownloading = "ipdb"
		if err := dl.Download(cli, c, installDir+dbDir[idx], true, false); err != nil {
			return err
		}
	}
	return nil
}
