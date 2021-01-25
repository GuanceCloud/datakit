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

type AppAuthResp struct {
	// API编号
	ApiId      *string         `json:"api_id,omitempty"`
	AuthResult *AuthResultResp `json:"auth_result,omitempty"`
	// 授权时间
	AuthTime *sdktime.SdkTime `json:"auth_time,omitempty"`
	// 授权关系编号
	Id *string `json:"id,omitempty"`
	// APP编号
	AppId *string `json:"app_id,omitempty"`
	// 授权者 - PROVIDER：API提供者授权 - CONSUMER：API消费者授权
	AuthRole *AppAuthRespAuthRole `json:"auth_role,omitempty"`
	// 授权通道类型 - NORMAL：普通通道 - GREEN：绿色通道  暂不支持，默认NORMAL
	AuthTunnel *AppAuthRespAuthTunnel `json:"auth_tunnel,omitempty"`
	// 绿色通道的白名单配置
	AuthWhitelist *[]string `json:"auth_whitelist,omitempty"`
	// 绿色通道的黑名单配置
	AuthBlacklist *[]string `json:"auth_blacklist,omitempty"`
}

func (o AppAuthResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AppAuthResp struct{}"
	}

	return strings.Join([]string{"AppAuthResp", string(data)}, " ")
}

type AppAuthRespAuthRole struct {
	value string
}

type AppAuthRespAuthRoleEnum struct {
	PROVIDER AppAuthRespAuthRole
	CONSUMER AppAuthRespAuthRole
}

func GetAppAuthRespAuthRoleEnum() AppAuthRespAuthRoleEnum {
	return AppAuthRespAuthRoleEnum{
		PROVIDER: AppAuthRespAuthRole{
			value: "PROVIDER",
		},
		CONSUMER: AppAuthRespAuthRole{
			value: "CONSUMER",
		},
	}
}

func (c AppAuthRespAuthRole) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *AppAuthRespAuthRole) UnmarshalJSON(b []byte) error {
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

type AppAuthRespAuthTunnel struct {
	value string
}

type AppAuthRespAuthTunnelEnum struct {
	NORMAL AppAuthRespAuthTunnel
	GREEN  AppAuthRespAuthTunnel
}

func GetAppAuthRespAuthTunnelEnum() AppAuthRespAuthTunnelEnum {
	return AppAuthRespAuthTunnelEnum{
		NORMAL: AppAuthRespAuthTunnel{
			value: "NORMAL",
		},
		GREEN: AppAuthRespAuthTunnel{
			value: "GREEN",
		},
	}
}

func (c AppAuthRespAuthTunnel) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *AppAuthRespAuthTunnel) UnmarshalJSON(b []byte) error {
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
