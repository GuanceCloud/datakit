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

// 备份参数对象。
type CreateManualBackupOption struct {
	// 实例ID。
	InstanceId string `json:"instance_id"`
	// 手动备份名称。 取值范围：长度为4~64位，必须以字母开头（A~Z或a~z），区分大小写，可以包含字母、数字（0~9）、中划线（-）或者下划线（_），不能包含其他特殊字符。
	Name string `json:"name"`
	// 手动备份描述。 取值范围：长度不超过256位，且不能包含>!<\"&'=特殊字符。
	Description *string `json:"description,omitempty"`
}

func (o CreateManualBackupOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateManualBackupOption struct{}"
	}

	return strings.Join([]string{"CreateManualBackupOption", string(data)}, " ")
}
