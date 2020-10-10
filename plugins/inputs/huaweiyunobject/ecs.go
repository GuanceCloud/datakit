package huaweiyunobject

import (
	"encoding/json"
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/sdk/huaweicloud"
)

const (
	ecsSampleConfig = `
#[inputs.huaweiyunobject.ecs]

## 地区和终端节点 https://developer.huaweicloud.com/endpoint?ECS  required
endpoint=""

# ## @param - custom tags - [list of ecs instanceid] - optional
#instanceids = ['']

# ## @param - custom tags - [list of excluded ecs instanceid] - optional
#exclude_instanceids = ['']

# ## @param - custom tags for ecs object - [list of key:value element] - optional
#[inputs.huaweiyunobject.ecs.tags]
# key1 = 'val1'
`
)

type Ecs struct {
	Tags map[string]string `toml:"tags,omitempty"`
	//	ProjectID          string            `toml:"project_id"`
	EndPoint           string   `toml:"endpoint"`
	InstancesIDs       []string `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string `toml:"exclude_instanceids,omitempty"`
}

func (e *Ecs) run(ag *objectAgent) {

	if e.EndPoint == `` {
		e.EndPoint = fmt.Sprintf(`ecs.%s.myhuaweicloud.com`, ag.RegionID)
	}

	cli := huaweicloud.NewHWClient(ag.AccessKeyID, ag.AccessKeySecret, e.EndPoint, ag.ProjectID, moduleLogger)

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		limit := 100
		offset := 1

		for {

			select {
			case <-ag.ctx.Done():
				return
			default:
			}
			opts := map[string]string{
				"limit":  fmt.Sprintf("%d", limit),
				"offset": fmt.Sprintf("%d", offset),
			}

			ecss, err := cli.EcsList(opts)
			if err != nil {
				moduleLogger.Errorf("%v", err)
				return
			}
			e.handleResponse(ecss, ag)

			if ecss.Count < offset*limit {
				break
			}

			offset++
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (e *Ecs) handleResponse(resp *huaweicloud.ListEcsResponse, ag *objectAgent) {

	moduleLogger.Debugf("ECS TotalCount=%d", resp.Count)

	var objs []map[string]interface{}

	for _, s := range resp.Servers {

		if len(e.ExcludeInstanceIDs) > 0 {
			exclude := false
			for _, v := range e.ExcludeInstanceIDs {
				if v == s.ID {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
		}

		obj := map[string]interface{}{
			`__name`: fmt.Sprintf(`%s(%s)`, s.InstanceName, s.ID),
		}

		obj[`accessIPv4`] = s.AccessIPv4
		obj[`accessIPv6`] = s.AccessIPv6
		obj[`creation_time`] = s.Addresses
		obj[`OS-EXT-AZ:availability_zone`] = s.AvailabilityZone
		obj[`config_drive`] = s.ConfigDrive
		obj[`created`] = fmt.Sprintf("%v", s.Created)
		obj[`description`] = s.Description
		obj[`OS-DCF:diskConfig`] = s.DiskConfig

		flavor, err := json.Marshal(s.Flavor)
		if err != nil {
			moduleLogger.Errorf(`%s, ignore`, err.Error())
			return
		}
		obj[`flavor`] = flavor

		obj[`hostId`] = s.HostID
		obj[`host_status`] = s.HostStatus
		obj[`OS-EXT-SRV-ATTR:hypervisor_hostname`] = s.HypervisorHostname
		obj[`id`] = s.ID

		image, err := json.Marshal(s.Image)
		if err != nil {
			moduleLogger.Errorf(`%s, ignore`, err.Error())
			return
		}
		obj[`image`] = image

		obj[`OS-EXT-SRV-ATTR:instance_name`] = s.InstanceName
		obj[`OS-EXT-SRV-ATTR:kernel_id`] = s.KernelID
		obj[`key_name`] = s.KeyName
		obj[`OS-EXT-SRV-ATTR:launch_index`] = s.LaunchIndex
		obj[`OS-SRV-USG:launched_at`] = s.LaunchedAt
		obj[`name`] = s.Name

		metadata, err := json.Marshal(s.Metadata)
		if err != nil {
			moduleLogger.Errorf(`%s, ignore`, err.Error())
			return
		}
		obj[`metadata`] = metadata

		osSchedulerHints, err := json.Marshal(s.OsSchedulerHints)
		if err != nil {
			moduleLogger.Errorf(`%s, ignore`, err.Error())
			return
		}
		obj[`os:scheduler_hints`] = osSchedulerHints

		obj[`OS-EXT-STS:power_state`] = s.PowerState
		obj[`progress`] = s.Progress
		obj[`OS-EXT-SRV-ATTR:ramdisk_id`] = s.RamdiskID
		obj[`OS-EXT-SRV-ATTR:reservation_id`] = s.ReservationID
		obj[`OS-EXT-SRV-ATTR:root_device_name`] = s.RootDeviceName

		securityGroups, err := json.Marshal(s.SecurityGroups)
		if err != nil {
			moduleLogger.Errorf(`%s, ignore`, err.Error())
			return
		}
		obj[`security_groups`] = securityGroups

		sTags, err := json.Marshal(s.Tags)
		if err != nil {
			moduleLogger.Errorf(`%s, ignore`, err.Error())
			return
		}
		obj[`tags`] = sTags

		obj[`OS-EXT-STS:task_state`] = s.TaskState
		obj[`tenant_id`] = s.TenantID
		obj[`OS-SRV-USG:terminated_at`] = s.TerminatedAt
		obj[`updated`] = s.Updated
		obj[`OS-EXT-SRV-ATTR:user_data`] = s.UserData
		obj[`user_id`] = s.UserID

		volumeAttached, err := json.Marshal(s.VolumeAttached)
		if err != nil {
			moduleLogger.Errorf(`%s, ignore`, err.Error())
			return
		}
		obj[`os-extended-volumes:volumes_attached`] = volumeAttached

		tags := map[string]interface{}{
			`__class`:                  `huaweiyun_ecs`,
			`provider`:                 `huaweiyun`,
			`enterprise_project_id`:    s.EnterpriseProjectID,
			`host`:                     s.Host,
			`OS-EXT-SRV-ATTR:hostname`: s.Hostname,
			`locked`:                   s.Locked,
			`status`:                   s.Status,
			`OS-EXT-STS:vm_state`:      s.VMState,
		}

		//tags on ecs instance
		for _, t := range s.SysTags {
			if _, have := tags[t.Key]; !have {
				tags[t.Key] = t.Value
			} else {
				tags[`custom_`+t.Key] = t.Value
			}
		}

		//add ecs object custom tags
		for k, v := range e.Tags {
			tags[k] = v
		}

		//add global tags
		for k, v := range ag.Tags {
			if _, have := tags[k]; !have {
				tags[k] = v
			}
		}

		obj["__tags"] = tags

		objs = append(objs, obj)
	}

	if len(objs) <= 0 {
		return
	}

	data, err := json.Marshal(&objs)
	if err == nil {
		io.NamedFeed(data, io.Object, inputName)
	} else {
		moduleLogger.Errorf("%s", err)
	}
}
