/*
 * BMS
 *
 * BMS Open API
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

// server字段数据结构说明
type ChangeBaremetalNameResponsesServers struct {
	// 裸金属服务器名称
	Name *string `json:"name,omitempty"`
	// 裸金属服务器唯一标识ID
	Id *string `json:"id,omitempty"`
	// 裸金属服务器当前状态。ACTIVE：运行中/正在关机/删除中BUILD：创建中ERROR：故障HARD_REBOOT：强制重启中REBOOT：重启中 SHUTOFF：关机/正在开机/删除中/重建中/重装操作系统中/重装操作系统失败/冻结
	Status *ChangeBaremetalNameResponsesServersStatus `json:"status,omitempty"`
	// 裸金属服务器创建时间。时间戳格式为ISO 8601：YYYY-MM-DDTHH:MM:SSZ，例如：2019-05-22T03:30:52Z
	Created *sdktime.SdkTime `json:"created,omitempty"`
	// 裸金属服务器上一次更新时间。时间戳格式为ISO 8601：YYYY-MM-DDTHH:MM:SSZ，例如：2019-05-22T04:30:52Z
	Updated *sdktime.SdkTime `json:"updated,omitempty"`
	Flavor  *FlavorInfo      `json:"flavor,omitempty"`
	Image   *Image           `json:"image,omitempty"`
	// 裸金属服务器所属租户ID，格式为UUID。该参数和project_id表示相同的概念。
	TenantId *string `json:"tenant_id,omitempty"`
	// SSH密钥名称
	KeyName *string `json:"key_name,omitempty"`
	// 裸金属服务器所属用户ID。
	UserId   *string        `json:"user_id,omitempty"`
	Metadata *MetadataInfos `json:"metadata,omitempty"`
	// 裸金属服务器的主机ID
	HostId    *string    `json:"hostId,omitempty"`
	Addresses *Addresses `json:"addresses,omitempty"`
	// 裸金属服务器所属安全组列表。
	SecurityGroups *[]SecurityGroups `json:"security_groups,omitempty"`
	// 裸金属服务器相关信息快捷链接
	Links *[]Links `json:"links,omitempty"`
	// 扩展属性，磁盘配置方式，取值为如下两种：MANUAL：API使用镜像中的分区方案和文件系统创建裸金属服务器。如果目标flavor磁盘较大，则API不会对剩余磁盘空间进行分区。AUTO：API使用与目标flavor磁盘大小相同的单个分区创建裸金属服务器，API会自动调整文件系统以适应整个分区。
	OSDCFdiskConfig *ChangeBaremetalNameResponsesServersOSDCFdiskConfig `json:"OS-DCF:diskConfig,omitempty"`
	// 扩展属性，可用分区编码。
	OSEXTAZavailabilityZone *string `json:"OS-EXT-AZ:availability_zone,omitempty"`
	// 扩展属性，裸金属服务器宿主名称
	OSEXTSRVATTRhost *string `json:"OS-EXT-SRV-ATTR:host,omitempty"`
	// 扩展属性，hypervisor主机名称，由Nova virt驱动提供
	OSEXTSRVATTRhypervisorHostname *string `json:"OS-EXT-SRV-ATTR:hypervisor_hostname,omitempty"`
	// 扩展属性，裸金属服务器实例ID
	OSEXTSRVATTRinstanceName *string `json:"OS-EXT-SRV-ATTR:instance_name,omitempty"`
	// 扩展属性，裸金属服务器电源状态。例如：0表示“NO STATE”1表示“RUNNING”4表示“SHUTDOWN”
	OSEXTSTSpowerState *ChangeBaremetalNameResponsesServersOSEXTSTSpowerState `json:"OS-EXT-STS:power_state,omitempty"`
	// 扩展属性，裸金属服务器任务状态。例如：rebooting表示重启中reboot_started表示普通重启reboot_started_hard表示强制重启powering-off表示关机中powering-on表示开机中rebuilding表示重建中scheduling表示调度中deleting表示删除中
	OSEXTSTStaskState *ChangeBaremetalNameResponsesServersOSEXTSTStaskState `json:"OS-EXT-STS:task_state,omitempty"`
	// 扩展属性，裸金属服务器状态。例如：RUNNING表示运行中SHUTOFF表示关机SUSPENDED表示暂停REBOOT表示重启
	OSEXTSTSvmState *ChangeBaremetalNameResponsesServersOSEXTSTSvmState `json:"OS-EXT-STS:vm_state,omitempty"`
	// 扩展属性，裸金属服务器启动时间。时间戳格式为ISO 8601，例如：2019-05-25T03:40:25.000000
	OSSRVUSGlaunchedAt *sdktime.SdkTime `json:"OS-SRV-USG:launched_at,omitempty"`
	// 扩展属性，裸金属服务器关闭时间。时间戳格式为ISO 8601，例如：2019-06-25T03:40:25.000000
	OSSRVUSGterminatedAt *sdktime.SdkTime `json:"OS-SRV-USG:terminated_at,omitempty"`
	// 裸金属服务器挂载的云硬盘信息。详情请参见表 os-extended-volumes:volumes_attached字段数据结构说明。
	OsExtendedVolumesvolumesAttached *[]OsExtendedVolumes `json:"os-extended-volumes:volumes_attached,omitempty"`
	// 预留属性
	AccessIPv4 *string `json:"accessIPv4,omitempty"`
	// 预留属性
	AccessIPv6 *string `json:"accessIPv6,omitempty"`
	Fault      *Fault  `json:"fault,omitempty"`
	// config drive信息
	ConfigDrive *string `json:"config_drive,omitempty"`
	// 预留属性
	Progress *int32 `json:"progress,omitempty"`
	// 裸金属服务器的描述信息。
	Description *string `json:"description,omitempty"`
	// 裸金属服务器宿主机状态。UP：服务正常UNKNOWN：状态未知DOWN：服务异常MAINTENANCE：维护状态空字符串：裸金属服务器无主机信息
	HostStatus *ChangeBaremetalNameResponsesServersHostStatus `json:"host_status,omitempty"`
	// 裸金属服务器的主机名
	OSEXTSRVATTRhostname *string `json:"OS-EXT-SRV-ATTR:hostname,omitempty"`
	// 批量创建场景，裸金属服务器的预留ID。当批量创建裸金属服务器时，这些服务器将拥有相同的reservation_id。您可以使用6.3.3-查询裸金属服务器详情列表API并指定reservation_id来过滤查询同一批创建的所有裸金属服务器。
	OSEXTSRVATTRreservationId *string `json:"OS-EXT-SRV-ATTR:reservation_id,omitempty"`
	// 批量创建场景，裸金属服务器的启动顺序
	OSEXTSRVATTRlaunchIndex *int32 `json:"OS-EXT-SRV-ATTR:launch_index,omitempty"`
	// 若使用AMI格式的镜像，则表示kernel image的UUID；否则，留空
	OSEXTSRVATTRkernelId *string `json:"OS-EXT-SRV-ATTR:kernel_id,omitempty"`
	// 若使用AMI格式镜像，则表示ramdisk image的UUID；否则，留空。
	OSEXTSRVATTRramdiskId *string `json:"OS-EXT-SRV-ATTR:ramdisk_id,omitempty"`
	// 裸金属服务器系统盘的设备名称，例如“/dev/sdb”。
	OSEXTSRVATTRrootDeviceName *string `json:"OS-EXT-SRV-ATTR:root_device_name,omitempty"`
	// 创建裸金属服务器时指定的user_data。取值为base64编码后的结果或空字符串。
	OSEXTSRVATTRuserData *string `json:"OS-EXT-SRV-ATTR:user_data,omitempty"`
	// 裸金属服务器实例是否为锁定状态。true：锁定false：未锁定
	Locked *bool `json:"locked,omitempty"`
	// 裸金属服务器标签
	Tags *[]string `json:"tags,omitempty"`
}

func (o ChangeBaremetalNameResponsesServers) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ChangeBaremetalNameResponsesServers struct{}"
	}

	return strings.Join([]string{"ChangeBaremetalNameResponsesServers", string(data)}, " ")
}

type ChangeBaremetalNameResponsesServersStatus struct {
	value string
}

type ChangeBaremetalNameResponsesServersStatusEnum struct {
	ACTIVE      ChangeBaremetalNameResponsesServersStatus
	BUILD       ChangeBaremetalNameResponsesServersStatus
	DELETED     ChangeBaremetalNameResponsesServersStatus
	ERROR       ChangeBaremetalNameResponsesServersStatus
	HARD_REBOOT ChangeBaremetalNameResponsesServersStatus
	REBOOT      ChangeBaremetalNameResponsesServersStatus
	REBUILD     ChangeBaremetalNameResponsesServersStatus
	SHUTOFF     ChangeBaremetalNameResponsesServersStatus
}

func GetChangeBaremetalNameResponsesServersStatusEnum() ChangeBaremetalNameResponsesServersStatusEnum {
	return ChangeBaremetalNameResponsesServersStatusEnum{
		ACTIVE: ChangeBaremetalNameResponsesServersStatus{
			value: "ACTIVE",
		},
		BUILD: ChangeBaremetalNameResponsesServersStatus{
			value: "BUILD",
		},
		DELETED: ChangeBaremetalNameResponsesServersStatus{
			value: "DELETED",
		},
		ERROR: ChangeBaremetalNameResponsesServersStatus{
			value: "ERROR",
		},
		HARD_REBOOT: ChangeBaremetalNameResponsesServersStatus{
			value: "HARD_REBOOT",
		},
		REBOOT: ChangeBaremetalNameResponsesServersStatus{
			value: "REBOOT",
		},
		REBUILD: ChangeBaremetalNameResponsesServersStatus{
			value: "REBUILD",
		},
		SHUTOFF: ChangeBaremetalNameResponsesServersStatus{
			value: "SHUTOFF",
		},
	}
}

func (c ChangeBaremetalNameResponsesServersStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ChangeBaremetalNameResponsesServersStatus) UnmarshalJSON(b []byte) error {
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

type ChangeBaremetalNameResponsesServersOSDCFdiskConfig struct {
	value string
}

type ChangeBaremetalNameResponsesServersOSDCFdiskConfigEnum struct {
	MANUAL ChangeBaremetalNameResponsesServersOSDCFdiskConfig
	AUTO   ChangeBaremetalNameResponsesServersOSDCFdiskConfig
}

func GetChangeBaremetalNameResponsesServersOSDCFdiskConfigEnum() ChangeBaremetalNameResponsesServersOSDCFdiskConfigEnum {
	return ChangeBaremetalNameResponsesServersOSDCFdiskConfigEnum{
		MANUAL: ChangeBaremetalNameResponsesServersOSDCFdiskConfig{
			value: "MANUAL",
		},
		AUTO: ChangeBaremetalNameResponsesServersOSDCFdiskConfig{
			value: "AUTO",
		},
	}
}

func (c ChangeBaremetalNameResponsesServersOSDCFdiskConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ChangeBaremetalNameResponsesServersOSDCFdiskConfig) UnmarshalJSON(b []byte) error {
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

type ChangeBaremetalNameResponsesServersOSEXTSTSpowerState struct {
	value int32
}

type ChangeBaremetalNameResponsesServersOSEXTSTSpowerStateEnum struct {
	E_0 ChangeBaremetalNameResponsesServersOSEXTSTSpowerState
	E_1 ChangeBaremetalNameResponsesServersOSEXTSTSpowerState
	E_4 ChangeBaremetalNameResponsesServersOSEXTSTSpowerState
}

func GetChangeBaremetalNameResponsesServersOSEXTSTSpowerStateEnum() ChangeBaremetalNameResponsesServersOSEXTSTSpowerStateEnum {
	return ChangeBaremetalNameResponsesServersOSEXTSTSpowerStateEnum{
		E_0: ChangeBaremetalNameResponsesServersOSEXTSTSpowerState{
			value: 0,
		}, E_1: ChangeBaremetalNameResponsesServersOSEXTSTSpowerState{
			value: 1,
		}, E_4: ChangeBaremetalNameResponsesServersOSEXTSTSpowerState{
			value: 4,
		},
	}
}

func (c ChangeBaremetalNameResponsesServersOSEXTSTSpowerState) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ChangeBaremetalNameResponsesServersOSEXTSTSpowerState) UnmarshalJSON(b []byte) error {
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

type ChangeBaremetalNameResponsesServersOSEXTSTStaskState struct {
	value string
}

type ChangeBaremetalNameResponsesServersOSEXTSTStaskStateEnum struct {
	REBOOTING           ChangeBaremetalNameResponsesServersOSEXTSTStaskState
	REBOOT_STARTED      ChangeBaremetalNameResponsesServersOSEXTSTStaskState
	REBOOT_STARTED_HARD ChangeBaremetalNameResponsesServersOSEXTSTStaskState
	POWERING_OFF        ChangeBaremetalNameResponsesServersOSEXTSTStaskState
	POWERING_ON         ChangeBaremetalNameResponsesServersOSEXTSTStaskState
	REBUILDING          ChangeBaremetalNameResponsesServersOSEXTSTStaskState
	SCHEDULING          ChangeBaremetalNameResponsesServersOSEXTSTStaskState
	DELETING            ChangeBaremetalNameResponsesServersOSEXTSTStaskState
}

func GetChangeBaremetalNameResponsesServersOSEXTSTStaskStateEnum() ChangeBaremetalNameResponsesServersOSEXTSTStaskStateEnum {
	return ChangeBaremetalNameResponsesServersOSEXTSTStaskStateEnum{
		REBOOTING: ChangeBaremetalNameResponsesServersOSEXTSTStaskState{
			value: "rebooting",
		},
		REBOOT_STARTED: ChangeBaremetalNameResponsesServersOSEXTSTStaskState{
			value: "reboot_started",
		},
		REBOOT_STARTED_HARD: ChangeBaremetalNameResponsesServersOSEXTSTStaskState{
			value: "reboot_started_hard",
		},
		POWERING_OFF: ChangeBaremetalNameResponsesServersOSEXTSTStaskState{
			value: "powering-off",
		},
		POWERING_ON: ChangeBaremetalNameResponsesServersOSEXTSTStaskState{
			value: "powering-on",
		},
		REBUILDING: ChangeBaremetalNameResponsesServersOSEXTSTStaskState{
			value: "rebuilding",
		},
		SCHEDULING: ChangeBaremetalNameResponsesServersOSEXTSTStaskState{
			value: "scheduling",
		},
		DELETING: ChangeBaremetalNameResponsesServersOSEXTSTStaskState{
			value: "deleting",
		},
	}
}

func (c ChangeBaremetalNameResponsesServersOSEXTSTStaskState) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ChangeBaremetalNameResponsesServersOSEXTSTStaskState) UnmarshalJSON(b []byte) error {
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

type ChangeBaremetalNameResponsesServersOSEXTSTSvmState struct {
	value string
}

type ChangeBaremetalNameResponsesServersOSEXTSTSvmStateEnum struct {
	RUNNING   ChangeBaremetalNameResponsesServersOSEXTSTSvmState
	SHUTOFF   ChangeBaremetalNameResponsesServersOSEXTSTSvmState
	SUSPENDED ChangeBaremetalNameResponsesServersOSEXTSTSvmState
	REBOOT    ChangeBaremetalNameResponsesServersOSEXTSTSvmState
}

func GetChangeBaremetalNameResponsesServersOSEXTSTSvmStateEnum() ChangeBaremetalNameResponsesServersOSEXTSTSvmStateEnum {
	return ChangeBaremetalNameResponsesServersOSEXTSTSvmStateEnum{
		RUNNING: ChangeBaremetalNameResponsesServersOSEXTSTSvmState{
			value: "RUNNING",
		},
		SHUTOFF: ChangeBaremetalNameResponsesServersOSEXTSTSvmState{
			value: "SHUTOFF",
		},
		SUSPENDED: ChangeBaremetalNameResponsesServersOSEXTSTSvmState{
			value: "SUSPENDED",
		},
		REBOOT: ChangeBaremetalNameResponsesServersOSEXTSTSvmState{
			value: "REBOOT",
		},
	}
}

func (c ChangeBaremetalNameResponsesServersOSEXTSTSvmState) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ChangeBaremetalNameResponsesServersOSEXTSTSvmState) UnmarshalJSON(b []byte) error {
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

type ChangeBaremetalNameResponsesServersHostStatus struct {
	value string
}

type ChangeBaremetalNameResponsesServersHostStatusEnum struct {
	UP          ChangeBaremetalNameResponsesServersHostStatus
	UNKNOWN     ChangeBaremetalNameResponsesServersHostStatus
	DOWN        ChangeBaremetalNameResponsesServersHostStatus
	MAINTENANCE ChangeBaremetalNameResponsesServersHostStatus
}

func GetChangeBaremetalNameResponsesServersHostStatusEnum() ChangeBaremetalNameResponsesServersHostStatusEnum {
	return ChangeBaremetalNameResponsesServersHostStatusEnum{
		UP: ChangeBaremetalNameResponsesServersHostStatus{
			value: "UP",
		},
		UNKNOWN: ChangeBaremetalNameResponsesServersHostStatus{
			value: "UNKNOWN",
		},
		DOWN: ChangeBaremetalNameResponsesServersHostStatus{
			value: "DOWN",
		},
		MAINTENANCE: ChangeBaremetalNameResponsesServersHostStatus{
			value: "MAINTENANCE",
		},
	}
}

func (c ChangeBaremetalNameResponsesServersHostStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ChangeBaremetalNameResponsesServersHostStatus) UnmarshalJSON(b []byte) error {
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
