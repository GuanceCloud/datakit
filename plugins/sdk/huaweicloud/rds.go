package huaweicloud

import (
	"encoding/json"
	"fmt"
)

const (
	rdsListUrl = "/v3/%s/instances"
)

type Datastore struct {
	Type    string `json:"type" required:"true"`
	Version string `json:"version" required:"true"`
}

type Ha struct {
	Mode            string `json:"mode" required:"true"`
	ReplicationMode string `json:"replication_mode,omitempty"`
}

type BackupStrategy struct {
	StartTime string `json:"start_time" required:"true"`
	KeepDays  int    `json:"keep_days,omitempty"`
}

type Volume struct {
	Type string `json:"type" required:"true"`
	Size int    `json:"size" required:"true"`
}

type ChargeInfo struct {
	ChargeMode  string `json:"charge_mode" required:"true"`
	PeriodType  string `json:"period_type,omitempty"`
	PeriodNum   int    `json:"period_num,omitempty"`
	IsAutoRenew string `json:"is_auto_renew,omitempty"`
	IsAutoPay   string `json:"is_auto_pay,omitempty"`
}

type ListRdsResponse struct {
	Instances  []RdsInstanceResponse `json:"instances"`
	TotalCount int                   `json:"total_count"`
}

type Nodes struct {
	Id               string `json:"id"`
	Name             string `json:"name"`
	Role             string `json:"role"`
	Status           string `json:"status"`
	AvailabilityZone string `json:"availability_zone"`
}
type RelatedInstance struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

type ChargeInfoMode struct {
	ChargeMode string `json:"charge_mode"`
}

type RdsInstanceResponse struct {
	Id                  string            `json:"id"`
	Name                string            `json:"name"`
	Status              string            `json:"status"`
	PrivateIps          []string          `json:"private_ips"`
	PublicIps           []string          `json:"public_ips"`
	Port                int               `json:"port"`
	Type                string            `json:"type"`
	Ha                  Ha                `json:"ha"`
	Region              string            `json:"region"`
	DataStore           Datastore         `json:"datastore"`
	Created             string            `json:"created"`
	Updated             string            `json:"updated"`
	DbUserName          string            `json:"db_user_name"`
	VpcId               string            `json:"vpc_id"`
	SubnetId            string            `json:"subnet_id"`
	SecurityGroupId     string            `json:"security_group_id"`
	FlavorRef           string            `json:"flavor_ref"`
	Volume              Volume            `json:"volume"`
	SwitchStrategy      string            `json:"switch_strategy"`
	BackupStrategy      BackupStrategy    `json:"backup_strategy"`
	MaintenanceWindow   string            `json:"maintenance_window"`
	Nodes               []Nodes           `json:"nodes"`
	RelatedInstance     []RelatedInstance `json:"related_instance"`
	DiskEncryptionId    string            `json:"disk_encryption_id"`
	EnterpriseProjectId string            `json:"enterprise_project_id"`
	TimeZone            string            `json:"time_zone"`
	ChargeInfo          ChargeInfoMode    `json:"charge_info"`
}

type ListRequestOpt struct {
	Id            string `q:"id"`
	Name          string `q:"name"`
	Type          string `q:"type"`
	DataStoreType string `q:"datastore_type"` //数据库类型，区分大小写。
	VpcId         string `q:"vpc_id"`
	SubnetId      string `q:"subnet_id"`
	Offset        int    `q:"offset"`
	Limit         int    `q:"limit"`
}

func (c *HWClient) RdsList(opts map[string]string) (res *ListRdsResponse, err error) {
	url := fmt.Sprintf(rdsListUrl, c.projectid)
	resp, err := c.Request("GET", url, opts, nil)
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
