package huaweiyunobject

import (
	"encoding/json"
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/sdk/huaweicloud"
)

const (
	mysqlSampleConfig = `
#[inputs.huaweiyunobject.mysql]
endpoint=""

# ## @param - custom tags - [list of mysql instanceid] - optional
#instanceids = ['']

# ## @param - custom tags - [list of excluded mysql instanceid] - optional
#exclude_instanceids = ['']

# ## @param - custom tags for mysql object - [list of key:value element] - optional
#[inputs.huaweiyunobject.mysql.tags]
# key1 = 'val1'
`
)

type Mysql struct {
	EndPoint string            `toml:"endpoint"`
	Tags     map[string]string `toml:"tags,omitempty"`
	//ProjectID          string            `toml:"project_id"`
	InstancesIDs       []string `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string `toml:"exclude_instanceids,omitempty"`
}

func (e *Mysql) run(ag *objectAgent) {

	if e.EndPoint == `` {
		e.EndPoint = fmt.Sprintf(`rds.%s.myhuaweicloud.com`, ag.RegionID)
	}
	cli := huaweicloud.NewHWClient(ag.AccessKeyID, ag.AccessKeySecret, e.EndPoint, ag.ProjectID, moduleLogger)

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		limit := 100
		offset := 0

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

			rdss, err := cli.RdsList(opts)
			if err != nil {
				moduleLogger.Errorf("%v", err)
				return
			}

			moduleLogger.Debugf("%+#v", rdss)
			e.handleResponse(rdss, ag)

			if rdss.TotalCount < offset+limit {
				break
			}

			offset = offset + limit
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (e *Mysql) handleResponse(resp *huaweicloud.ListRdsResponse, ag *objectAgent) {

	moduleLogger.Debugf("mysql TotalCount=%d", resp.TotalCount)

	var objs []map[string]interface{}

	for _, inst := range resp.Instances {

		if len(e.ExcludeInstanceIDs) > 0 {
			exclude := false
			for _, v := range e.ExcludeInstanceIDs {
				if v == inst.Id {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
		}

		obj := map[string]interface{}{
			`__name`: fmt.Sprintf(`%s(%s)`, inst.Name, inst.Id),
		}

		backupStrategy, err := json.Marshal(inst.BackupStrategy)
		if err != nil {
			moduleLogger.Errorf(`%s, ignore`, err.Error())
			return
		}
		obj[`backup_strategy`] = backupStrategy

		obj[`created`] = inst.Created

		obj[`version`] = inst.DataStore.Version
		obj[`db_user_name`] = inst.DbUserName

		obj[`disk_encryption_id`] = inst.DiskEncryptionId

		obj[`enterprise_project_id`] = inst.EnterpriseProjectId
		obj[`flavor_ref`] = inst.FlavorRef
		//obj[`ha_mode`] = inst.Ha.Mode
		//obj[`ha_replication_mode`] = inst.Ha.ReplicationMode
		obj[`id`] = inst.Id
		obj[`maintenance_window`] = inst.MaintenanceWindow

		nodes, err := json.Marshal(inst.Nodes)
		if err != nil {
			moduleLogger.Errorf(`%s, ignore`, err.Error())
			return
		}
		obj[`nodes`] = nodes

		obj[`port`] = inst.Port

		privateIps, err := json.Marshal(inst.PrivateIps)
		if err != nil {
			moduleLogger.Errorf(`%s, ignore`, err.Error())
			return
		}
		obj[`private_ips`] = privateIps

		publicIps, err := json.Marshal(inst.PublicIps)
		if err != nil {
			moduleLogger.Errorf(`%s, ignore`, err.Error())
			return
		}
		obj[`pupblic_ips`] = publicIps

		relatedInstance, err := json.Marshal(inst.RelatedInstance)
		if err != nil {
			moduleLogger.Errorf(`%s, ignore`, err.Error())
			return
		}
		obj[`related_instance`] = relatedInstance

		obj[`subnet_id`] = inst.SubnetId

		obj[`switch_strategy`] = inst.SwitchStrategy
		obj[`time_zone`] = inst.TimeZone
		obj[`updated`] = inst.Updated
		obj[`volume.size`] = inst.Volume.Size

		tags := map[string]interface{}{
			`__class`:             `huaweiyun_mysql`,
			`provider`:            `huaweiyun`,
			`datastore_type`:      inst.DataStore.Type,
			`charge_mode`:         inst.ChargeInfo.ChargeMode,
			`ha.mode`:             inst.Ha.Mode,
			`ha.replication_mode`: inst.Ha.ReplicationMode,
			`name`:                inst.Name,
			`region`:              inst.Region,
			`security_group_id`:   inst.SecurityGroupId,
			`status`:              inst.Status,
			`type`:                inst.Type,

			`volume.type`: inst.Volume.Type,
			`vpc_id`:      inst.VpcId,
		}

		//add mysql object custom tags
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
