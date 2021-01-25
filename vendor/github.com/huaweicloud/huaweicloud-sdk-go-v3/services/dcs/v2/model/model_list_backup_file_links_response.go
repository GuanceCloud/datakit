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

// Response Object
type ListBackupFileLinksResponse struct {
	// OBS桶内文件路径。
	FilePath *string `json:"file_path,omitempty"`
	// OBS桶名。
	BucketName *string `json:"bucket_name,omitempty"`
	// 备份文件下链接集合，链接数最大为64个。
	Links          *[]LinksItem `json:"links,omitempty"`
	HttpStatusCode int          `json:"-"`
}

func (o ListBackupFileLinksResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListBackupFileLinksResponse struct{}"
	}

	return strings.Join([]string{"ListBackupFileLinksResponse", string(data)}, " ")
}
