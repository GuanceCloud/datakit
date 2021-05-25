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
type ChangeServerOsWithCloudInitRequestBody struct {
	OsChange *ChangeServerOsWithCloudInitOption `json:"os-change"`
}

func (o ChangeServerOsWithCloudInitRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ChangeServerOsWithCloudInitRequestBody struct{}"
	}

	return strings.Join([]string{"ChangeServerOsWithCloudInitRequestBody", string(data)}, " ")
}
