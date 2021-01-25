/*
 * RabbitMQ
 *
 * RabbitMQ Document API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// 创建实例请求体。
type CreateInstanceReq struct {
	// 实例名称。  由英文字符开头，只能由英文字母、数字、中划线、下划线组成，长度为4~64的字符。
	Name string `json:"name"`
	// 实例的描述信息。  长度不超过1024的字符串。  > \\与\"在json报文中属于特殊字符，如果参数值中需要显示\\或者\"字符，请在字符前增加转义字符\\，比如\\\\或者\\\"。
	Description *string `json:"description,omitempty"`
	// 消息引擎：rabbitmq。
	Engine CreateInstanceReqEngine `json:"engine"`
	// 消息引擎的版本。   - RabbitMQ版本有：3.7.17
	EngineVersion CreateInstanceReqEngineVersion `json:"engine_version"`
	// 消息存储空间，单位GB。   - 单机RabbitMQ实例的存储空间的取值范围100GB~90000GB。   - 集群RabbitMQ实例的存储空间的取值范围为100GB*节点数~90000GB、200GB*节点数~90000GB、300GB*节点数~90000GB。
	StorageSpace int32 `json:"storage_space"`
	// 认证用户名，只能由英文字母、数字、中划线组成，长度为4~64的字符。
	AccessUser string `json:"access_user"`
	// 实例的认证密码。  复杂度要求： - 输入长度为8到32位的字符串。 - 必须包含如下四种字符中的两种组合：   - 小写字母   - 大写字母   - 数字   - 特殊字符包括（`~!@#$%^&*()-_=+\\|[{}]:'\",<.>/?）
	Password string `json:"password"`
	// 租户VPC ID。
	VpcId string `json:"vpc_id"`
	// 租户安全组ID。
	SecurityGroupId string `json:"security_group_id"`
	// 子网ID。
	SubnetId string `json:"subnet_id"`
	// 创建节点到指定且有资源的可用区ID。该参数不能为空数组或者数组的值为空，详情请参考[查询可用区信息](https://support.huaweicloud.com/api-rabbitmq/ListAvailableZones.html)查询得到。
	AvailableZones []string `json:"available_zones"`
	// 产品标识。  获取方法，请参考[查询产品规格列表](https://support.huaweicloud.com/api-rabbitmq/ListProducts.html)。
	ProductId string `json:"product_id"`
	// 维护时间窗开始时间，格式为HH:mm。 - 维护时间窗开始和结束时间必须为指定的时间段，可参考[查询维护时间窗时间段](https://support.huaweicloud.com/api-rabbitmq/ShowMaintainWindows.html)获取。 - 开始时间必须为22:00、02:00、06:00、10:00、14:00和18:00。 - 该参数不能单独为空，若该值为空，则结束时间也为空。系统分配一个默认开始时间02:00。
	MaintainBegin *string `json:"maintain_begin,omitempty"`
	// 维护时间窗结束时间，格式为HH:mm。 - 维护时间窗开始和结束时间必须为指定的时间段，可参考[查询维护时间窗时间段](https://support.huaweicloud.com/api-rabbitmq/ShowMaintainWindows.html)获取。 - 结束时间在开始时间基础上加四个小时，即当开始时间为22:00时，结束时间为02:00。 - 该参数不能单独为空，若该值为空，则开始时间也为空，系统分配一个默认结束时间06:00。
	MaintainEnd *string `json:"maintain_end,omitempty"`
	// RabbitMQ实例是否开启公网访问功能。 - true：开启 - false：不开启
	EnablePublicip *bool `json:"enable_publicip,omitempty"`
	// RabbitMQ实例绑定的弹性IP地址的ID。 如果开启了公网访问功能（即enable_publicip为true），该字段为必选。
	PublicipId *string `json:"publicip_id,omitempty"`
	// 是否打开SSL加密访问。 - true：打开SSL加密访问。 - false：不打开SSL加密访问。
	SslEnable *bool `json:"ssl_enable,omitempty"`
	// 存储IO规格。如何选择磁盘类型请参考[磁盘类型及性能介绍](https://support.huaweicloud.com/productdesc-evs/zh-cn_topic_0044524691.html)。 取值范围：   - dms.physical.storage.normal：   - dms.physical.storage.high   - dms.physical.storage.ultra
	StorageSpecCode CreateInstanceReqStorageSpecCode `json:"storage_spec_code"`
	// 企业项目ID。若为企业项目账号，该参数必填。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	// 标签列表。
	Tags *[]CreateInstanceReqTags `json:"tags,omitempty"`
}

func (o CreateInstanceReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateInstanceReq struct{}"
	}

	return strings.Join([]string{"CreateInstanceReq", string(data)}, " ")
}

type CreateInstanceReqEngine struct {
	value string
}

type CreateInstanceReqEngineEnum struct {
	RABBITMQ CreateInstanceReqEngine
}

func GetCreateInstanceReqEngineEnum() CreateInstanceReqEngineEnum {
	return CreateInstanceReqEngineEnum{
		RABBITMQ: CreateInstanceReqEngine{
			value: "rabbitmq",
		},
	}
}

func (c CreateInstanceReqEngine) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateInstanceReqEngine) UnmarshalJSON(b []byte) error {
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

type CreateInstanceReqEngineVersion struct {
	value string
}

type CreateInstanceReqEngineVersionEnum struct {
	E_3_7_17 CreateInstanceReqEngineVersion
}

func GetCreateInstanceReqEngineVersionEnum() CreateInstanceReqEngineVersionEnum {
	return CreateInstanceReqEngineVersionEnum{
		E_3_7_17: CreateInstanceReqEngineVersion{
			value: "3.7.17",
		},
	}
}

func (c CreateInstanceReqEngineVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateInstanceReqEngineVersion) UnmarshalJSON(b []byte) error {
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

type CreateInstanceReqStorageSpecCode struct {
	value string
}

type CreateInstanceReqStorageSpecCodeEnum struct {
	DMS_PHYSICAL_STORAGE_NORMAL CreateInstanceReqStorageSpecCode
	DMS_PHYSICAL_STORAGE_HIGH   CreateInstanceReqStorageSpecCode
	DMS_PHYSICAL_STORAGE_ULTRA  CreateInstanceReqStorageSpecCode
}

func GetCreateInstanceReqStorageSpecCodeEnum() CreateInstanceReqStorageSpecCodeEnum {
	return CreateInstanceReqStorageSpecCodeEnum{
		DMS_PHYSICAL_STORAGE_NORMAL: CreateInstanceReqStorageSpecCode{
			value: "dms.physical.storage.normal",
		},
		DMS_PHYSICAL_STORAGE_HIGH: CreateInstanceReqStorageSpecCode{
			value: "dms.physical.storage.high",
		},
		DMS_PHYSICAL_STORAGE_ULTRA: CreateInstanceReqStorageSpecCode{
			value: "dms.physical.storage.ultra",
		},
	}
}

func (c CreateInstanceReqStorageSpecCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateInstanceReqStorageSpecCode) UnmarshalJSON(b []byte) error {
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
