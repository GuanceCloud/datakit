/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type RestoreToExistingInstanceRequestBody struct {
	Source *RestoreToExistingInstanceRequestBodySource `json:"source"`
	Target *RestoreToExistingInstanceRequestBodyTarget `json:"target"`
}

func (o RestoreToExistingInstanceRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RestoreToExistingInstanceRequestBody struct{}"
	}

	return strings.Join([]string{"RestoreToExistingInstanceRequestBody", string(data)}, " ")
}
