package v1

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bms/v1/model"
)

type BmsClient struct {
	HcClient *http_client.HcHttpClient
}

func NewBmsClient(hcClient *http_client.HcHttpClient) *BmsClient {
	return &BmsClient{HcClient: hcClient}
}

func BmsClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//裸金属服务器创建成功后，如果发现磁盘不够用或者当前磁盘不满足要求，可以将已有云硬盘挂载给裸金属服务器，作为数据盘使用
func (c *BmsClient) AttachBaremetalServerVolume(request *model.AttachBaremetalServerVolumeRequest) (*model.AttachBaremetalServerVolumeResponse, error) {
	requestDef := GenReqDefForAttachBaremetalServerVolume()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AttachBaremetalServerVolumeResponse), nil
	}
}

//根据给定的裸金属服务器ID列表，批量重启裸金属服务器
func (c *BmsClient) BatchRebootBaremetalServers(request *model.BatchRebootBaremetalServersRequest) (*model.BatchRebootBaremetalServersResponse, error) {
	requestDef := GenReqDefForBatchRebootBaremetalServers()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchRebootBaremetalServersResponse), nil
	}
}

//根据给定的裸金属服务器ID列表，批量启动裸金属服务器
func (c *BmsClient) BatchStartBaremetalServers(request *model.BatchStartBaremetalServersRequest) (*model.BatchStartBaremetalServersResponse, error) {
	requestDef := GenReqDefForBatchStartBaremetalServers()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchStartBaremetalServersResponse), nil
	}
}

//根据给定的裸金属服务器ID列表，批量关闭裸金属服务器
func (c *BmsClient) BatchStopBaremetalServers(request *model.BatchStopBaremetalServersRequest) (*model.BatchStopBaremetalServersResponse, error) {
	requestDef := GenReqDefForBatchStopBaremetalServers()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchStopBaremetalServersResponse), nil
	}
}

//修改裸金属服务器名称
func (c *BmsClient) ChangeBaremetalServerName(request *model.ChangeBaremetalServerNameRequest) (*model.ChangeBaremetalServerNameResponse, error) {
	requestDef := GenReqDefForChangeBaremetalServerName()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ChangeBaremetalServerNameResponse), nil
	}
}

//创建一台或多台裸金属服务器,裸金属服务器的登录鉴权方式包括两种：密钥对、密码。为安全起见，推荐使用密钥对方式
func (c *BmsClient) CreateBareMetalServers(request *model.CreateBareMetalServersRequest) (*model.CreateBareMetalServersResponse, error) {
	requestDef := GenReqDefForCreateBareMetalServers()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateBareMetalServersResponse), nil
	}
}

//将挂载至裸金属服务器中的磁盘卸载；对于挂载在系统盘盘位（也就是“/dev/sda”挂载点）上的磁盘，不允许执行卸载操作；对于挂载在数据盘盘位（非“/dev/sda”挂载点）上的磁盘，支持离线卸载和在线卸载（裸金属服务器处于“运行中”状态）磁盘
func (c *BmsClient) DetachBaremetalServerVolume(request *model.DetachBaremetalServerVolumeRequest) (*model.DetachBaremetalServerVolumeResponse, error) {
	requestDef := GenReqDefForDetachBaremetalServerVolume()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DetachBaremetalServerVolumeResponse), nil
	}
}

//获取裸金属服务器详细信息，该接口支持查询裸金属服务器的计费方式，以及是否被冻结
func (c *BmsClient) ListBareMetalServerDetails(request *model.ListBareMetalServerDetailsRequest) (*model.ListBareMetalServerDetailsResponse, error) {
	requestDef := GenReqDefForListBareMetalServerDetails()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListBareMetalServerDetailsResponse), nil
	}
}

//用户根据设置的请求条件筛选裸金属服务器，并获取裸金属服务器的详细信息。该接口支持查询裸金属服务器计费方式，以及是否被冻结。
func (c *BmsClient) ListBareMetalServers(request *model.ListBareMetalServersRequest) (*model.ListBareMetalServersResponse, error) {
	requestDef := GenReqDefForListBareMetalServers()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListBareMetalServersResponse), nil
	}
}

//查询裸金属服务器的规格详情和规格的扩展信息。您可以调用此接口查询“baremetal:extBootType”参数取值，以确认某个规格是否支持快速发放
func (c *BmsClient) ListBaremetalFlavorDetailExtends(request *model.ListBaremetalFlavorDetailExtendsRequest) (*model.ListBaremetalFlavorDetailExtendsResponse, error) {
	requestDef := GenReqDefForListBaremetalFlavorDetailExtends()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListBaremetalFlavorDetailExtendsResponse), nil
	}
}

