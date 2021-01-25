/*
 * EIP
 *
 * 云服务接口
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// marker分页结构
type Pager struct {
	// 页码url
	Href *string `json:"href,omitempty"`
	// next:下一页  previous:前一页
	Rel *PagerRel `json:"rel,omitempty"`
}

func (o Pager) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Pager struct{}"
	}

	return strings.Join([]string{"Pager", string(data)}, " ")
}

type PagerRel struct {
	value string
}

type PagerRelEnum struct {
	NEXT     PagerRel
	PREVIOUS PagerRel
}

func GetPagerRelEnum() PagerRelEnum {
	return PagerRelEnum{
		NEXT: PagerRel{
			value: "next",
		},
		PREVIOUS: PagerRel{
			value: "previous",
		},
	}
}

func (c PagerRel) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *PagerRel) UnmarshalJSON(b []byte) error {
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
