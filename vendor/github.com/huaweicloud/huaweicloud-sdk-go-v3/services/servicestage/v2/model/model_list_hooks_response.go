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
type ListHooksResponse struct {
	// hook列表。
	Hooks          *[]Hook `json:"hooks,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ListHooksResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListHooksResponse struct{}"
	}

	return strings.Join([]string{"ListHooksResponse", string(data)}, " ")
}
