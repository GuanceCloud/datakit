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

// 更新SSH密钥对描述
type UpdateKeypairDescriptionRequestBody struct {
	Keypair *UpdateKeypairDescriptionReq `json:"keypair"`
}

func (o UpdateKeypairDescriptionRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateKeypairDescriptionRequestBody struct{}"
	}

	return strings.Join([]string{"UpdateKeypairDescriptionRequestBody", string(data)}, " ")
}
