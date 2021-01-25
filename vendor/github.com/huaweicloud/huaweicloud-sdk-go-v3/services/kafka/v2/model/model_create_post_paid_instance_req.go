/*
 * Kafka
 *
 * Kafka Document API
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
type CreatePostPaidInstanceReq struct {
	// 实例名称。  由英文字符开头，只能由英文字母、数字、中划线、下划线组成，长度为4~64的字符。
	Name string `json:"name"`
	// 实例的描述信息。  长度不超过1024的字符串。  > \\与\"在json报文中属于特殊字符，如果参数值中需要显示\\或者\"字符，请在字符前增加转义字符\\，比如\\\\或者\\\"。
	Description *string `json:"description,omitempty"`
	// 消息引擎。取值填写为：kafka。
	Engine CreatePostPaidInstanceReqEngine `json:"engine"`
	// 消息引擎的版本。取值填写为：1.1.0和2.3.0。
	EngineVersion CreatePostPaidInstanceReqEngineVersion `json:"engine_version"`
	// Kafka实例的基准带宽，表示单位时间内传送的最大数据量，单位MB。 取值范围：   - 100MB   - 300MB   - 600MB   - 1200MB
	Specification CreatePostPaidInstanceReqSpecification `json:"specification"`
	// 消息存储空间，单位GB。   - Kafka实例规格为100MB时，存储空间取值范围600GB ~ 90000GB。   - Kafka实例规格为300MB时，存储空间取值范围1200GB ~ 90000GB。   - Kafka实例规格为600MB时，存储空间取值范围2400GB ~ 90000GB。   - Kafka实例规格为1200MB，存储空间取值范围4800GB ~ 90000GB。
	StorageSpace int32 `json:"storage_space"`
	// Kafka实例的最大分区数量。   - 参数specification为100MB时，取值300   - 参数specification为300MB时，取值900   - 参数specification为600MB时，取值1800   - 参数specification为1200MB时，取值1800
	PartitionNum CreatePostPaidInstanceReqPartitionNum `json:"partition_num"`
	// 当ssl_enable为true时，该参数必选，ssl_enable为false时，该参数无效。  认证用户名，只能由英文字母、数字、中划线组成，长度为4~64的字符。
	AccessUser *string `json:"access_user,omitempty"`
	// 当ssl_enable为true时，该参数必选，ssl_enable为false时，该参数无效。  实例的认证密码。  复杂度要求： - 输入长度为8到32位的字符串。 - 必须包含如下四种字符中的两种组合：   - 小写字母   - 大写字母   - 数字   - 特殊字符包括（`~!@#$%^&*()-_=+\\|[{}]:'\",<.>/?）
	Password *string `json:"password,omitempty"`
	// 虚拟私有云ID。  获取方法如下： - 方法1：登录虚拟私有云服务的控制台界面，在虚拟私有云的详情页面查找VPC ID。 - 方法2：通过虚拟私有云服务的API接口查询，具体操作可参考[查询VPC列表](https://support.huaweicloud.com/api-vpc/vpc_api01_0003.html)。
	VpcId string `json:"vpc_id"`
	// 指定实例所属的安全组。  获取方法如下： - 方法1：登录虚拟私有云服务的控制台界面，在安全组的详情页面查找安全组ID。 - 方法2：通过虚拟私有云服务的API接口查询，具体操作可参考[查询安全组列表](https://support.huaweicloud.com/api-vpc/vpc_sg01_0002.html)。
	SecurityGroupId string `json:"security_group_id"`
	// 子网信息。  获取方法如下： - 方法1：登录虚拟私有云服务的控制台界面，单击VPC下的子网，进入子网详情页面，查找网络ID。 - 方法2：通过虚拟私有云服务的API接口查询，具体操作可参考[查询子网列表](https://support.huaweicloud.com/api-vpc/vpc_subnet01_0003.html)。
	SubnetId string `json:"subnet_id"`
	// 创建节点到指定且有资源的可用区ID。该参数不能为空数组或者数组的值为空，详情请参考[查询可用区信息](https://support.huaweicloud.com/api-kafka/ListAvailableZones.html)查询得到。在查询时，请注意查看该可用区是否有资源。  创建Kafka实例，支持节点部署在1个或3个及3个以上的可用区。在为节点指定可用区时，用逗号分隔开。
	AvailableZones []string `json:"available_zones"`
	// 产品标识。  获取方法，请参考查询[产品规格列表](https://support.huaweicloud.com/api-kafka/ListProducts.html)。
	ProductId string `json:"product_id"`
	// 表示登录Kafka Manager的用户名。只能由英文字母、数字、中划线组成，长度为4~64的字符。
	KafkaManagerUser string `json:"kafka_manager_user"`
	// 表示登录Kafka Manager的密码。  复杂度要求：   - 输入长度为8到32位的字符串。   - 必须包含如下四种字符中的两种组合：       - 小写字母       - 大写字母       - 数字       - 特殊字符包括（`~!@#$%^&*()-_=+\\|[{}]:'\",<.>/?）
	KafkaManagerPassword string `json:"kafka_manager_password"`
	// 维护时间窗开始时间，格式为HH:mm。 - 维护时间窗开始和结束时间必须为指定的时间段，可参考[查询维护时间窗时间段](https://support.huaweicloud.com/api-kafka/ShowMaintainWindows.html)获取。 - 开始时间必须为22:00、02:00、06:00、10:00、14:00和18:00。 - 该参数不能单独为空，若该值为空，则结束时间也为空。系统分配一个默认开始时间02:00。
	MaintainBegin *string `json:"maintain_begin,omitempty"`
	// 维护时间窗结束时间，格式为HH:mm。 - 维护时间窗开始和结束时间必须为指定的时间段，可参考[查询维护时间窗时间段](https://support.huaweicloud.com/api-kafka/ShowMaintainWindows.html)获取。 - 结束时间在开始时间基础上加四个小时，即当开始时间为22:00时，结束时间为02:00。 - 该参数不能单独为空，若该值为空，则开始时间也为空，系统分配一个默认结束时间06:00。
	MaintainEnd *string `json:"maintain_end,omitempty"`
	// 是否开启公网访问功能。默认不开启公网。 - true：开启 - false：不开启
	EnablePublicip *bool `json:"enable_publicip,omitempty"`
	// 表示公网带宽，单位是Mbit/s。 取值范围： - Kafka实例规格为100MB时，公网带宽取值范围3到900，且必须为实例节点个数的倍数。 - Kafka实例规格为300MB时，公网带宽取值范围3到900，且必须为实例节点个数的倍数。 - Kafka实例规格为600MB时，公网带宽取值范围4到1200，且必须为实例节点个数的倍数。 - Kafka实例规格为1200MB时，公网带宽取值范围8到2400，且必须为实例节点个数的倍数。
	PublicBandwidth *int32 `json:"public_bandwidth,omitempty"`
	// 实例绑定的弹性IP地址的ID。  如果开启了公网访问功能（即enable_publicip为true），该字段为必选。
	PublicipId *string `json:"publicip_id,omitempty"`
	// 是否打开SSL加密访问。 - true：打开SSL加密访问。 - false：不打开SSL加密访问。
	SslEnable *bool `json:"ssl_enable,omitempty"`
	// 磁盘的容量到达容量阈值后，对于消息的处理策略。  取值如下： - produce_reject：表示拒绝消息写入。 - time_base：表示自动删除最老消息。
	RetentionPolicy *CreatePostPaidInstanceReqRetentionPolicy `json:"retention_policy,omitempty"`
	// 是否开启消息转储功能。  默认不开启消息转储。
	ConnectorEnable *bool `json:"connector_enable,omitempty"`
	// 是否打开kafka自动创建topic功能。 - true：开启 - false：关闭  当您选择开启，表示生产或消费一个未创建的Topic时，会自动创建一个包含3个分区和3个副本的Topic。
	EnableAutoTopic *bool `json:"enable_auto_topic,omitempty"`
	// 存储IO规格。如何选择磁盘类型请参考[磁盘类型及性能介绍](https://support.huaweicloud.com/productdesc-evs/zh-cn_topic_0044524691.html)。 取值范围：   - 参数specification为100MB时，取值dms.physical.storage.high或者dms.physical.storage.ultra   - 参数specification为300MB时，取值dms.physical.storage.high或者dms.physical.storage.ultra   - 参数specification为600MB时，取值dms.physical.storage.ultra   - 参数specification为1200MB时，取值dms.physical.storage.ultra存储IO规格。如何选择磁盘类型请参考磁盘类型及性能介绍。
	StorageSpecCode CreatePostPaidInstanceReqStorageSpecCode `json:"storage_spec_code"`
	// 企业项目ID。若为企业项目账号，该参数必填。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	// 标签列表。
	Tags *[]CreatePostPaidInstanceReqTags `json:"tags,omitempty"`
}

func (o CreatePostPaidInstanceReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePostPaidInstanceReq struct{}"
	}

	return strings.Join([]string{"CreatePostPaidInstanceReq", string(data)}, " ")
}

type CreatePostPaidInstanceReqEngine struct {
	value string
}

type CreatePostPaidInstanceReqEngineEnum struct {
	KAFKA CreatePostPaidInstanceReqEngine
}

func GetCreatePostPaidInstanceReqEngineEnum() CreatePostPaidInstanceReqEngineEnum {
	return CreatePostPaidInstanceReqEngineEnum{
		KAFKA: CreatePostPaidInstanceReqEngine{
			value: "kafka",
		},
	}
}

func (c CreatePostPaidInstanceReqEngine) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreatePostPaidInstanceReqEngine) UnmarshalJSON(b []byte) error {
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

type CreatePostPaidInstanceReqEngineVersion struct {
	value string
}

type CreatePostPaidInstanceReqEngineVersionEnum struct {
	E_1_1_0 CreatePostPaidInstanceReqEngineVersion
	E_2_3_0 CreatePostPaidInstanceReqEngineVersion
}

func GetCreatePostPaidInstanceReqEngineVersionEnum() CreatePostPaidInstanceReqEngineVersionEnum {
	return CreatePostPaidInstanceReqEngineVersionEnum{
		E_1_1_0: CreatePostPaidInstanceReqEngineVersion{
			value: "1.1.0",
		},
		E_2_3_0: CreatePostPaidInstanceReqEngineVersion{
			value: "2.3.0",
		},
	}
}

func (c CreatePostPaidInstanceReqEngineVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreatePostPaidInstanceReqEngineVersion) UnmarshalJSON(b []byte) error {
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

type CreatePostPaidInstanceReqSpecification struct {
	value string
}

type CreatePostPaidInstanceReqSpecificationEnum struct {
	E_100_MB  CreatePostPaidInstanceReqSpecification
	E_300_MB  CreatePostPaidInstanceReqSpecification
	E_600_MB  CreatePostPaidInstanceReqSpecification
	E_1200_MB CreatePostPaidInstanceReqSpecification
}

func GetCreatePostPaidInstanceReqSpecificationEnum() CreatePostPaidInstanceReqSpecificationEnum {
	return CreatePostPaidInstanceReqSpecificationEnum{
		E_100_MB: CreatePostPaidInstanceReqSpecification{
			value: "100MB",
		},
		E_300_MB: CreatePostPaidInstanceReqSpecification{
			value: "300MB",
		},
		E_600_MB: CreatePostPaidInstanceReqSpecification{
			value: "600MB",
		},
		E_1200_MB: CreatePostPaidInstanceReqSpecification{
			value: "1200MB",
		},
	}
}

func (c CreatePostPaidInstanceReqSpecification) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreatePostPaidInstanceReqSpecification) UnmarshalJSON(b []byte) error {
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

type CreatePostPaidInstanceReqPartitionNum struct {
	value int32
}

type CreatePostPaidInstanceReqPartitionNumEnum struct {
	E_300  CreatePostPaidInstanceReqPartitionNum
	E_900  CreatePostPaidInstanceReqPartitionNum
	E_1800 CreatePostPaidInstanceReqPartitionNum
}

func GetCreatePostPaidInstanceReqPartitionNumEnum() CreatePostPaidInstanceReqPartitionNumEnum {
	return CreatePostPaidInstanceReqPartitionNumEnum{
		E_300: CreatePostPaidInstanceReqPartitionNum{
			value: 300,
		}, E_900: CreatePostPaidInstanceReqPartitionNum{
			value: 900,
		}, E_1800: CreatePostPaidInstanceReqPartitionNum{
			value: 1800,
		},
	}
}

func (c CreatePostPaidInstanceReqPartitionNum) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreatePostPaidInstanceReqPartitionNum) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("int32")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(int32)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to int32 error")
	}
}

type CreatePostPaidInstanceReqRetentionPolicy struct {
	value string
}

type CreatePostPaidInstanceReqRetentionPolicyEnum struct {
	TIME_BASE      CreatePostPaidInstanceReqRetentionPolicy
	PRODUCE_REJECT CreatePostPaidInstanceReqRetentionPolicy
}

func GetCreatePostPaidInstanceReqRetentionPolicyEnum() CreatePostPaidInstanceReqRetentionPolicyEnum {
	return CreatePostPaidInstanceReqRetentionPolicyEnum{
		TIME_BASE: CreatePostPaidInstanceReqRetentionPolicy{
			value: "time_base",
		},
		PRODUCE_REJECT: CreatePostPaidInstanceReqRetentionPolicy{
			value: "produce_reject",
		},
	}
}

func (c CreatePostPaidInstanceReqRetentionPolicy) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreatePostPaidInstanceReqRetentionPolicy) UnmarshalJSON(b []byte) error {
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

type CreatePostPaidInstanceReqStorageSpecCode struct {
	value string
}

type CreatePostPaidInstanceReqStorageSpecCodeEnum struct {
	DMS_PHYSICAL_STORAGE_NORMAL CreatePostPaidInstanceReqStorageSpecCode
	DMS_PHYSICAL_STORAGE_HIGH   CreatePostPaidInstanceReqStorageSpecCode
	DMS_PHYSICAL_STORAGE_ULTRA  CreatePostPaidInstanceReqStorageSpecCode
}

func GetCreatePostPaidInstanceReqStorageSpecCodeEnum() CreatePostPaidInstanceReqStorageSpecCodeEnum {
	return CreatePostPaidInstanceReqStorageSpecCodeEnum{
		DMS_PHYSICAL_STORAGE_NORMAL: CreatePostPaidInstanceReqStorageSpecCode{
			value: "dms.physical.storage.normal",
		},
		DMS_PHYSICAL_STORAGE_HIGH: CreatePostPaidInstanceReqStorageSpecCode{
			value: "dms.physical.storage.high",
		},
		DMS_PHYSICAL_STORAGE_ULTRA: CreatePostPaidInstanceReqStorageSpecCode{
			value: "dms.physical.storage.ultra",
		},
	}
}

func (c CreatePostPaidInstanceReqStorageSpecCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreatePostPaidInstanceReqStorageSpecCode) UnmarshalJSON(b []byte) error {
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
