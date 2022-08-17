// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dl "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/downloader"
)

type ipdbInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Time    int64  `json:"time"` // ms
}

func installIPDB(ipdbType string) error {
	baseURL := "https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit"
	ipdb, err := InstallIPDB(baseURL, ipdbType)
	if err != nil {
		return err
	}

	infof("\n\nInstall ipdb successfully and its information is as follows.\n")

	infof("ipdb_type: %s\nversion: %s\ntime: %s\n\n", ipdbType, ipdb.Version, time.Unix(ipdb.Time/1000, 0))
	infof("To enable it, you should set `pipeline.ipdb_type` to \"%s\" in `conf.d/datakit.conf` and restart the datakit!\n", ipdbType)

	return nil
}

func InstallIPDB(baseURL string, ipdbType string) (*ipdbInfo, error) {
	ipdbBaseURL := baseURL + "/ipdb/"
	ipdbJSONURL := ipdbBaseURL + ipdbType + ".json"
	installDir := datakit.DataDir + "/ipdb/" + ipdbType

	// nolint:gosec
	if resp, err := http.Get(ipdbJSONURL); err != nil {
		return nil, err
	} else {
		res, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		// nolint:errcheck
		defer resp.Body.Close()

		ipdb := ipdbInfo{}

		if err := json.Unmarshal(res, &ipdb); err != nil {
			return nil, fmt.Errorf("read IPDB json info '%s' failed: %w", ipdbJSONURL, err)
		}

		ipdbURL := ipdbBaseURL + ipdb.Name

		infof("Start downloading ipdb ...\n")

		dl.CurDownloading = "ipdb"
		cli := getcli()
		if err := dl.Download(cli, ipdbURL, installDir, true, false); err != nil {
			return nil, fmt.Errorf("download %s to %s failed: %w", ipdbURL, installDir, err)
		}
		return &ipdb, nil
	}
}
