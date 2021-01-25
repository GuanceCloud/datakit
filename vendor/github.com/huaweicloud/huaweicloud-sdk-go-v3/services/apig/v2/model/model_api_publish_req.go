/*
 * APIG
 *
 * API网关（API Gateway）是为开发者、合作伙伴提供的高性能、高可用、高安全的API托管服务，帮助用户轻松构建、管理和发布任意规模的API。
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

type ApiPublishReq struct {
	// 需要进行的操作。 - online：发布 - offline：下线
	Action ApiPublishReqAction `json:"action"`
	// 环境的编号，即：API需要发布到哪个环境
	EnvId string `json:"env_id"`
	// API的编号，即：需要进行发布或下线的API的编号
	ApiId string `json:"api_id"`
	// 对发布动作的简述。字符长度不超过255 > 中文字符必须为UTF-8或者unicode编码。
	Remark *string `json:"remark,omitempty"`
}

func (o ApiPublishReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApiPublishReq struct{}"
	}

	return strings.Join([]string{"ApiPublishReq", string(data)}, " ")
}

type ApiPublishReqAction struct {
	value string
}

type ApiPublishReqActionEnum struct {
	ONLINE  ApiPublishReqAction
	OFFLINE ApiPublishReqAction
}

func GetApiPublishReqActionEnum() ApiPublishReqActionEnum {
	return ApiPublishReqActionEnum{
		ONLINE: ApiPublishReqAction{
			value: "online",
		},
		OFFLINE: ApiPublishReqAction{
			value: "offline",
		},
	}
}

func (c ApiPublishReqAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ApiPublishReqAction) UnmarshalJSON(b []byte) error {
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
