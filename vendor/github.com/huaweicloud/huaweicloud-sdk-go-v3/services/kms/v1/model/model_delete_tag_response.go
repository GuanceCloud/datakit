/*
 * kms
 *
 * KMS v1.0 API, open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type DeleteTagResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteTagResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteTagResponse struct{}"
	}

	return strings.Join([]string{"DeleteTagResponse", string(data)}, " ")
}
