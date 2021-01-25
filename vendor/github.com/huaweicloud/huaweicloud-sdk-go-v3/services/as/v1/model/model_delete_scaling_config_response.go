/*
 * AS
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type DeleteScalingConfigResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteScalingConfigResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteScalingConfigResponse struct{}"
	}

	return strings.Join([]string{"DeleteScalingConfigResponse", string(data)}, " ")
}
