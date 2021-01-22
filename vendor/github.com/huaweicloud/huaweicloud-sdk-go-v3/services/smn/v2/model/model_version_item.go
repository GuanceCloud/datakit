/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type VersionItem struct {
	// 版本ID（版本号），如v2。
	Id string `json:"id"`
	// 若该版本API支持微版本，则返回支持的最小微版本号；若不支持微版本，则返回空。
	MinVersion string `json:"min_version"`
	// 版本状态，为如下3种：  CURRENT：表示该版本为主推版本。  SUPPORTED：表示为老版本，但是现在还继续支持。  DEPRECATED：表示为废弃版本，存在后续删除的可能。
	Status string `json:"status"`
	// 版本发布时间，要求用UTC时间格式表示。如v2发布的时间2014-06-28T12:20:21Z。
	Updated string `json:"updated"`
	// 若该版本API支持微版本，则返回支持的最大微版本号；若不支持微版本，则返回空。
	Version string `json:"version"`
	// API的URL地址。
	Links []LinksItem `json:"links"`
}

func (o VersionItem) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "VersionItem struct{}"
	}

	return strings.Join([]string{"VersionItem", string(data)}, " ")
}
