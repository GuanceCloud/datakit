/*
 * kps
 *
 * kps v3 版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type DeleteKeypairResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteKeypairResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteKeypairResponse struct{}"
	}

	return strings.Join([]string{"DeleteKeypairResponse", string(data)}, " ")
}
