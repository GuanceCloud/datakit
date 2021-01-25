/*
 * DIS
 *
 * DIS v1 API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type GetCheckpointResponse struct {
	// 序列号，用来记录该通道的消费检查点。
	SequenceNumber *string `json:"sequence_number,omitempty"`
	// 用户消费程序端的元数据信息。
	Metadata       *string `json:"metadata,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o GetCheckpointResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "GetCheckpointResponse struct{}"
	}

	return strings.Join([]string{"GetCheckpointResponse", string(data)}, " ")
}
