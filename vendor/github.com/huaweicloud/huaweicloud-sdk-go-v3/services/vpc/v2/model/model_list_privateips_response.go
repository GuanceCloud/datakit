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
type ListPrivateipsResponse struct {
	// 私有IP列表对象
	Privateips     *[]Privateip `json:"privateips,omitempty"`
	HttpStatusCode int          `json:"-"`
}

func (o ListPrivateipsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPrivateipsResponse struct{}"
	}

	return strings.Join([]string{"ListPrivateipsResponse", string(data)}, " ")
}
