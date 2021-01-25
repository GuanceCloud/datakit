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

type AgentPayInfo struct {
	// |参数名称：是否代付. 0, 租户自己支付;1，合作伙伴代付。如果是待支付状态，这个地方是表明是否申请了代付人支付，如果是已支付状态，是表明是不是由代付人支付。| |参数的约束及描述：支付类型. 0, 租户自己支付;1，合作伙伴代付。|
	IsAgentPay *int32 `json:"is_agent_pay,omitempty"`
	// |参数名称：代付人，当payingType=1的时候有值| |参数约束及描述：代付人，当payingType=1的时候有值|
	PayingAgentId *string `json:"paying_agent_id,omitempty"`
}

func (o AgentPayInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AgentPayInfo struct{}"
	}

	return strings.Join([]string{"AgentPayInfo", string(data)}, " ")
}
