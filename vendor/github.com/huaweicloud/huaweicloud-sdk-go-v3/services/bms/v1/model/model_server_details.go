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
type ServerDetails struct {
	// 裸金属服务器ID，格式为UUID
	Id string `json:"id"`
	// 创建裸金属服务器的用户ID，格式为UUID。
	UserId *string `json:"user_id,omitempty"`
	// 裸金属服务器名称
	Name string `json:"name"`
	// 裸金属服务器创建时间。时间戳格式为ISO 8601：YYYY-MM-DDTHH:MM:SSZ，例如：2019-05-22T03:30:52Z
	Created *sdktime.SdkTime `json:"created,omitempty"`
	// 裸金属服务器更新时间。时间戳格式为ISO 8601：YYYY-MM-DDTHH:MM:SSZ，例如：2019-05-22T04:30:52Z
	Updated *sdktime.SdkTime `json:"updated,omitempty"`
	// 裸金属服务器所属租户ID，格式为UUID。该参数和project_id表示相同的概念。
	TenantId string `json:"tenant_id"`
	// 裸金属服务器对应的主机ID
	HostId *string `json:"hostId,omitempty"`
	// 裸金属服务器的网络属性。详情请参见表3 addresses数据结构说明。
	Addresses map[string][]AddressInfo `json:"addresses,omitempty"`
	// 裸金属服务器使用的密钥对名称
	KeyName *string      `json:"key_name,omitempty"`
	Image   *ImageInfo   `json:"image,omitempty"`
	Flavor  *FlavorInfos `json:"flavor,omitempty"`
	// 裸金属服务器所属安全组。详情请参见表7 security_groups数据结构说明。
	SecurityGroups *[]SecurityGroupsList `json:"security_groups,omitempty"`
	// 预留属性
	AccessIPv4 *string `json:"accessIPv4,omitempty"`
	// 预留属性
	AccessIPv6 *string `json:"accessIPv6,omitempty"`
	// 裸金属服务器当前状态信息。取值范围：ACTIVE：运行中/正在关机/删除中BUILD：创建中ERROR：故障HARD_REBOOT：强制重启中REBOOT：重启中裸金属服务器当前状态信息。取值范围：ACTIVE：运行中/正在关机/删除中BUILD：创建中ERROR：故障HARD_REBOOT：强制重启中REBOOT：重启中
	Status ServerDetailsStatus `json:"status"`
	// 预留属性
	Progress *int32 `json:"progress,omitempty"`
	// 是否为裸金属服务器配置config drive分区。取值为：True或空字符串
	ConfigDrive *string       `json:"config_drive,omitempty"`
	Metadata    *MetadataList `json:"metadata"`
	// 扩展属性，裸金属服务器当前的任务状态。例如：rebooting：重启中reboot_started：普通重启reboot_started_hard：强制重启powering-off：关机中powering-on：开机中rebuilding：重建中scheduling：调度中deleting：删除中
	OSEXTSTStaskState *ServerDetailsOSEXTSTStaskState `json:"OS-EXT-STS:task_state,omitempty"`
	// 扩展属性，裸金属服务器的稳定状态。例如：active：运行中shutoff：关机suspended：暂停reboot：重启
	OSEXTSTSvmState *ServerDetailsOSEXTSTSvmState `json:"OS-EXT-STS:vm_state,omitempty"`
	// 扩展属性，裸金属服务器宿主名称
	OSEXTSRVATTRhost *string `json:"OS-EXT-SRV-ATTR:host,omitempty"`
	// 扩展属性，裸金属服务器实例ID
	OSEXTSRVATTRinstanceName *string `json:"OS-EXT-SRV-ATTR:instance_name,omitempty"`
	// 扩展属性，裸金属服务器电源状态。例如：0表示“NO STATE”1表示“RUNNING”4表示“SHUTDOWN”
	OSEXTSTSpowerState *ServerDetailsOSEXTSTSpowerState `json:"OS-EXT-STS:power_state,omitempty"`
	// 扩展属性，裸金属服务器所在虚拟化主机名。
	OSEXTSRVATTRhypervisorHostname *string `json:"OS-EXT-SRV-ATTR:hypervisor_hostname,omitempty"`
	// 扩展属性，裸金属服务器所在可用分区名称。
	OSEXTAZavailabilityZone *string `json:"OS-EXT-AZ:availability_zone,omitempty"`
	// 扩展属性，磁盘配置，取值为以下两种：MANUAL：API使用镜像中的分区方案和文件系统创建裸金属服务器。如果目标flavor磁盘较大，则API不会对剩余磁盘空间进行分区。AUTO：API使用与目标flavor磁盘大小相同的单个分区创建裸金属服务器，API会自动调整文件系统以适应整个分区。
	OSDCFdiskConfig *ServerDetailsOSDCFdiskConfig `json:"OS-DCF:diskConfig,omitempty"`
	Fault           *Fault                        `json:"fault,omitempty"`
	// 裸金属服务器启动时间。时间戳格式为ISO 8601，例如：2019-05-22T03:23:59.000000
	OSSRVUSGlaunchedAt *sdktime.SdkTime `json:"OS-SRV-USG:launched_at,omitempty"`
	// 裸金属服务器删除时间。时间戳格式为ISO 8601，例如：2019-05-22T04:23:59.000000
	OSSRVUSGterminatedAt *sdktime.SdkTime `json:"OS-SRV-USG:terminated_at,omitempty"`
	// 挂载到裸金属服务器上的磁盘。详情请参见表9 os-extended-volumes:volumes_attached 数据结构说明。
	OsExtendedVolumesvolumesAttached *[]OsExtendedVolumesInfo `json:"os-extended-volumes:volumes_attached,omitempty"`
	// 裸金属服务器的描述信息
	Description *string `json:"description,omitempty"`
	// 裸金属服务器宿主机状态。UP：服务正常UNKNOWN：状态未知DOWN：服务异常MAINTENANCE：维护状态空字符串：裸金属服务器无主机信息
	HostStatus *ServerDetailsHostStatus `json:"host_status,omitempty"`
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
	// 裸金属服务器系统盘的设备名称，例如“/dev/sda”。
	OSEXTSRVATTRrootDeviceName *string `json:"OS-EXT-SRV-ATTR:root_device_name,omitempty"`
	// 创建裸金属服务器时指定的user_data，取值为base64编码后的结果或空字符串。
	OSEXTSRVATTRuserData *string `json:"OS-EXT-SRV-ATTR:user_data,omitempty"`
	// 裸金属服务器是否为锁定状态。true：锁定false：未锁定
	Locked *bool `json:"locked,omitempty"`
	// 裸金属服务器标签。
	Tags             *[]string       `json:"tags,omitempty"`
	OsschedulerHints *SchedulerHints `json:"os:scheduler_hints,omitempty"`
	// 裸金属服务器所属的企业项目ID
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	// 裸金属服务器系统标签。详情请参见表12 sys_tags数据结构说明。
	SysTags *[]SystemTags `json:"sys_tags,omitempty"`
}

