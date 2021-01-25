/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListFlavorsResponse struct {
	// 资源规格列表。
	Flavors        *[]FlavorView `json:"flavors,omitempty"`
	HttpStatusCode int           `json:"-"`
}

func (o ListFlavorsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListFlavorsResponse struct{}"
	}

	return strings.Join([]string{"ListFlavorsResponse", string(data)}, " ")
}
