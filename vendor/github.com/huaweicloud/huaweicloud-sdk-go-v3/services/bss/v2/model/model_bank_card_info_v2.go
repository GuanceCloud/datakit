/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type BankCardInfoV2 struct {
	// |参数名称：银行卡账号。当identifyType为4时，不能为空。银行账号输入规则：^[0-9a-zA-Z]，可以包含特殊横杠（-）字符。| |参数约束及描述：银行卡账号。当identifyType为4时，不能为空。银行账号输入规则：^[0-9a-zA-Z]，可以包含特殊横杠（-）字符。|
	BankAccount string `json:"bank_account"`
	// |参数名称：国家/区号码。例如：0086：中国大陆区号码。| |参数约束及描述：国家/区号码。例如：0086：中国大陆区号码。|
	Areacode string `json:"areacode"`
	// |参数名称：手机号码。| |参数约束及描述：手机号码。|
	Mobile string `json:"mobile"`
	// |参数名称：验证码。| |参数约束及描述：验证码。|
	VerificationCode string `json:"verification_code"`
}

func (o BankCardInfoV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BankCardInfoV2 struct{}"
	}

	return strings.Join([]string{"BankCardInfoV2", string(data)}, " ")
}
