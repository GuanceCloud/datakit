package v1

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cbr/v1/model"
)

type CbrClient struct {
	HcClient *http_client.HcHttpClient
}

func NewCbrClient(hcClient *http_client.HcHttpClient) *CbrClient {
	return &CbrClient{HcClient: hcClient}
}

func CbrClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//添加备份可共享的成员，只有云服务器备份可以添加备份共享成员，且仅支持在同一区域的不同用户间共享。
func (c *CbrClient) AddMember(request *model.AddMemberRequest) (*model.AddMemberResponse, error) {
	requestDef := GenReqDefForAddMember()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AddMemberResponse), nil
	}
}

//存储库添加资源
func (c *CbrClient) AddVaultResource(request *model.AddVaultResourceRequest) (*model.AddVaultResourceResponse, error) {
	requestDef := GenReqDefForAddVaultResource()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AddVaultResourceResponse), nil
	}
}

//存储库设置策略
func (c *CbrClient) AssociateVaultPolicy(request *model.AssociateVaultPolicyRequest) (*model.AssociateVaultPolicyResponse, error) {
	requestDef := GenReqDefForAssociateVaultPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AssociateVaultPolicyResponse), nil
	}
}

//为指定实例批量添加或删除标签 标签管理服务需要使用该接口批量管理实例的标签。 一个资源上最多有10个标签。 此接口为幂等接口：     创建时如果请求体中存在重复key则报错。     创建时，不允许重复key，如果数据库存在就覆盖。     删除时，允许重复key。     删除时，如果删除的标签不存在，默认处理成功,删除时不对标签字符集范围做校验。key长度127个字符，value为255个字符。删除时tags结构体不能缺失，key不能为空，或者空字符串。
func (c *CbrClient) BatchCreateAndDeleteVaultTags(request *model.BatchCreateAndDeleteVaultTagsRequest) (*model.BatchCreateAndDeleteVaultTagsResponse, error) {
	requestDef := GenReqDefForBatchCreateAndDeleteVaultTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchCreateAndDeleteVaultTagsResponse), nil
	}
}

//跨区域复制备份。
func (c *CbrClient) CopyBackup(request *model.CopyBackupRequest) (*model.CopyBackupResponse, error) {
	requestDef := GenReqDefForCopyBackup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CopyBackupResponse), nil
	}
}

//执行复制
func (c *CbrClient) CopyCheckpoint(request *model.CopyCheckpointRequest) (*model.CopyCheckpointResponse, error) {
	requestDef := GenReqDefForCopyCheckpoint()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CopyCheckpointResponse), nil
	}
}

//对存储库执行备份，生成备份还原点
func (c *CbrClient) CreateCheckpoint(request *model.CreateCheckpointRequest) (*model.CreateCheckpointResponse, error) {
	requestDef := GenReqDefForCreateCheckpoint()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateCheckpointResponse), nil
	}
}

//[创建策略，策略分为备份策略和复制策略。](tag:hws,hws_hk) [创建备份策略。](tag:dt,ocb,tlf,sbc,fcs_vm,ctc)
func (c *CbrClient) CreatePolicy(request *model.CreatePolicyRequest) (*model.CreatePolicyResponse, error) {
	requestDef := GenReqDefForCreatePolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreatePolicyResponse), nil
	}
}

//创建存储库
func (c *CbrClient) CreateVault(request *model.CreateVaultRequest) (*model.CreateVaultResponse, error) {
	requestDef := GenReqDefForCreateVault()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateVaultResponse), nil
	}
}

//一个资源上最多有10个标签。 此接口为幂等接口：创建时，如果创建的标签已经存在（key相同），则覆盖。
func (c *CbrClient) CreateVaultTags(request *model.CreateVaultTagsRequest) (*model.CreateVaultTagsResponse, error) {
	requestDef := GenReqDefForCreateVaultTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateVaultTagsResponse), nil
	}
}

//删除单个备份。
func (c *CbrClient) DeleteBackup(request *model.DeleteBackupRequest) (*model.DeleteBackupResponse, error) {
	requestDef := GenReqDefForDeleteBackup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteBackupResponse), nil
	}
}

