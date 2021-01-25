package v3

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rds/v3/model"
)

type RdsClient struct {
	HcClient *http_client.HcHttpClient
}

func NewRdsClient(hcClient *http_client.HcHttpClient) *RdsClient {
	return &RdsClient{HcClient: hcClient}
}

func RdsClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//绑定和解绑弹性公网IP。
func (c *RdsClient) AttachEip(request *model.AttachEipRequest) (*model.AttachEipResponse, error) {
	requestDef := GenReqDefForAttachEip()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AttachEipResponse), nil
	}
}

//批量添加删除标签。
func (c *RdsClient) BatchTagAction(request *model.BatchTagActionRequest) (*model.BatchTagActionResponse, error) {
	requestDef := GenReqDefForBatchTagAction()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchTagActionResponse), nil
	}
}

//更改主备实例的同步模式.
func (c *RdsClient) ChangeFailoverMode(request *model.ChangeFailoverModeRequest) (*model.ChangeFailoverModeResponse, error) {
	requestDef := GenReqDefForChangeFailoverMode()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ChangeFailoverModeResponse), nil
	}
}

//切换主备实例的倒换策略.
func (c *RdsClient) ChangeFailoverStrategy(request *model.ChangeFailoverStrategyRequest) (*model.ChangeFailoverStrategyResponse, error) {
	requestDef := GenReqDefForChangeFailoverStrategy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ChangeFailoverStrategyResponse), nil
	}
}

//设置可维护时间段
func (c *RdsClient) ChangeOpsWindow(request *model.ChangeOpsWindowRequest) (*model.ChangeOpsWindowResponse, error) {
	requestDef := GenReqDefForChangeOpsWindow()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ChangeOpsWindowResponse), nil
	}
}

//创建参数模板。
func (c *RdsClient) CreateConfiguration(request *model.CreateConfigurationRequest) (*model.CreateConfigurationResponse, error) {
	requestDef := GenReqDefForCreateConfiguration()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateConfigurationResponse), nil
	}
}

//创建数据库实例/恢复到新实例。
func (c *RdsClient) CreateInstance(request *model.CreateInstanceRequest) (*model.CreateInstanceResponse, error) {
	requestDef := GenReqDefForCreateInstance()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateInstanceResponse), nil
	}
}

//删除参数模板。
func (c *RdsClient) DeleteConfiguration(request *model.DeleteConfigurationRequest) (*model.DeleteConfigurationResponse, error) {
	requestDef := GenReqDefForDeleteConfiguration()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteConfigurationResponse), nil
	}
}

//删除实例。
func (c *RdsClient) DeleteInstance(request *model.DeleteInstanceRequest) (*model.DeleteInstanceResponse, error) {
	requestDef := GenReqDefForDeleteInstance()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteInstanceResponse), nil
	}
}

//删除手动备份。
func (c *RdsClient) DeleteManualBackup(request *model.DeleteManualBackupRequest) (*model.DeleteManualBackupResponse, error) {
	requestDef := GenReqDefForDeleteManualBackup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteManualBackupResponse), nil
	}
}

//创建手动备份。
func (c *RdsClient) DoManualBackup(request *model.DoManualBackupRequest) (*model.DoManualBackupResponse, error) {
	requestDef := GenReqDefForDoManualBackup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DoManualBackupResponse), nil
	}
}

//获取日志信息
func (c *RdsClient) DownloadSlowlog(request *model.DownloadSlowlogRequest) (*model.DownloadSlowlogResponse, error) {
	requestDef := GenReqDefForDownloadSlowlog()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DownloadSlowlogResponse), nil
	}
}

//应用参数模板。
func (c *RdsClient) EnableConfiguration(request *model.EnableConfigurationRequest) (*model.EnableConfigurationResponse, error) {
	requestDef := GenReqDefForEnableConfiguration()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.EnableConfigurationResponse), nil
	}
}

//获取审计日志列表。
func (c *RdsClient) ListAuditlogs(request *model.ListAuditlogsRequest) (*model.ListAuditlogsResponse, error) {
	requestDef := GenReqDefForListAuditlogs()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListAuditlogsResponse), nil
	}
}

//获取备份列表。
func (c *RdsClient) ListBackups(request *model.ListBackupsRequest) (*model.ListBackupsResponse, error) {
	requestDef := GenReqDefForListBackups()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListBackupsResponse), nil
	}
}

//查询SQLServer可用字符集
func (c *RdsClient) ListCollations(request *model.ListCollationsRequest) (*model.ListCollationsResponse, error) {
	requestDef := GenReqDefForListCollations()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListCollationsResponse), nil
	}
}

