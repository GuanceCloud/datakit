/*
 * kms
 *
 * KMS v1.0 API, open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ApiVersionDetail struct {
	// 版本ID（版本号），如“v1.0”。
	Id *string `json:"id,omitempty"`
	// JSON对象，详情请参见links字段数据结构说明。
	Links *[]ApiLink `json:"links,omitempty"`
	// 若该版本API支持微版本，则填支持的最大微版本号，如果不支持微版本，则返回空字符串。
	Version *string `json:"version,omitempty"`
	// 版本状态，包含如下3种：  - CURRENT：表示该版本为主推版本。  - SUPPORTED：表示为老版本，但是现在还继续支持。  - DEPRECATED：表示为废弃版本，存在后续删除的可能。
	Status *string `json:"status,omitempty"`
	// 版本发布时间，要求用UTC时间表示。如v1.发布的时间2014-06-28T12:20:21Z。
	Updated *string `json:"updated,omitempty"`
	// 若该版本API 支持微版本，则填支持的最小微版本号，如果不支持微版本，则返回空字符串。
	MinVersion *string `json:"min_version,omitempty"`
}

func (o ApiVersionDetail) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApiVersionDetail struct{}"
	}

	return strings.Join([]string{"ApiVersionDetail", string(data)}, " ")
}
