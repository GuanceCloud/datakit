/*
 * DIS
 *
 * DIS v1 API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type DeleteAppRequest struct {
	AppName string `json:"app_name"`
}

func (o DeleteAppRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteAppRequest struct{}"
	}

	return strings.Join([]string{"DeleteAppRequest", string(data)}, " ")
}
