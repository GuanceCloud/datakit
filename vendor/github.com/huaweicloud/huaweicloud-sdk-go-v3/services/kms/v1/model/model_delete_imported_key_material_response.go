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
type DeleteImportedKeyMaterialResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteImportedKeyMaterialResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteImportedKeyMaterialResponse struct{}"
	}

	return strings.Join([]string{"DeleteImportedKeyMaterialResponse", string(data)}, " ")
}
