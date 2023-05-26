// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"os"
	"path/filepath"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ip2isp"
)

func BuildISP() {
	curDir, _ := os.Getwd()

	inputIPDir := filepath.Join(curDir, "china-operator-ip")
	ip2ispFile := filepath.Join(curDir, "pipeline", "ip2isp", "ip2isp.txt")
	if err := os.Remove(ip2ispFile); err != nil {
		l.Warnf("os.Remove: %s, ignored", err.Error())
	}

	if err := ip2isp.MergeISP(inputIPDir, ip2ispFile); err != nil {
		l.Errorf("MergeIsp failed: %v", err)
	} else {
		l.Infof("merge ip2isp file in `%v`", ip2ispFile)
	}

	inputFile := filepath.Join(curDir, "IP2LOCATION-LITE-DB11.CSV")
	outputFile := filepath.Join(curDir, "pipeline", "ip2isp", "contry_city.yaml")
	if !datakit.FileExist(inputFile) {
		l.Errorf("%v not exist, you can download from `https://lite.ip2location.com/download?id=9`", inputFile)
		os.Exit(0)
	}

	if err := os.Remove(ip2ispFile); err != nil {
		l.Warnf("os.Remove: %s, ignored", err.Error())
	}

	if err := ip2isp.BuildContryCity(inputFile, outputFile); err != nil {
		l.Errorf("BuildContryCity failed: %v", err)
	} else {
		l.Infof("contry and city list in file  `%v`", outputFile)
	}
}
