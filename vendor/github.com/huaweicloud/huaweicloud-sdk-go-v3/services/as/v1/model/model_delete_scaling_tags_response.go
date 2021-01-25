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
type DeleteScalingTagsResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteScalingTagsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteScalingTagsResponse struct{}"
	}

	return strings.Join([]string{"DeleteScalingTagsResponse", string(data)}, " ")
}
