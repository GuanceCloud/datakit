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

// Response Object
type CreatePrivateipResponse struct {
	// 私有IP列表对象
	Privateips     *[]Privateip `json:"privateips,omitempty"`
	HttpStatusCode int          `json:"-"`
}

func (o CreatePrivateipResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePrivateipResponse struct{}"
	}

	return strings.Join([]string{"CreatePrivateipResponse", string(data)}, " ")
}
