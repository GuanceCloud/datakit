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
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"
	"strings"
)

type AppAuthBindedApiResp struct {
	// API的编号
	ApiId *string `json:"api_id,omitempty"`
	// APP的名称
	AppName *string `json:"app_name,omitempty"`
	// api授权绑定的环境ID
	EnvId *string `json:"env_id,omitempty"`
	// api授权绑定的环境名称
	EnvName *string `json:"env_name,omitempty"`
	// API绑定的分组ID
	GroupId *string `json:"group_id,omitempty"`
	// API绑定的分组名称
	GroupName *string `json:"group_name,omitempty"`
	// API类型
	ApiType *int32 `json:"api_type,omitempty"`
	// API的名称
	ApiName *string `json:"api_name,omitempty"`
	// APP的编号
	AppId *string `json:"app_id,omitempty"`
	// 授权创建的时间
	AuthTime *sdktime.SdkTime `json:"auth_time,omitempty"`
	// APP的创建者，取值如下： - USER：租户自己创建 - MARKET：API市场分配
	AppCreator *string `json:"app_creator,omitempty"`
	// APP的类型  默认为apig，暂不支持其他类型
	AppType *AppAuthBindedApiRespAppType `json:"app_type,omitempty"`
	// 授权关系编号
	Id *string `json:"id,omitempty"`
	// API的描述信息
	ApiRemark *string `json:"api_remark,omitempty"`
	// 授权者
	AuthRole *string `json:"auth_role,omitempty"`
	// 授权通道类型 - NORMAL：普通通道 - GREEN：绿色通道  暂不支持，默认NORMAL
	AuthTunnel *AppAuthBindedApiRespAuthTunnel `json:"auth_tunnel,omitempty"`
	// 绿色通道的白名单配置
	AuthWhitelist *[]string `json:"auth_whitelist,omitempty"`
	// 绿色通道的黑名单配置
	AuthBlacklist *[]string `json:"auth_blacklist,omitempty"`
}

func (o AppAuthBindedApiResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AppAuthBindedApiResp struct{}"
	}

	return strings.Join([]string{"AppAuthBindedApiResp", string(data)}, " ")
}

type AppAuthBindedApiRespAppType struct {
	value string
}

type AppAuthBindedApiRespAppTypeEnum struct {
	APIG AppAuthBindedApiRespAppType
	ROMA AppAuthBindedApiRespAppType
}

func GetAppAuthBindedApiRespAppTypeEnum() AppAuthBindedApiRespAppTypeEnum {
	return AppAuthBindedApiRespAppTypeEnum{
		APIG: AppAuthBindedApiRespAppType{
			value: "apig",
		},
		ROMA: AppAuthBindedApiRespAppType{
			value: "roma",
		},
	}
}

func (c AppAuthBindedApiRespAppType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *AppAuthBindedApiRespAppType) UnmarshalJSON(b []byte) error {
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

type AppAuthBindedApiRespAuthTunnel struct {
	value string
}

type AppAuthBindedApiRespAuthTunnelEnum struct {
	NORMAL AppAuthBindedApiRespAuthTunnel
	GREEN  AppAuthBindedApiRespAuthTunnel
}

func GetAppAuthBindedApiRespAuthTunnelEnum() AppAuthBindedApiRespAuthTunnelEnum {
	return AppAuthBindedApiRespAuthTunnelEnum{
		NORMAL: AppAuthBindedApiRespAuthTunnel{
			value: "NORMAL",
		},
		GREEN: AppAuthBindedApiRespAuthTunnel{
			value: "GREEN",
		},
	}
}

func (c AppAuthBindedApiRespAuthTunnel) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *AppAuthBindedApiRespAuthTunnel) UnmarshalJSON(b []byte) error {
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
