/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// Response Object
type ShowContentResponse struct {
	// 文件路径。
	Path *string `json:"path,omitempty"`
	// commit 哈希。
	Sha *string `json:"sha,omitempty"`
	// 编码方式：base64或者text/plain。
	Encoding *ShowContentResponseEncoding `json:"encoding,omitempty"`
	// 文件内容。
	Content        *string `json:"content,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowContentResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowContentResponse struct{}"
	}

	return strings.Join([]string{"ShowContentResponse", string(data)}, " ")
}

type ShowContentResponseEncoding struct {
	value string
}

type ShowContentResponseEncodingEnum struct {
	BASE64     ShowContentResponseEncoding
	TEXT_PLAIN ShowContentResponseEncoding
}

func GetShowContentResponseEncodingEnum() ShowContentResponseEncodingEnum {
	return ShowContentResponseEncodingEnum{
		BASE64: ShowContentResponseEncoding{
			value: "base64",
		},
		TEXT_PLAIN: ShowContentResponseEncoding{
			value: "text/plain",
		},
	}
}

func (c ShowContentResponseEncoding) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowContentResponseEncoding) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("string")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(string)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to string error")
	}
}
