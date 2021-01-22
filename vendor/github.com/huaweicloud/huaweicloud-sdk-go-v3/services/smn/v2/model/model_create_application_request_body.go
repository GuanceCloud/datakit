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

type CreateApplicationRequestBody struct {
	// 应用名。  最大支持64个字符，只能包含英文字母、下划线和数字。
	Name string `json:"name"`
	// 应用平台。  目前仅支持HMS、APNS、APNS_SANDBOX。  HMS是为开发者提供的消息推送平台。  APNS和APNS_SANDBOX是用于推送iOS消息的服务平台。
	Platform string `json:"platform"`
	// 对于HMS平台是APP ID，只能包含英文字母和数字，最大20个字符。 对于苹果APNS、APNS_SandBox平台是推送证书，大小不超过8K，且是Base64编码。
	PlatformPrincipal string `json:"platform_principal"`
	// 对于HMS平台是APP SECRET， 只能包含英文字母和数字，32到64个字符。  对于苹果APNS、APNS_SandBox平台是推送证书的私钥（private key）， 大小不超过8K，且是Base64编码。
	PlatformCredential string `json:"platform_credential"`
}

func (o CreateApplicationRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateApplicationRequestBody struct{}"
	}

	return strings.Join([]string{"CreateApplicationRequestBody", string(data)}, " ")
}
