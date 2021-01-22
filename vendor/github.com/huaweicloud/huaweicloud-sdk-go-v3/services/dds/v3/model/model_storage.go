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

// 实例磁盘类型信息。
type Storage struct {
	// 磁盘类型名称，可能取值如下： - ULTRAHIGH，表示SSD。
	Name string `json:"name"`
	// 其中key是可用区编号，value是规格所在az的状态，包含以下状态： - normal，在售。 - unsupported，暂不支持该规格。 - sellout，售罄。
	AzStatus map[string]string `json:"az_status"`
}

func (o Storage) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Storage struct{}"
	}

	return strings.Join([]string{"Storage", string(data)}, " ")
}
