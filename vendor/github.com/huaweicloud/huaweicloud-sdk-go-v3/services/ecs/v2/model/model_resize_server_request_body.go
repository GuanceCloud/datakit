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
type ResizeServerRequestBody struct {
	Resize *ResizePrePaidServerOption `json:"resize"`
}

func (o ResizeServerRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResizeServerRequestBody struct{}"
	}

	return strings.Join([]string{"ResizeServerRequestBody", string(data)}, " ")
}