//删除指定的备份共享成员
func (c *CbrClient) DeleteMember(request *model.DeleteMemberRequest) (*model.DeleteMemberResponse, error) {
	requestDef := GenReqDefForDeleteMember()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteMemberResponse), nil
	}
}

//删除策略
func (c *CbrClient) DeletePolicy(request *model.DeletePolicyRequest) (*model.DeletePolicyResponse, error) {
	requestDef := GenReqDefForDeletePolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeletePolicyResponse), nil
	}
}

//删除存储库。若删除储存库，将一并删除存储库中的所有备份。
func (c *CbrClient) DeleteVault(request *model.DeleteVaultRequest) (*model.DeleteVaultResponse, error) {
	requestDef := GenReqDefForDeleteVault()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteVaultResponse), nil
	}
}

//幂等接口：删除时，如果删除的标签不存在，返回404。Key不能为空或者空字符串。
func (c *CbrClient) DeleteVaultTag(request *model.DeleteVaultTagRequest) (*model.DeleteVaultTagResponse, error) {
	requestDef := GenReqDefForDeleteVaultTag()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteVaultTagResponse), nil
	}
}

//存储库解除策略
func (c *CbrClient) DisassociateVaultPolicy(request *model.DisassociateVaultPolicyRequest) (*model.DisassociateVaultPolicyResponse, error) {
	requestDef := GenReqDefForDisassociateVaultPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DisassociateVaultPolicyResponse), nil
	}
}

//同步备份副本
func (c *CbrClient) ImportBackup(request *model.ImportBackupRequest) (*model.ImportBackupResponse, error) {
	requestDef := GenReqDefForImportBackup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ImportBackupResponse), nil
	}
}

//查询所有副本
func (c *CbrClient) ListBackups(request *model.ListBackupsRequest) (*model.ListBackupsResponse, error) {
	requestDef := GenReqDefForListBackups()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListBackupsResponse), nil
	}
}

//查询任务列表
func (c *CbrClient) ListOpLogs(request *model.ListOpLogsRequest) (*model.ListOpLogsResponse, error) {
	requestDef := GenReqDefForListOpLogs()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListOpLogsResponse), nil
	}
}

//查询策略列表
func (c *CbrClient) ListPolicies(request *model.ListPoliciesRequest) (*model.ListPoliciesResponse, error) {
	requestDef := GenReqDefForListPolicies()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListPoliciesResponse), nil
	}
}

//查询可保护性资源列表
func (c *CbrClient) ListProtectable(request *model.ListProtectableRequest) (*model.ListProtectableResponse, error) {
	requestDef := GenReqDefForListProtectable()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListProtectableResponse), nil
	}
}

//查询存储库列表
func (c *CbrClient) ListVault(request *model.ListVaultRequest) (*model.ListVaultResponse, error) {
	requestDef := GenReqDefForListVault()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListVaultResponse), nil
	}
}

//支持资源迁移到另一个存储库，不删除备份。
func (c *CbrClient) MigrateVaultResource(request *model.MigrateVaultResourceRequest) (*model.MigrateVaultResourceResponse, error) {
	requestDef := GenReqDefForMigrateVaultResource()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.MigrateVaultResourceResponse), nil
	}
}

//移除存储库中的资源，若移除资源，将一并删除该资源在保管库中的备份
func (c *CbrClient) RemoveVaultResource(request *model.RemoveVaultResourceRequest) (*model.RemoveVaultResourceResponse, error) {
	requestDef := GenReqDefForRemoveVaultResource()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.RemoveVaultResourceResponse), nil
	}
}

//恢复备份数据
func (c *CbrClient) RestoreBackup(request *model.RestoreBackupRequest) (*model.RestoreBackupResponse, error) {
	requestDef := GenReqDefForRestoreBackup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.RestoreBackupResponse), nil
	}
}

//根据指定id查询单个副本。
func (c *CbrClient) ShowBackup(request *model.ShowBackupRequest) (*model.ShowBackupResponse, error) {
	requestDef := GenReqDefForShowBackup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowBackupResponse), nil
	}
}

