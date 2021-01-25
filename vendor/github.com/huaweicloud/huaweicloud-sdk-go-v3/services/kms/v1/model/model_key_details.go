/*
 * kms
 *
 * KMS v1.0 API, open API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

type KeyDetails struct {
	// 密钥ID。
	KeyId *string `json:"key_id,omitempty"`
	// 用户域ID。
	DomainId *string `json:"domain_id,omitempty"`
	// 密钥别名。
	KeyAlias *string `json:"key_alias,omitempty"`
	// 密钥区域。
	Realm *string `json:"realm,omitempty"`
	// 密钥描述。
	KeyDescription *string `json:"key_description,omitempty"`
	// 密钥创建时间，时间戳，即从1970年1月1日至该时间的总秒数。
	CreationDate *string `json:"creation_date,omitempty"`
	// 密钥计划删除时间，时间戳，即从1970年1月1日至该时间的总秒数。
	ScheduledDeletionDate *string `json:"scheduled_deletion_date,omitempty"`
	// 密钥状态，满足正则匹配“^[1-5]{1}$”，枚举如下：  - “1”表示待激活状态  - “2”表示启用状态  - “3”表示禁用状态  - “4”表示计划删除状态  - “5”表示等待导入状态
	KeyState *string `json:"key_state,omitempty"`
	// 默认主密钥标识，默认主密钥标识为1，非默认标识为0。
	DefaultKeyFlag *string `json:"default_key_flag,omitempty"`
	// 密钥类型。
	KeyType *string `json:"key_type,omitempty"`
	// 密钥材料失效时间，时间戳，即从1970年1月1日至该时间的总秒数。
	ExpirationTime *string `json:"expiration_time,omitempty"`
	// 密钥来源，默认为“kms”，枚举如下：  - kms表示密钥材料由kms生成kms表示密钥材料由kms生成  - external表示密钥材料由外部导入
	Origin *KeyDetailsOrigin `json:"origin,omitempty"`
	// 密钥轮换状态，默认为“false”，表示关闭密钥轮换功能。
	KeyRotationEnabled *string `json:"key_rotation_enabled,omitempty"`
	// 企业项目ID，默认为“0”。  - 对于开通企业项目的用户，表示资源处于默认企业项目下。  - 对于未开通企业项目的用户，表示资源未处于企业项目下。
	SysEnterpriseProjectId *string `json:"sys_enterprise_project_id,omitempty"`
}

func (o KeyDetails) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "KeyDetails struct{}"
	}

	return strings.Join([]string{"KeyDetails", string(data)}, " ")
}

type KeyDetailsOrigin struct {
	value string
}

type KeyDetailsOriginEnum struct {
	KMS      KeyDetailsOrigin
	EXTERNAL KeyDetailsOrigin
}

func GetKeyDetailsOriginEnum() KeyDetailsOriginEnum {
	return KeyDetailsOriginEnum{
		KMS: KeyDetailsOrigin{
			value: "kms",
		},
		EXTERNAL: KeyDetailsOrigin{
			value: "external",
		},
	}
}

func (c KeyDetailsOrigin) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *KeyDetailsOrigin) UnmarshalJSON(b []byte) error {
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
