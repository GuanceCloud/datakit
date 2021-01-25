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

// Response Object
type DeleteAppResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteAppResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteAppResponse struct{}"
	}

	return strings.Join([]string{"DeleteAppResponse", string(data)}, " ")
}
