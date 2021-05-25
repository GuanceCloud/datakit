/*
 * ECS
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// This is a auto create Body Object
type ResizePostPaidServerRequestBody struct {
	Resize *ResizePostPaidServerOption `json:"resize"`
}

func (o ResizePostPaidServerRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResizePostPaidServerRequestBody struct{}"
	}

	return strings.Join([]string{"ResizePostPaidServerRequestBody", string(data)}, " ")
}
