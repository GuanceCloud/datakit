/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// server字段数据结构说明
type CreateServers struct {
	// 裸金属服务器使用的镜像ID或者镜像资源的URL。ID格式为通用唯一识别码（Universally Unique Identifier，简称UUID）。镜像ID可以从镜像服务控制台获取，或者参考《镜像服务API参考》的“查询镜像列表”章节查询。在使用“查询镜像列表”API查询时，可以添加过滤字段“?virtual_env_type=Ironic”来筛选裸金属服务器镜像。
	ImageRef string `json:"imageRef"`
	// 裸金属服务器使用的规格ID，格式为physical.x.x。规格ID可以从裸金属服务器控制台获取，也可以通过7.7.1-查询裸金属服务器规格信息列表（OpenStack原生）API查询。 说明：裸金属服务器规格与镜像间的约束关系请参见裸金属服务器类型与支持的操作系统版本。对于physical.x.x.hba类型的规格，申请的租户只能是DeC租户，且只能挂载DESS卷。
	FlavorRef string `json:"flavorRef"`
	// 裸金属服务器名称。取值范围：只能由中文字符、英文字母（a~z，A~Z）、数字（0~9）、下划线（_）、中划线（-）、点（.）组成，且长度为[1-63]个字符。创建的裸金属服务器数量大于1时，为区分不同裸金属服务器，创建过程中系统会自动在名称后加“-0000”的类似标记。故此时名称的长度为[1-58]个字符。
	Name     string        `json:"name"`
	Metadata *MetaDataInfo `json:"metadata"`
	// 创建裸金属服务器过程中待注入的用户数据。支持注入文本、文本文件或gzip文件。约束：注入内容，需要进行base64格式编码。注入内容（编码之前的内容）最大长度32KB。当key_name没有指定时，user_data注入的数据默认为裸金属服务器root帐户的登录密码。创建密码方式鉴权的Linux裸金属服务器时为必填项，为root用户注入自定义初始化密码。建议密码复杂度如下：长度为8-26位。密码至少必须包含大写字母（A-Z）、小写字母（a-z）、数字（0-9）和特殊字符（!@$%^-_=+[{}]:,./?）中的三种。示例：使用明文密码（存在安全风险），以密码cloud.1234为例：#!/bin/bash echo 'root:Cloud.1234' | chpasswd ;使用密码：#!/bin/bash echo 'root:$6$V6azyeLwcD3CHlpY$BN3VVq18fmCkj66B4zdHLWevqcxlig' | chpasswd -e其中，$6$V6azyeLwcD3CHlpY$BN3VVq18fmCkj66B4zdHLWevqcxlig为密文密码
	UserData *string `json:"user_data,omitempty"`
	// 如果需要使用密码方式登录裸金属服务器，可使用adminPass字段指定裸金属服务器管理员帐户初始登录密码。其中，Linux管理员帐户为root，Windows管理员帐户为Administrator。密码复杂度要求：长度为8-26位。密码至少必须包含大写字母、小写字母、数字和特殊字符（!@$%^-_=+[{}]:,./?）中的三种。Linux系统密码不能包含用户名或用户名的逆序。Windows系统密码不能包含用户名或用户名的逆序，不能包含用户名中超过两个连续字符的部分。
	AdminPass *string `json:"adminPass,omitempty"`
	// 扩展属性，指定密钥的名称。如果需要使用SSH密钥方式登录裸金属服务器，请指定已有密钥的名称。密钥可以通过7.10.3-创建和导入SSH密钥（OpenStack原生）API创建，或者使用7.10.1-查询SSH密钥列表（OpenStack原生）API查询已有的密钥。约束：当key_name和user_data同时指定时，user_data只能用做用户数据注入。Windows裸金属服务器登录时，首先需要将密钥解析为密码，然后通过远程登录工具进行登录。具体请参见“MSTSC密码方式登录”“MSTSC密码方式登录”。
	KeyName *string `json:"key_name,omitempty"`
	// 指定裸金属服务器的安全组。详情请参见表 security_groups字段数据结构说明。
	SecurityGroups *[]SecurityGroupsInfo `json:"security_groups,omitempty"`
	// 指定裸金属服务器的网卡信息。详情请参见表 nics字段数据结构说明。约束：一个裸金属服务器最多挂载2个网卡，参数中第一个网卡会作为裸金属服务器的主网卡。若用户指定了多组网卡参数，需保证各组参数都属于同一VPC。
	Nics []Nics `json:"nics"`
	// 裸金属服务器对应可用区信息，需要指定可用区（AZ）的名称。请参考地区和终端节点获取。
	AvailabilityZone string `json:"availability_zone"`
	// 创建裸金属服务器所属虚拟私有云（VPC），需要指定已有VPC的ID，UUID格式。VPC的ID可以从网络控制台或者参考《虚拟私有云API参考》的“查询VPC”。
	Vpcid    string    `json:"vpcid"`
	Publicip *PublicIp `json:"publicip,omitempty"`
	// 创建裸金属服务器的数量。约束：不传该字段时默认取值为1。租户的配额足够时，最大值为24。
	Count      *int32      `json:"count,omitempty"`
	RootVolume *RootVolume `json:"root_volume,omitempty"`
	// 裸金属服务器对应数据盘相关配置。每一个数据结构代表一个待创建的数据盘。详情请参见表 data_volumes字段数据结构说明。约束：目前裸金属服务器最多可挂载60块云硬盘（包括系统盘和数据盘）。
	DataVolumes    *[]DataVolumes        `json:"data_volumes,omitempty"`
	Extendparam    *ExtendParam          `json:"extendparam"`
	SchedulerHints *CreateSchedulerHints `json:"schedulerHints,omitempty"`
	// 裸金属服务器的标签。详情请参见表 server_tags字段数据结构说明。 说明：创建裸金属服务器时，一台裸金属服务器最多可以添加10个标签。其中，__type_baremetal为系统内部标签，因此实际能添加的标签为9个。
	ServerTags *[]SystemTags `json:"server_tags,omitempty"`
}

func (o CreateServers) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateServers struct{}"
	}

	return strings.Join([]string{"CreateServers", string(data)}, " ")
}
