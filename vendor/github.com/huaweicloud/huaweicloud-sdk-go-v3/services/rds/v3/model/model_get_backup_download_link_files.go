/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type GetBackupDownloadLinkFiles struct {
	// 文件名。
	Name string `json:"name"`
	// 文件大小，单位为KB。
	Size int64 `json:"size"`
	// 文件下载链接。
	DownloadLink string `json:"download_link"`
	// 下载链接过期时间，格式为“yyyy-mm-ddThh:mm:ssZ”。其中，T指某个时间的开始，Z指时区偏移量，例如北京时间偏移显示为+0800。
	LinkExpiredTime string `json:"link_expired_time"`
}

func (o GetBackupDownloadLinkFiles) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "GetBackupDownloadLinkFiles struct{}"
	}

	return strings.Join([]string{"GetBackupDownloadLinkFiles", string(data)}, " ")
}
