/*
 * CCE
 *
 * CCE开放API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type AwakeClusterResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o AwakeClusterResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AwakeClusterResponse struct{}"
	}

	return strings.Join([]string{"AwakeClusterResponse", string(data)}, " ")
}