//获取参数模板列表，包括所有数据库的默认参数模板和用户创建的参数模板。
func (c *RdsClient) ListConfigurations(request *model.ListConfigurationsRequest) (*model.ListConfigurationsResponse, error) {
	requestDef := GenReqDefForListConfigurations()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListConfigurationsResponse), nil
	}
}

//查询数据库引擎的版本。
func (c *RdsClient) ListDatastores(request *model.ListDatastoresRequest) (*model.ListDatastoresResponse, error) {
	requestDef := GenReqDefForListDatastores()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListDatastoresResponse), nil
	}
}

//查询数据库错误日志。
func (c *RdsClient) ListErrorLogs(request *model.ListErrorLogsRequest) (*model.ListErrorLogsResponse, error) {
	requestDef := GenReqDefForListErrorLogs()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListErrorLogsResponse), nil
	}
}

//查询数据库规格。
func (c *RdsClient) ListFlavors(request *model.ListFlavorsRequest) (*model.ListFlavorsResponse, error) {
	requestDef := GenReqDefForListFlavors()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListFlavorsResponse), nil
	}
}

//查询数据库实例列表。
func (c *RdsClient) ListInstances(request *model.ListInstancesRequest) (*model.ListInstancesResponse, error) {
	requestDef := GenReqDefForListInstances()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListInstancesResponse), nil
	}
}

//获取任务信息。
func (c *RdsClient) ListJobInfo(request *model.ListJobInfoRequest) (*model.ListJobInfoResponse, error) {
	requestDef := GenReqDefForListJobInfo()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListJobInfoResponse), nil
	}
}

//获取所有任务详细信息。
func (c *RdsClient) ListJobInfoDetail(request *model.ListJobInfoDetailRequest) (*model.ListJobInfoDetailResponse, error) {
	requestDef := GenReqDefForListJobInfoDetail()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListJobInfoDetailResponse), nil
	}
}

//获取跨区域备份列表。
func (c *RdsClient) ListOffSiteBackups(request *model.ListOffSiteBackupsRequest) (*model.ListOffSiteBackupsResponse, error) {
	requestDef := GenReqDefForListOffSiteBackups()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListOffSiteBackupsResponse), nil
	}
}

//查询跨区域备份实例列表。
func (c *RdsClient) ListOffSiteInstances(request *model.ListOffSiteInstancesRequest) (*model.ListOffSiteInstancesResponse, error) {
	requestDef := GenReqDefForListOffSiteInstances()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListOffSiteInstancesResponse), nil
	}
}

//查询跨区域可恢复时间段。 如果您备份策略中的保存天数设置较长，建议您传入查询日期“date”。
func (c *RdsClient) ListOffSiteRestoreTimes(request *model.ListOffSiteRestoreTimesRequest) (*model.ListOffSiteRestoreTimesResponse, error) {
	requestDef := GenReqDefForListOffSiteRestoreTimes()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListOffSiteRestoreTimesResponse), nil
	}
}

//查询项目标签。
func (c *RdsClient) ListProjectTags(request *model.ListProjectTagsRequest) (*model.ListProjectTagsResponse, error) {
	requestDef := GenReqDefForListProjectTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListProjectTagsResponse), nil
	}
}

//查询可恢复时间段。 如果您备份策略中的保存天数设置较长，建议您传入查询日期“date”。
func (c *RdsClient) ListRestoreTimes(request *model.ListRestoreTimesRequest) (*model.ListRestoreTimesResponse, error) {
	requestDef := GenReqDefForListRestoreTimes()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListRestoreTimesResponse), nil
	}
}

//查询数据库慢日志。
func (c *RdsClient) ListSlowLogs(request *model.ListSlowLogsRequest) (*model.ListSlowLogsResponse, error) {
	requestDef := GenReqDefForListSlowLogs()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListSlowLogsResponse), nil
	}
}

//获取慢日志统计信息
func (c *RdsClient) ListSlowlogStatistics(request *model.ListSlowlogStatisticsRequest) (*model.ListSlowlogStatisticsResponse, error) {
	requestDef := GenReqDefForListSlowlogStatistics()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListSlowlogStatisticsResponse), nil
	}
}

//查询数据库磁盘类型。
func (c *RdsClient) ListStorageTypes(request *model.ListStorageTypesRequest) (*model.ListStorageTypesResponse, error) {
	requestDef := GenReqDefForListStorageTypes()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListStorageTypesResponse), nil
	}
}

//迁移主备实例的备机
func (c *RdsClient) MigrateFollower(request *model.MigrateFollowerRequest) (*model.MigrateFollowerResponse, error) {
	requestDef := GenReqDefForMigrateFollower()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.MigrateFollowerResponse), nil
	}
}