//根据还原点ID查询指定还原点
func (c *CbrClient) ShowCheckpoint(request *model.ShowCheckpointRequest) (*model.ShowCheckpointResponse, error) {
	requestDef := GenReqDefForShowCheckpoint()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowCheckpointResponse), nil
	}
}

//获取备份成员的详情
func (c *CbrClient) ShowMemberDetail(request *model.ShowMemberDetailRequest) (*model.ShowMemberDetailResponse, error) {
	requestDef := GenReqDefForShowMemberDetail()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowMemberDetailResponse), nil
	}
}

//获取备份共享成员的列表信息
func (c *CbrClient) ShowMembersDetail(request *model.ShowMembersDetailRequest) (*model.ShowMembersDetailResponse, error) {
	requestDef := GenReqDefForShowMembersDetail()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowMembersDetailResponse), nil
	}
}

//根据指定任务ID查询任务
func (c *CbrClient) ShowOpLog(request *model.ShowOpLogRequest) (*model.ShowOpLogResponse, error) {
	requestDef := GenReqDefForShowOpLog()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowOpLogResponse), nil
	}
}

//查询单个策略
func (c *CbrClient) ShowPolicy(request *model.ShowPolicyRequest) (*model.ShowPolicyResponse, error) {
	requestDef := GenReqDefForShowPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowPolicyResponse), nil
	}
}

//根据ID查询可保护性资源
func (c *CbrClient) ShowProtectable(request *model.ShowProtectableRequest) (*model.ShowProtectableResponse, error) {
	requestDef := GenReqDefForShowProtectable()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowProtectableResponse), nil
	}
}

//查询本区域的复制能力
func (c *CbrClient) ShowReplicationCapabilities(request *model.ShowReplicationCapabilitiesRequest) (*model.ShowReplicationCapabilitiesResponse, error) {
	requestDef := GenReqDefForShowReplicationCapabilities()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowReplicationCapabilitiesResponse), nil
	}
}

//根据ID查询指定存储库
func (c *CbrClient) ShowVault(request *model.ShowVaultRequest) (*model.ShowVaultResponse, error) {
	requestDef := GenReqDefForShowVault()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowVaultResponse), nil
	}
}

//查询租户在指定Region和实例类型的所有标签集合 标签管理服务需要能够列出当前租户全部已使用的标签集合，为各服务Console打标签和过滤实例时提供标签联想功能
func (c *CbrClient) ShowVaultProjectTag(request *model.ShowVaultProjectTagRequest) (*model.ShowVaultProjectTagResponse, error) {
	requestDef := GenReqDefForShowVaultProjectTag()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowVaultProjectTagResponse), nil
	}
}

//使用标签过滤实例 标签管理服务需要提供按标签过滤各服务实例并汇总显示在列表中，需要各服务提供查询能力
func (c *CbrClient) ShowVaultResourceInstances(request *model.ShowVaultResourceInstancesRequest) (*model.ShowVaultResourceInstancesResponse, error) {
	requestDef := GenReqDefForShowVaultResourceInstances()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowVaultResourceInstancesResponse), nil
	}
}

//查询指定实例的标签信息 标签管理服务需要使用该接口查询指定实例的全部标签数据
func (c *CbrClient) ShowVaultTag(request *model.ShowVaultTagRequest) (*model.ShowVaultTagResponse, error) {
	requestDef := GenReqDefForShowVaultTag()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowVaultTagResponse), nil
	}
}

//更新备份共享成员的状态，需要接收方执行此API。
func (c *CbrClient) UpdateMemberStatus(request *model.UpdateMemberStatusRequest) (*model.UpdateMemberStatusResponse, error) {
	requestDef := GenReqDefForUpdateMemberStatus()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateMemberStatusResponse), nil
	}
}

//修改策略
func (c *CbrClient) UpdatePolicy(request *model.UpdatePolicyRequest) (*model.UpdatePolicyResponse, error) {
	requestDef := GenReqDefForUpdatePolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdatePolicyResponse), nil
	}
}

//根据存储库ID修改存储库
func (c *CbrClient) UpdateVault(request *model.UpdateVaultRequest) (*model.UpdateVaultResponse, error) {
	requestDef := GenReqDefForUpdateVault()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateVaultResponse), nil
	}
}
