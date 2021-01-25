/*
 * DMS
 *
 * DMS Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ConfirmConsumptionMessagesReq struct {
	// 确认消息数组。
	Message *[]ConfirmDeadLettersMessagesReqMessage `json:"message,omitempty"`
}

func (o ConfirmConsumptionMessagesReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ConfirmConsumptionMessagesReq struct{}"
	}

	return strings.Join([]string{"ConfirmConsumptionMessagesReq", string(data)}, " ")
}