//修改参数模板参数。
func (c *RdsClient) ModifyConfiguration(request *model.ModifyConfigurationRequest) (*model.ModifyConfigurationResponse, error) {
	requestDef := GenReqDefForModifyConfiguration()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ModifyConfigurationResponse), nil
	}
}

//修改指定实例的参数。
func (c *RdsClient) ModifyInstanceConfiguration(request *model.ModifyInstanceConfigurationRequest) (*model.ModifyInstanceConfigurationResponse, error) {
	requestDef := GenReqDefForModifyInstanceConfiguration()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ModifyInstanceConfigurationResponse), nil
	}
}

//重置数据库密码.
func (c *RdsClient) ResetPwd(request *model.ResetPwdRequest) (*model.ResetPwdResponse, error) {
	requestDef := GenReqDefForResetPwd()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ResetPwdResponse), nil
	}
}

//表级时间点恢复。
func (c *RdsClient) RestoreTables(request *model.RestoreTablesRequest) (*model.RestoreTablesResponse, error) {
	requestDef := GenReqDefForRestoreTables()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.RestoreTablesResponse), nil
	}
}

//恢复到已有实例。
func (c *RdsClient) RestoreToExistingInstance(request *model.RestoreToExistingInstanceRequest) (*model.RestoreToExistingInstanceResponse, error) {
	requestDef := GenReqDefForRestoreToExistingInstance()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.RestoreToExistingInstanceResponse), nil
	}
}

//设置审计日志策略。
func (c *RdsClient) SetAuditlogPolicy(request *model.SetAuditlogPolicyRequest) (*model.SetAuditlogPolicyResponse, error) {
	requestDef := GenReqDefForSetAuditlogPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.SetAuditlogPolicyResponse), nil
	}
}

//设置自动备份策略。
func (c *RdsClient) SetBackupPolicy(request *model.SetBackupPolicyRequest) (*model.SetBackupPolicyResponse, error) {
	requestDef := GenReqDefForSetBackupPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.SetBackupPolicyResponse), nil
	}
}

//设置跨区域备份策略。
func (c *RdsClient) SetOffSiteBackupPolicy(request *model.SetOffSiteBackupPolicyRequest) (*model.SetOffSiteBackupPolicyResponse, error) {
	requestDef := GenReqDefForSetOffSiteBackupPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.SetOffSiteBackupPolicyResponse), nil
	}
}

//修改安全组
func (c *RdsClient) SetSecurityGroup(request *model.SetSecurityGroupRequest) (*model.SetSecurityGroupResponse, error) {
	requestDef := GenReqDefForSetSecurityGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.SetSecurityGroupResponse), nil
	}
}

//生成审计日志下载链接。
func (c *RdsClient) ShowAuditlogDownloadLink(request *model.ShowAuditlogDownloadLinkRequest) (*model.ShowAuditlogDownloadLinkResponse, error) {
	requestDef := GenReqDefForShowAuditlogDownloadLink()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowAuditlogDownloadLinkResponse), nil
	}
}

//查询审计日志策略。
func (c *RdsClient) ShowAuditlogPolicy(request *model.ShowAuditlogPolicyRequest) (*model.ShowAuditlogPolicyResponse, error) {
	requestDef := GenReqDefForShowAuditlogPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowAuditlogPolicyResponse), nil
	}
}

//获取备份下载链接。
func (c *RdsClient) ShowBackupDownloadLink(request *model.ShowBackupDownloadLinkRequest) (*model.ShowBackupDownloadLinkResponse, error) {
	requestDef := GenReqDefForShowBackupDownloadLink()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowBackupDownloadLinkResponse), nil
	}
}

//查询自动备份策略。
func (c *RdsClient) ShowBackupPolicy(request *model.ShowBackupPolicyRequest) (*model.ShowBackupPolicyResponse, error) {
	requestDef := GenReqDefForShowBackupPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowBackupPolicyResponse), nil
	}
}

//获取指定参数模板的参数。
func (c *RdsClient) ShowConfiguration(request *model.ShowConfigurationRequest) (*model.ShowConfigurationResponse, error) {
	requestDef := GenReqDefForShowConfiguration()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowConfigurationResponse), nil
	}
}

//获取指定实例的参数模板。
func (c *RdsClient) ShowInstanceConfiguration(request *model.ShowInstanceConfigurationRequest) (*model.ShowInstanceConfigurationResponse, error) {
	requestDef := GenReqDefForShowInstanceConfiguration()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowInstanceConfigurationResponse), nil
	}
}

