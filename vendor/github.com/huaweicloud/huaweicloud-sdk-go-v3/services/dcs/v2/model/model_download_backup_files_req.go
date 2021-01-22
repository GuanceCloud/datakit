/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type DownloadBackupFilesReq struct {
	// 设置URL的有效期，必须在5分钟和24小时之内，单位为秒。
	Expiration int32 `json:"expiration"`
}

func (o DownloadBackupFilesReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DownloadBackupFilesReq struct{}"
	}

	return strings.Join([]string{"DownloadBackupFilesReq", string(data)}, " ")
}