func (o ServerDetails) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ServerDetails struct{}"
	}

	return strings.Join([]string{"ServerDetails", string(data)}, " ")
}

type ServerDetailsStatus struct {
	value string
}

type ServerDetailsStatusEnum struct {
	ACTIVE  ServerDetailsStatus
	BUILD   ServerDetailsStatus
	ERROR   ServerDetailsStatus
	REBOOT  ServerDetailsStatus
	SHUTOFF ServerDetailsStatus
}

func GetServerDetailsStatusEnum() ServerDetailsStatusEnum {
	return ServerDetailsStatusEnum{
		ACTIVE: ServerDetailsStatus{
			value: "ACTIVE",
		},
		BUILD: ServerDetailsStatus{
			value: "BUILD",
		},
		ERROR: ServerDetailsStatus{
			value: "ERROR",
		},
		REBOOT: ServerDetailsStatus{
			value: "REBOOT",
		},
		SHUTOFF: ServerDetailsStatus{
			value: "SHUTOFF",
		},
	}
}

func (c ServerDetailsStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ServerDetailsStatus) UnmarshalJSON(b []byte) error {
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

type ServerDetailsOSEXTSTStaskState struct {
	value string
}

type ServerDetailsOSEXTSTStaskStateEnum struct {
	REBOOTING           ServerDetailsOSEXTSTStaskState
	REBOOT_STARTED      ServerDetailsOSEXTSTStaskState
	REBOOT_STARTED_HARD ServerDetailsOSEXTSTStaskState
	POWERING_OFF        ServerDetailsOSEXTSTStaskState
	POWERING_ON         ServerDetailsOSEXTSTStaskState
	REBUILDING          ServerDetailsOSEXTSTStaskState
	SCHEDULING          ServerDetailsOSEXTSTStaskState
	DELETING            ServerDetailsOSEXTSTStaskState
}

func GetServerDetailsOSEXTSTStaskStateEnum() ServerDetailsOSEXTSTStaskStateEnum {
	return ServerDetailsOSEXTSTStaskStateEnum{
		REBOOTING: ServerDetailsOSEXTSTStaskState{
			value: "rebooting",
		},
		REBOOT_STARTED: ServerDetailsOSEXTSTStaskState{
			value: "reboot_started",
		},
		REBOOT_STARTED_HARD: ServerDetailsOSEXTSTStaskState{
			value: "reboot_started_hard",
		},
		POWERING_OFF: ServerDetailsOSEXTSTStaskState{
			value: "powering-off",
		},
		POWERING_ON: ServerDetailsOSEXTSTStaskState{
			value: "powering-on",
		},
		REBUILDING: ServerDetailsOSEXTSTStaskState{
			value: "rebuilding",
		},
		SCHEDULING: ServerDetailsOSEXTSTStaskState{
			value: "scheduling",
		},
		DELETING: ServerDetailsOSEXTSTStaskState{
			value: "deleting",
		},
	}
}

func (c ServerDetailsOSEXTSTStaskState) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ServerDetailsOSEXTSTStaskState) UnmarshalJSON(b []byte) error {
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

type ServerDetailsOSEXTSTSvmState struct {
	value string
}

type ServerDetailsOSEXTSTSvmStateEnum struct {
	ACTIVE    ServerDetailsOSEXTSTSvmState
	SHUTOFF   ServerDetailsOSEXTSTSvmState
	SUSPENDED ServerDetailsOSEXTSTSvmState
	REBOOT    ServerDetailsOSEXTSTSvmState
}

func GetServerDetailsOSEXTSTSvmStateEnum() ServerDetailsOSEXTSTSvmStateEnum {
	return ServerDetailsOSEXTSTSvmStateEnum{
		ACTIVE: ServerDetailsOSEXTSTSvmState{
			value: "active",
		},
		SHUTOFF: ServerDetailsOSEXTSTSvmState{
			value: "shutoff",
		},
		SUSPENDED: ServerDetailsOSEXTSTSvmState{
			value: "suspended",
		},
		REBOOT: ServerDetailsOSEXTSTSvmState{
			value: "reboot",
		},
	}
}

func (c ServerDetailsOSEXTSTSvmState) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ServerDetailsOSEXTSTSvmState) UnmarshalJSON(b []byte) error {
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

type ServerDetailsOSEXTSTSpowerState struct {
	value int32
}

type ServerDetailsOSEXTSTSpowerStateEnum struct {
	E_0 ServerDetailsOSEXTSTSpowerState
	E_1 ServerDetailsOSEXTSTSpowerState
	E_4 ServerDetailsOSEXTSTSpowerState
}

func GetServerDetailsOSEXTSTSpowerStateEnum() ServerDetailsOSEXTSTSpowerStateEnum {
	return ServerDetailsOSEXTSTSpowerStateEnum{
		E_0: ServerDetailsOSEXTSTSpowerState{
			value: 0,
		}, E_1: ServerDetailsOSEXTSTSpowerState{
			value: 1,
		}, E_4: ServerDetailsOSEXTSTSpowerState{
			value: 4,
		},
	}
}

func (c ServerDetailsOSEXTSTSpowerState) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ServerDetailsOSEXTSTSpowerState) UnmarshalJSON(b []byte) error {
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

type ServerDetailsOSDCFdiskConfig struct {
	value string
}

type ServerDetailsOSDCFdiskConfigEnum struct {
	MANUAL ServerDetailsOSDCFdiskConfig
	AUTO   ServerDetailsOSDCFdiskConfig
}

func GetServerDetailsOSDCFdiskConfigEnum() ServerDetailsOSDCFdiskConfigEnum {
	return ServerDetailsOSDCFdiskConfigEnum{
		MANUAL: ServerDetailsOSDCFdiskConfig{
			value: "MANUAL",
		},
		AUTO: ServerDetailsOSDCFdiskConfig{
			value: "AUTO",
		},
	}
}

func (c ServerDetailsOSDCFdiskConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ServerDetailsOSDCFdiskConfig) UnmarshalJSON(b []byte) error {
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

type ServerDetailsHostStatus struct {
	value string
}

type ServerDetailsHostStatusEnum struct {
	UP          ServerDetailsHostStatus
	UNKNOWN     ServerDetailsHostStatus
	DOWN        ServerDetailsHostStatus
	MAINTENANCE ServerDetailsHostStatus
}

func GetServerDetailsHostStatusEnum() ServerDetailsHostStatusEnum {
	return ServerDetailsHostStatusEnum{
		UP: ServerDetailsHostStatus{
			value: "UP",
		},
		UNKNOWN: ServerDetailsHostStatus{
			value: "UNKNOWN",
		},
		DOWN: ServerDetailsHostStatus{
			value: "DOWN",
		},
		MAINTENANCE: ServerDetailsHostStatus{
			value: "MAINTENANCE",
		},
	}
}

func (c ServerDetailsHostStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ServerDetailsHostStatus) UnmarshalJSON(b []byte) error {
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
