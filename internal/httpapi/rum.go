// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"archive/zip"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-sourcemap/sourcemap"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ip2isp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput/funcs"
)

var (
	sourcemapCache = make(map[string]map[string]*sourcemap.Consumer)
	sourcemapLock  sync.Mutex
)

func geoTags(srcip string) (tags map[string]string) {
	// default set to be unknown
	tags = map[string]string{
		"city":     "unknown",
		"province": "unknown",
		"country":  "unknown",
		"isp":      "unknown",
		"ip":       srcip,
	}

	if srcip == "" {
		return
	}

	ipInfo, err := funcs.Geo(srcip)

	l.Debugf("ipinfo(%s): %+#v", srcip, ipInfo)

	if err != nil {
		l.Warnf("geo failed: %s, ignored", err)
		return
	}

	// avoid nil pointer error
	if ipInfo == nil {
		return tags
	}

	switch ipInfo.Country { // #issue 354
	case "TW":
		ipInfo.Country = "CN"
		ipInfo.Region = "Taiwan"
	case "MO":
		ipInfo.Country = "CN"
		ipInfo.Region = "Macao"
	case "HK":
		ipInfo.Country = "CN"
		ipInfo.Region = "Hong Kong"
	}

	if len(ipInfo.City) > 0 {
		tags["city"] = ipInfo.City
	}

	if len(ipInfo.Region) > 0 {
		tags["province"] = ipInfo.Region
	}

	if len(ipInfo.Country) > 0 {
		tags["country"] = ipInfo.Country
	}

	if len(srcip) > 0 {
		tags["ip"] = srcip
	}

	if isp := ip2isp.SearchISP(srcip); len(isp) > 0 {
		tags["isp"] = isp
	}

	return tags
}

func GetRumSourcemapDir() string {
	return datakit.DataRUMDir
}

func loadZipFile(zipFile string) (map[string]*sourcemap.Consumer, error) {
	sourcemapItem := make(map[string]*sourcemap.Consumer)

	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return nil, err
	}
	defer zipReader.Close() //nolint:errcheck

	for _, f := range zipReader.File {
		if !f.FileInfo().IsDir() && strings.HasSuffix(f.Name, ".map") {
			file, err := f.Open()
			if err != nil {
				l.Debugf("ignore sourcemap %s, %s", f.Name, err.Error())
				continue
			}
			defer file.Close() // nolint:errcheck

			content, err := ioutil.ReadAll(file)
			if err != nil {
				l.Debugf("ignore sourcemap %s, %s", f.Name, err.Error())
				continue
			}

			smap, err := sourcemap.Parse(f.Name, content)
			if err != nil {
				l.Debugf("sourcemap parse failed, %s", err.Error())
				continue
			}

			sourcemapItem[f.Name] = smap
		}
	}

	return sourcemapItem, nil
}

func updateSourcemapCache(zipFile string) {
	fileName := filepath.Base(zipFile)
	if strings.HasSuffix(fileName, ".zip") {
		if sourcemapItem, err := loadZipFile(zipFile); err != nil {
			l.Debugf("load zip file error: %s", err.Error())
		} else {
			sourcemapLock.Lock()
			sourcemapCache[fileName] = sourcemapItem
			sourcemapLock.Unlock()
			l.Debugf("load sourcemap: %s", fileName)
		}
	}
}

func deleteSourcemapCache(zipFile string) {
	fileName := filepath.Base(zipFile)
	if strings.HasSuffix(fileName, ".zip") {
		sourcemapLock.Lock()
		delete(sourcemapCache, fileName)
		sourcemapLock.Unlock()
	}
}
