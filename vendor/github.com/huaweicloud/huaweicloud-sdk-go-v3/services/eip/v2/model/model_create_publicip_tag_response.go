/*
 * EIP
 *
 * 云服务接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type CreatePublicipTagResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CreatePublicipTagResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePublicipTagResponse struct{}"
	}

	return strings.Join([]string{"CreatePublicipTagResponse", string(data)}, " ")
}
