/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// 热key记录结构体
type HotkeysBody struct {
	// key名称
	Name *string `json:"name,omitempty"`
	// key类型
	Type *HotkeysBodyType `json:"type,omitempty"`
	// 热key所在的分片，仅在实例类型为集群时支持,格式为ip:port
	Shard *string `json:"shard,omitempty"`
	// 热key所在的db
	Db *int32 `json:"db,omitempty"`
	// key的value大小。
	Size *int32 `json:"size,omitempty"`
	// key大小的单位。type为string时，单位是：byte；type为list/set/zset/hash时，单位是：count
	Unit *string `json:"unit,omitempty"`
	// 表示某个key在一段时间的访问频度，会随着访问的频率而变化。  该值并不是简单的访问频率值，而是一个基于概率的对数计数器结果，最大为255(可表示100万次访问)，超过255后如果继续频繁访问该值并不会继续增大，同时默认如果每过一分钟没有访问，该值会衰减1。
	Freq *int32 `json:"freq,omitempty"`
}

func (o HotkeysBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "HotkeysBody struct{}"
	}

	return strings.Join([]string{"HotkeysBody", string(data)}, " ")
}

type HotkeysBodyType struct {
	value string
}

type HotkeysBodyTypeEnum struct {
	STRING HotkeysBodyType
	LIST   HotkeysBodyType
	SET    HotkeysBodyType
	ZSET   HotkeysBodyType
	HASH   HotkeysBodyType
}

func GetHotkeysBodyTypeEnum() HotkeysBodyTypeEnum {
	return HotkeysBodyTypeEnum{
		STRING: HotkeysBodyType{
			value: "string",
		},
		LIST: HotkeysBodyType{
			value: "list",
		},
		SET: HotkeysBodyType{
			value: "set",
		},
		ZSET: HotkeysBodyType{
			value: "zset",
		},
		HASH: HotkeysBodyType{
			value: "hash",
		},
	}
}

func (c HotkeysBodyType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *HotkeysBodyType) UnmarshalJSON(b []byte) error {
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
