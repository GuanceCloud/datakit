/*
 * DDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListStorageTypeResponse struct {
	// 实例磁盘类型信息。
	StorageType *[]Storage `json:"storage_type,omitempty"`
	// 实例专属存储信息。
	DssPoolInfo    *[]DssPoolInfo `json:"dss_pool_info,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o ListStorageTypeResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListStorageTypeResponse struct{}"
	}

	return strings.Join([]string{"ListStorageTypeResponse", string(data)}, " ")
}
