/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListConversionsResponse struct {
	// |参数名称：度量信息| |参数约束以及描述：度量信息|
	Conversions    *[]Conversion `json:"conversions,omitempty"`
	HttpStatusCode int           `json:"-"`
}

func (o ListConversionsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListConversionsResponse struct{}"
	}

	return strings.Join([]string{"ListConversionsResponse", string(data)}, " ")
}
