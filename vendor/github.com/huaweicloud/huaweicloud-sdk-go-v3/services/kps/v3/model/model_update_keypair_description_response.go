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
type UpdateKeypairDescriptionResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o UpdateKeypairDescriptionResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateKeypairDescriptionResponse struct{}"
	}

	return strings.Join([]string{"UpdateKeypairDescriptionResponse", string(data)}, " ")
}
