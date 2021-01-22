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

// Request Object
type ShowApplicationDetailRequest struct {
	ApplicationId string `json:"application_id"`
}

func (o ShowApplicationDetailRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowApplicationDetailRequest struct{}"
	}

	return strings.Join([]string{"ShowApplicationDetailRequest", string(data)}, " ")
}
