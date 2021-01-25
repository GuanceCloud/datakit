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
type DeletePublicipResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeletePublicipResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeletePublicipResponse struct{}"
	}

	return strings.Join([]string{"DeletePublicipResponse", string(data)}, " ")
}
