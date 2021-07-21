package huaweicloud

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	ecsListUrl = "/v1/%s/cloudservers/detail"
)

type Address struct {
	Version string `json:"version"`
	Addr    string `json:"addr"`
	MacAddr string `json:"OS-EXT-IPS-MAC:mac_addr"`
	PortID  string `json:"OS-EXT-IPS:port_id"`
	Type    string `json:"OS-EXT-IPS:type"`
}

type SysTags struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SecurityGroups struct {
	Name string `json:"name"`
}

type VolumeAttached struct {
	ID                  string `json:"id"`
	DeleteOnTermination string `json:"delete_on_termination"`
	BootIndex           string `json:"bootIndex"`
	Device              string `json:"device"`
}

type OsSchedulerHints struct {
	Group []string `json:"group"`
}

// Metadata is only used for method that requests details on a single server, by ID.
// Because metadata struct must be a map.
type Metadata struct {
	ChargingMode      string `json:"charging_mode"`
	OrderID           string `json:"metering.order_id"`
	ProductID         string `json:"metering.product_id"`
	VpcID             string `json:"vpc_id"`
	EcmResStatus      string `json:"EcmResStatus"`
	ImageID           string `json:"metering.image_id"`
	Imagetype         string `json:"metering.imagetype"`
	Resourcespeccode  string `json:"metering.resourcespeccode"`
	ImageName         string `json:"image_name"`
	OsBit             string `json:"os_bit"`
	LockCheckEndpoint string `json:"lock_check_endpoint"`
	LockSource        string `json:"lock_source"`
	LockSourceID      string `json:"lock_source_id"`
	LockScene         string `json:"lock_scene"`
	VirtualEnvType    string `json:"virtual_env_type"`
}

type Flavor struct {
	Disk  string `json:"disk"`
	Vcpus string `json:"vcpus"`
	RAM   string `json:"ram"`
	ID    string `json:"id"`
	Name  string `json:"name"`
}

// Image defines a image struct in details of a server.
type Image struct {
	ID string `json:"id"`
}

type CloudServer struct {
	Status              string               `json:"status"`
	Updated             time.Time            `json:"updated"`
	HostID              string               `json:"hostId"`
	Addresses           map[string][]Address `json:"addresses"`
	ID                  string               `json:"id"`
	Name                string               `json:"name"`
	AccessIPv4          string               `json:"accessIPv4"`
	AccessIPv6          string               `json:"accessIPv6"`
	Created             time.Time            `json:"created"`
	Tags                []string             `json:"tags"`
	Description         string               `json:"description"`
	Locked              *bool                `json:"locked"`
	ConfigDrive         string               `json:"config_drive"`
	TenantID            string               `json:"tenant_id"`
	UserID              string               `json:"user_id"`
	HostStatus          string               `json:"host_status"`
	EnterpriseProjectID string               `json:"enterprise_project_id"`
	SysTags             []SysTags            `json:"sys_tags"`
	Flavor              Flavor               `json:"flavor"`
	Metadata            Metadata             `json:"metadata"`
	SecurityGroups      []SecurityGroups     `json:"security_groups"`
	KeyName             string               `json:"key_name"`
	Image               Image                `json:"image"`
	Progress            *int                 `json:"progress"`
	PowerState          *int                 `json:"OS-EXT-STS:power_state"`
	VMState             string               `json:"OS-EXT-STS:vm_state"`
	TaskState           string               `json:"OS-EXT-STS:task_state"`
	DiskConfig          string               `json:"OS-DCF:diskConfig"`
	AvailabilityZone    string               `json:"OS-EXT-AZ:availability_zone"`
	LaunchedAt          string               `json:"OS-SRV-USG:launched_at"`
	TerminatedAt        string               `json:"OS-SRV-USG:terminated_at"`
	RootDeviceName      string               `json:"OS-EXT-SRV-ATTR:root_device_name"`
	RamdiskID           string               `json:"OS-EXT-SRV-ATTR:ramdisk_id"`
	KernelID            string               `json:"OS-EXT-SRV-ATTR:kernel_id"`
	LaunchIndex         *int                 `json:"OS-EXT-SRV-ATTR:launch_index"`
	ReservationID       string               `json:"OS-EXT-SRV-ATTR:reservation_id"`
	Hostname            string               `json:"OS-EXT-SRV-ATTR:hostname"`
	UserData            string               `json:"OS-EXT-SRV-ATTR:user_data"`
	Host                string               `json:"OS-EXT-SRV-ATTR:host"`
	InstanceName        string               `json:"OS-EXT-SRV-ATTR:instance_name"`
	HypervisorHostname  string               `json:"OS-EXT-SRV-ATTR:hypervisor_hostname"`
	VolumeAttached      []VolumeAttached     `json:"os-extended-volumes:volumes_attached"`
	OsSchedulerHints    OsSchedulerHints     `json:"os:scheduler_hints"`
}

type ListEcsResponse struct {
	Servers []CloudServer `json:"servers"`
	Count   int           `json:"count"`
}

func (c *HWClient) EcsList(opt map[string]string) (res *ListEcsResponse, err error) {
	url := fmt.Sprintf(ecsListUrl, c.projectid)
	resp, err := c.Request("GET", url, opt, nil)
	if err != nil {
		c.logger.Errorf("%s", err)
		return nil, err
	}

	//c.logger.Debugf("resp %s", string(resp))
	err = json.Unmarshal(resp, &res)
	if err != nil {
		c.logger.Errorf("%s", err)
		return nil, err
	}

	return res, nil
}