//查询跨区域备份策略。
func (c *RdsClient) ShowOffSiteBackupPolicy(request *model.ShowOffSiteBackupPolicyRequest) (*model.ShowOffSiteBackupPolicyResponse, error) {
	requestDef := GenReqDefForShowOffSiteBackupPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowOffSiteBackupPolicyResponse), nil
	}
}

//手动倒换主备.
func (c *RdsClient) StartFailover(request *model.StartFailoverRequest) (*model.StartFailoverResponse, error) {
	requestDef := GenReqDefForStartFailover()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.StartFailoverResponse), nil
	}
}

//变更实例规格/扩容实例磁盘/重启实例/单机转主备。
func (c *RdsClient) StartInstanceAction(request *model.StartInstanceActionRequest) (*model.StartInstanceActionResponse, error) {
	requestDef := GenReqDefForStartInstanceAction()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.StartInstanceActionResponse), nil
	}
}

//SSL开关
func (c *RdsClient) SwitchSsl(request *model.SwitchSslRequest) (*model.SwitchSslResponse, error) {
	requestDef := GenReqDefForSwitchSsl()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.SwitchSslResponse), nil
	}
}

//修改内网地址
func (c *RdsClient) UpdateDataIp(request *model.UpdateDataIpRequest) (*model.UpdateDataIpResponse, error) {
	requestDef := GenReqDefForUpdateDataIp()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateDataIpResponse), nil
	}
}

//修改数据库端口
func (c *RdsClient) UpdatePort(request *model.UpdatePortRequest) (*model.UpdatePortResponse, error) {
	requestDef := GenReqDefForUpdatePort()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdatePortResponse), nil
	}
}

//授权数据库帐号。
func (c *RdsClient) AllowDbUserPrivilege(request *model.AllowDbUserPrivilegeRequest) (*model.AllowDbUserPrivilegeResponse, error) {
	requestDef := GenReqDefForAllowDbUserPrivilege()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AllowDbUserPrivilegeResponse), nil
	}
}

//创建数据库。
func (c *RdsClient) CreateDatabase(request *model.CreateDatabaseRequest) (*model.CreateDatabaseResponse, error) {
	requestDef := GenReqDefForCreateDatabase()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateDatabaseResponse), nil
	}
}

//创建数据库用户。
func (c *RdsClient) CreateDbUser(request *model.CreateDbUserRequest) (*model.CreateDbUserResponse, error) {
	requestDef := GenReqDefForCreateDbUser()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateDbUserResponse), nil
	}
}

//删除数据库。
func (c *RdsClient) DeleteDatabase(request *model.DeleteDatabaseRequest) (*model.DeleteDatabaseResponse, error) {
	requestDef := GenReqDefForDeleteDatabase()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteDatabaseResponse), nil
	}
}

//删除数据库用户。
func (c *RdsClient) DeleteDbUser(request *model.DeleteDbUserRequest) (*model.DeleteDbUserResponse, error) {
	requestDef := GenReqDefForDeleteDbUser()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteDbUserResponse), nil
	}
}

//查询指定用户的已授权数据库。
func (c *RdsClient) ListAuthorizedDatabases(request *model.ListAuthorizedDatabasesRequest) (*model.ListAuthorizedDatabasesResponse, error) {
	requestDef := GenReqDefForListAuthorizedDatabases()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListAuthorizedDatabasesResponse), nil
	}
}

//查询指定数据库的已授权用户。
func (c *RdsClient) ListAuthorizedDbUsers(request *model.ListAuthorizedDbUsersRequest) (*model.ListAuthorizedDbUsersResponse, error) {
	requestDef := GenReqDefForListAuthorizedDbUsers()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListAuthorizedDbUsersResponse), nil
	}
}

//查询数据库列表。
func (c *RdsClient) ListDatabases(request *model.ListDatabasesRequest) (*model.ListDatabasesResponse, error) {
	requestDef := GenReqDefForListDatabases()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListDatabasesResponse), nil
	}
}

//查询数据库用户列表。
func (c *RdsClient) ListDbUsers(request *model.ListDbUsersRequest) (*model.ListDbUsersResponse, error) {
	requestDef := GenReqDefForListDbUsers()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListDbUsersResponse), nil
	}
}

//解除数据库帐号权限。
func (c *RdsClient) Revoke(request *model.RevokeRequest) (*model.RevokeResponse, error) {
	requestDef := GenReqDefForRevoke()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.RevokeResponse), nil
	}
}

//设置数据库账号密码
func (c *RdsClient) SetDbUserPwd(request *model.SetDbUserPwdRequest) (*model.SetDbUserPwdResponse, error) {
	requestDef := GenReqDefForSetDbUserPwd()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.SetDbUserPwdResponse), nil
	}
}
