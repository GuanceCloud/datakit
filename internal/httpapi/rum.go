// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ip2isp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput/funcs"
)

var (
	sourcemapLock         sync.RWMutex
	sourcemapCallbackList []SourcemapCallbackFunc
)

type SourcemapCallbackFunc func(string)

func RegisterSourcemapCallback(callback SourcemapCallbackFunc) {
	sourcemapLock.Lock()
	defer sourcemapLock.Unlock()
	sourcemapCallbackList = append(sourcemapCallbackList, callback)
}

func runSourcemapCallback(zipFilePath string) {
	sourcemapLock.RLock()
	defer sourcemapLock.RUnlock()
	for _, cb := range sourcemapCallbackList {
		cb(zipFilePath)
	}
}

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