//重装裸金属服务器的操作系统。快速发放裸金属服务器支持裸金属服务器数据盘不变的情况下，使用原镜像重装系统盘。重装操作系统支持密码或者密钥注入
func (c *BmsClient) ReinstallBaremetalServerOs(request *model.ReinstallBaremetalServerOsRequest) (*model.ReinstallBaremetalServerOsResponse, error) {
	requestDef := GenReqDefForReinstallBaremetalServerOs()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ReinstallBaremetalServerOsResponse), nil
	}
}

//在裸金属服务器支持一键重置密码功能的前提下，重置裸金属服务器管理帐号（root用户或Administrator用户）的密码。可以通过6.10.1-查询是否支持一键重置密码API查询是否支持一键重置密码。
func (c *BmsClient) ResetPwdOneClick(request *model.ResetPwdOneClickRequest) (*model.ResetPwdOneClickResponse, error) {
	requestDef := GenReqDefForResetPwdOneClick()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ResetPwdOneClickResponse), nil
	}
}

//查询裸金属服务器的网卡信息，比如网卡的IP地址、MAC地址
func (c *BmsClient) ShowBaremetalServerInterfaceAttachments(request *model.ShowBaremetalServerInterfaceAttachmentsRequest) (*model.ShowBaremetalServerInterfaceAttachmentsResponse, error) {
	requestDef := GenReqDefForShowBaremetalServerInterfaceAttachments()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowBaremetalServerInterfaceAttachmentsResponse), nil
	}
}

//查询裸金属服务器挂载的磁盘信息
func (c *BmsClient) ShowBaremetalServerVolumeInfo(request *model.ShowBaremetalServerVolumeInfoRequest) (*model.ShowBaremetalServerVolumeInfoResponse, error) {
	requestDef := GenReqDefForShowBaremetalServerVolumeInfo()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowBaremetalServerVolumeInfoResponse), nil
	}
}

//查询是否支持一键重置密码
func (c *BmsClient) ShowResetPwd(request *model.ShowResetPwdRequest) (*model.ShowResetPwdResponse, error) {
	requestDef := GenReqDefForShowResetPwd()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowResetPwdResponse), nil
	}
}

//查询该租户下，所有资源的配额信息，包括已使用配额
func (c *BmsClient) ShowTenantQuota(request *model.ShowTenantQuotaRequest) (*model.ShowTenantQuotaResponse, error) {
	requestDef := GenReqDefForShowTenantQuota()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowTenantQuotaResponse), nil
	}
}

//获取Windows裸金属服务器初始安装时系统生成的管理员帐户（Administrator帐户或Cloudbase-init设置的帐户）随机密码。如果裸金属服务器是通过私有镜像创建的，请确保已安装Cloudbase-init。公共镜像默认已安装该软件
func (c *BmsClient) ShowWindowsBaremetalServerPwd(request *model.ShowWindowsBaremetalServerPwdRequest) (*model.ShowWindowsBaremetalServerPwdResponse, error) {
	requestDef := GenReqDefForShowWindowsBaremetalServerPwd()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowWindowsBaremetalServerPwdResponse), nil
	}
}

//更新裸金属服务器元数据。如果元数据中没有待更新字段，则自动添加该字段。如果元数据中已存在待更新字段，则直接更新字段值；如果元数据中的字段不再请求参数中，则保持不变
func (c *BmsClient) UpdateBaremetalServerMetadata(request *model.UpdateBaremetalServerMetadataRequest) (*model.UpdateBaremetalServerMetadataResponse, error) {
	requestDef := GenReqDefForUpdateBaremetalServerMetadata()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateBaremetalServerMetadataResponse), nil
	}
}

//清除Windows裸金属服务器初始安装时系统生成的密码记录。清除密码后，不影响裸金属服务器密码登录功能，但不能再使用获取密码功能来查询该裸金属服务器密码。如果裸金属服务器是通过私有镜像创建的，请确保已安装Cloudbase-init。公共镜像默认已安装该软件
func (c *BmsClient) WindowsBaremetalServerCleanPwd(request *model.WindowsBaremetalServerCleanPwdRequest) (*model.WindowsBaremetalServerCleanPwdResponse, error) {
	requestDef := GenReqDefForWindowsBaremetalServerCleanPwd()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.WindowsBaremetalServerCleanPwdResponse), nil
	}
}

//查询裸金属服务指定接口版本的信息
func (c *BmsClient) ShowSpecifiedVersion(request *model.ShowSpecifiedVersionRequest) (*model.ShowSpecifiedVersionResponse, error) {
	requestDef := GenReqDefForShowSpecifiedVersion()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowSpecifiedVersionResponse), nil
	}
}

//查询Job的执行状态。对于创建裸金属服务器物理机、挂卸卷等异步API，命令下发后，会返回job_id，通过job_id可以查询任务的执行状态
func (c *BmsClient) ShowJobInfos(request *model.ShowJobInfosRequest) (*model.ShowJobInfosResponse, error) {
	requestDef := GenReqDefForShowJobInfos()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowJobInfosResponse), nil
	}
}
