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

// Request Object
type DeleteTagRequest struct {
	KeyId     string `json:"key_id"`
	Key       string `json:"key"`
	VersionId string `json:"version_id"`
}

func (o DeleteTagRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteTagRequest struct{}"
	}

	return strings.Join([]string{"DeleteTagRequest", string(data)}, " ")
}
