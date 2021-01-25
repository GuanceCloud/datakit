/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type UpdateApplicationRequestBody struct {
	// 对于HMS平台是APP ID，只能包含英文字母和数字，最大20个字符。  对于苹果APNS、APNS_SandBox平台是推送证书，大小不超过8K，且是Base64编码。
	PlatformPrincipal string `json:"platform_principal"`
	// 对于HMS平台是APP SECRET， 只能包含英文字母和数字，32到64个字符。  对于苹果APNS、APNS_SandBox平台是推送证书的私钥（private key）， 大小不超过8K，且是Base64编码。
	PlatformCredential string `json:"platform_credential"`
}

func (o UpdateApplicationRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateApplicationRequestBody struct{}"
	}

	return strings.Join([]string{"UpdateApplicationRequestBody", string(data)}, " ")
}
