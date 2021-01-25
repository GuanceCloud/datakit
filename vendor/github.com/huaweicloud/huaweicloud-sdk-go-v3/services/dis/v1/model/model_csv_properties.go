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

type CsvProperties struct {
	// 数据分隔符。
	Delimiter *string `json:"delimiter,omitempty"`
}

func (o CsvProperties) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CsvProperties struct{}"
	}

	return strings.Join([]string{"CsvProperties", string(data)}, " ")
}
