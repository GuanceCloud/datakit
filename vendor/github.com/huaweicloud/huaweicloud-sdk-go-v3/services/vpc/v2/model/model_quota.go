/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

//
type Quota struct {
	// 资源列表对象
	Resources []ResourceResult `json:"resources"`
}

func (o Quota) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Quota struct{}"
	}

	return strings.Join([]string{"Quota", string(data)}, " ")
}
