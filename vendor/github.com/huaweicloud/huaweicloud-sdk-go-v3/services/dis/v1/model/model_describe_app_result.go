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

type DescribeAppResult struct {
	// App的名称。
	AppName *string `json:"app_name,omitempty"`
	// App的唯一标识符。
	AppId *string `json:"app_id,omitempty"`
	// App创建的时间，单位毫秒。
	CreateTime *int64 `json:"create_time,omitempty"`
}

func (o DescribeAppResult) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DescribeAppResult struct{}"
	}

	return strings.Join([]string{"DescribeAppResult", string(data)}, " ")
}
