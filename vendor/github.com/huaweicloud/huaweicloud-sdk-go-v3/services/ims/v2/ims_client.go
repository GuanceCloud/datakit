package v2

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ims/v2/model"
)

type ImsClient struct {
	HcClient *http_client.HcHttpClient
}

func NewImsClient(hcClient *http_client.HcHttpClient) *ImsClient {
	return &ImsClient{HcClient: hcClient}
}

func ImsClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//该接口为扩展接口，主要用于镜像共享时用户将多个镜像共享给多个用户。 该接口为异步接口，返回job_id说明任务下发成功，查询异步任务状态，如果是success说明任务执行成功，如果是failed说明任务执行失败。如何查询异步任务，请参见异步任务查询。
func (c *ImsClient) BatchAddMembers(request *model.BatchAddMembersRequest) (*model.BatchAddMembersResponse, error) {
	requestDef := GenReqDefForBatchAddMembers()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchAddMembersResponse), nil
	}
}

//该接口为扩展接口，主要用于取消镜像共享。 该接口为异步接口，返回job_id说明任务下发成功，查询异步任务状态，如果是success说明任务执行成功，如果是failed说明任务执行失败。如何查询异步任务，请参见异步任务查询。
func (c *ImsClient) BatchDeleteMembers(request *model.BatchDeleteMembersRequest) (*model.BatchDeleteMembersResponse, error) {
	requestDef := GenReqDefForBatchDeleteMembers()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchDeleteMembersResponse), nil
	}
}

//该接口为扩展接口，主要用于用户接受或者拒绝多个共享镜像时批量更新镜像成员的状态。 该接口为异步接口，返回job_id说明任务下发成功，查询异步任务状态，如果是success说明任务执行成功，如果是failed说明任务执行失败。如何查询异步任务，请参见异步任务查询。
func (c *ImsClient) BatchUpdateMembers(request *model.BatchUpdateMembersRequest) (*model.BatchUpdateMembersResponse, error) {
	requestDef := GenReqDefForBatchUpdateMembers()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchUpdateMembersResponse), nil
	}
}

//该接口为扩展接口，用户在一个区域制作的私有镜像，可以通过跨Region复制镜像将镜像复制到其他区域，在其他区域发放相同类型的云服务器，帮助用户实现区域间的业务迁移。 该接口为异步接口，返回job_id说明任务下发成功，查询异步任务状态，如果是success说明任务执行成功，如果是failed说明任务执行失败。 如何查询异步任务，请参见异步任务进度查询。
func (c *ImsClient) CopyImageCrossRegion(request *model.CopyImageCrossRegionRequest) (*model.CopyImageCrossRegionResponse, error) {
	requestDef := GenReqDefForCopyImageCrossRegion()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CopyImageCrossRegionResponse), nil
	}
}

//该接口为扩展接口，主要用于用户将一个已有镜像复制为另一个镜像。复制镜像时，可以更改镜像的加密等属性，以满足不同的场景。 该接口为异步接口，返回job_id说明任务下发成功，查询异步任务状态，如果是success说明任务执行成功，如果是failed说明任务执行失败。如何查询异步任务，请参见异步任务查询。
func (c *ImsClient) CopyImageInRegion(request *model.CopyImageInRegionRequest) (*model.CopyImageInRegionResponse, error) {
	requestDef := GenReqDefForCopyImageInRegion()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CopyImageInRegionResponse), nil
	}
}

//使用上传至OBS桶中的外部数据卷镜像文件制作数据镜像。作为异步接口，调用成功，只是说明后台收到了制作请求，镜像是否制作成功需要通过异步任务查询接口查询该任务的执行状态。具体请参考异步任务查询。
func (c *ImsClient) CreateDataImage(request *model.CreateDataImageRequest) (*model.CreateDataImageResponse, error) {
	requestDef := GenReqDefForCreateDataImage()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateDataImageResponse), nil
	}
}

//本接口用于制作私有镜像，支持： 使用云服务器制作私有镜像。 使用上传至OBS桶中的外部镜像文件制作私有镜像。 使用数据卷制作系统盘镜像。 作为异步接口，调用成功，只是说明云平台收到了制作请求，镜像是否制作成功需要通过异步任务查询接口查询该任务的执行状态，具体请参考异步任务查询。
func (c *ImsClient) CreateImage(request *model.CreateImageRequest) (*model.CreateImageResponse, error) {
	requestDef := GenReqDefForCreateImage()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateImageResponse), nil
	}
}

//该接口主要用于为某个镜像增加或修改一个自定义标签。通过自定义标签，用户可以将镜像进行分类。
func (c *ImsClient) CreateOrUpdateTags(request *model.CreateOrUpdateTagsRequest) (*model.CreateOrUpdateTagsResponse, error) {
	requestDef := GenReqDefForCreateOrUpdateTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateOrUpdateTagsResponse), nil
	}
}

//使用云服务器或者云服务器备份制作整机镜像。作为异步接口，调用成功，只是说明后台收到了制作整机镜像的请求，镜像是否制作成功需要通过异步任务查询接口查询该任务的执行状态，具体请参考异步任务查询。
func (c *ImsClient) CreateWholeImage(request *model.CreateWholeImageRequest) (*model.CreateWholeImageResponse, error) {
	requestDef := GenReqDefForCreateWholeImage()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateWholeImageResponse), nil
	}
}

//该接口为扩展接口，用于用户将自己的私有镜像导出到指定的OBS桶中。
func (c *ImsClient) ExportImage(request *model.ExportImageRequest) (*model.ExportImageResponse, error) {
	requestDef := GenReqDefForExportImage()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ExportImageResponse), nil
	}
}

//使用上传至OBS桶中的超大外部镜像文件制作私有镜像，目前仅支持RAW或ZVHD2格式镜像文件。且要求镜像文件大小不能超过1TB。 由于快速导入功能要求提前转换镜像文件格式为RAW或ZVHD2格式，因此镜像文件小于128GB时推荐您优先使用常规的创建私有镜像的方式。 作为异步接口，调用成功，只是说明后台收到了制作请求，镜像是否制作成功需要通过异步任务查询接口查询该任务的执行状态，具体请参考异步任务查询。
func (c *ImsClient) ImportImageQuick(request *model.ImportImageQuickRequest) (*model.ImportImageQuickResponse, error) {
	requestDef := GenReqDefForImportImageQuick()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ImportImageQuickResponse), nil
	}
}

//根据不同条件查询镜像列表信息。 可以在URI后面用‘?’和‘&’添加不同的查询条件组合，请参考请求样例。
func (c *ImsClient) ListImages(request *model.ListImagesRequest) (*model.ListImagesResponse, error) {
	requestDef := GenReqDefForListImages()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListImagesResponse), nil
	}
}

//查询当前区域弹性云服务器的OS兼容性列表。
func (c *ImsClient) ListOsVersions(request *model.ListOsVersionsRequest) (*model.ListOsVersionsResponse, error) {
	requestDef := GenReqDefForListOsVersions()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListOsVersionsResponse), nil
	}
}

//根据不同条件查询镜像标签列表信息。
func (c *ImsClient) ListTags(request *model.ListTagsRequest) (*model.ListTagsResponse, error) {
	requestDef := GenReqDefForListTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListTagsResponse), nil
	}
}

//该接口用于将镜像文件注册为云平台未初始化的私有镜像。 使用该接口注册镜像的具体步骤如下： 将镜像文件上传到OBS个人桶中。具体操作请参见《对象存储服务客户端指南（OBS Browser）》或《对象存储服务API参考》。 使用创建镜像元数据接口创建镜像元数据。调用成功后，保存该镜像的ID。创建镜像元数据请参考创建镜像元数据（OpenStack原生）。 根据2得到的镜像ID，使用注册镜像接口注册OBS桶中的镜像文件。 注册镜像接口作为异步接口，调用成功后，说明后台收到了注册请求。需要根据镜像ID查询该镜像状态验证镜像注册是否成功。当镜像状态变为“active”时，表示镜像注册成功。 如何查询异步任务，请参见异步任务查询。
func (c *ImsClient) RegisterImage(request *model.RegisterImageRequest) (*model.RegisterImageResponse, error) {
	requestDef := GenReqDefForRegisterImage()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.RegisterImageResponse), nil
	}
}

//该接口为扩展接口，主要用于查询租户在当前Region的私有镜像的配额数量。
func (c *ImsClient) ShowImageQuota(request *model.ShowImageQuotaRequest) (*model.ShowImageQuotaResponse, error) {
	requestDef := GenReqDefForShowImageQuota()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowImageQuotaResponse), nil
	}
}

//更新镜像信息接口，主要用于镜像属性的修改。当前仅支持可用（active）状态的镜像更新相关信息。
func (c *ImsClient) UpdateImage(request *model.UpdateImageRequest) (*model.UpdateImageResponse, error) {
	requestDef := GenReqDefForUpdateImage()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateImageResponse), nil
	}
}

//该接口为扩展接口，主要用于查询异步接口执行情况，比如查询导出镜像任务的执行状态。
func (c *ImsClient) ShowJob(request *model.ShowJobRequest) (*model.ShowJobResponse, error) {
	requestDef := GenReqDefForShowJob()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowJobResponse), nil
	}
}

//用户共享镜像给其他用户时，使用该接口向该镜像成员中添加接受镜像用户的项目ID。
func (c *ImsClient) GlanceAddImageMember(request *model.GlanceAddImageMemberRequest) (*model.GlanceAddImageMemberResponse, error) {
	requestDef := GenReqDefForGlanceAddImageMember()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.GlanceAddImageMemberResponse), nil
	}
}

//创建镜像元数据。调用创建镜像元数据接口成功后，只是创建了镜像的元数据，镜像对应的实际镜像文件并不存在
func (c *ImsClient) GlanceCreateImageMetadata(request *model.GlanceCreateImageMetadataRequest) (*model.GlanceCreateImageMetadataResponse, error) {
	requestDef := GenReqDefForGlanceCreateImageMetadata()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.GlanceCreateImageMetadataResponse), nil
	}
}

//该接口主要用于为某个镜像添加一个自定义标签。通过自定义标签，用户可以将镜像进行分类。
func (c *ImsClient) GlanceCreateTag(request *model.GlanceCreateTagRequest) (*model.GlanceCreateTagResponse, error) {
	requestDef := GenReqDefForGlanceCreateTag()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.GlanceCreateTagResponse), nil
	}
}

//该接口主要用于删除镜像，用户可以通过该接口将自己的私有镜像删除。
func (c *ImsClient) GlanceDeleteImage(request *model.GlanceDeleteImageRequest) (*model.GlanceDeleteImageResponse, error) {
	requestDef := GenReqDefForGlanceDeleteImage()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.GlanceDeleteImageResponse), nil
	}
}

//该接口用于取消对某个用户的镜像共享。
func (c *ImsClient) GlanceDeleteImageMember(request *model.GlanceDeleteImageMemberRequest) (*model.GlanceDeleteImageMemberResponse, error) {
	requestDef := GenReqDefForGlanceDeleteImageMember()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.GlanceDeleteImageMemberResponse), nil
	}
}

//该接口主要用于删除某个镜像的自定义标签，通过该接口，用户可以将私有镜像中一些不用的标签删除。
func (c *ImsClient) GlanceDeleteTag(request *model.GlanceDeleteTagRequest) (*model.GlanceDeleteTagResponse, error) {
	requestDef := GenReqDefForGlanceDeleteTag()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.GlanceDeleteTagResponse), nil
	}
}

//该接口主要用于查询镜像成员列表视图，通过视图，用户可以了解到镜像成员包含哪些属性，同时也可以了解每个属性的数据类型。
func (c *ImsClient) GlanceListImageMemberSchemas(request *model.GlanceListImageMemberSchemasRequest) (*model.GlanceListImageMemberSchemasResponse, error) {
	requestDef := GenReqDefForGlanceListImageMemberSchemas()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.GlanceListImageMemberSchemasResponse), nil
	}
}

//该接口用于共享镜像过程中，获取接受该镜像的成员列表。
func (c *ImsClient) GlanceListImageMembers(request *model.GlanceListImageMembersRequest) (*model.GlanceListImageMembersResponse, error) {
	requestDef := GenReqDefForGlanceListImageMembers()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.GlanceListImageMembersResponse), nil
	}
}

//该接口主要用于查询镜像列表视图，通过该接口用户可以了解到镜像列表的详细情况和数据结构。
func (c *ImsClient) GlanceListImageSchemas(request *model.GlanceListImageSchemasRequest) (*model.GlanceListImageSchemasResponse, error) {
	requestDef := GenReqDefForGlanceListImageSchemas()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.GlanceListImageSchemasResponse), nil
	}
}

//获取镜像列表。 使用本接口查询镜像列表时，需要使用分页查询才能返回全部的镜像列表。 分页说明 分页是指返回一组镜像的一个子集，在返回的时候会存在下个子集的链接和首个子集的链接，默认返回的子集中数量为25，用户也可以通过使用limit和marker两个参数自己分页，指定返回子集中需要返回的数量。 响应中的参数first是查询首页的URL。next是查询下一页的URL。当查询镜像列表最后一页时，不存在next。
func (c *ImsClient) GlanceListImages(request *model.GlanceListImagesRequest) (*model.GlanceListImagesResponse, error) {
	requestDef := GenReqDefForGlanceListImages()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.GlanceListImagesResponse), nil
	}
}

//查询单个镜像详情，用户可以通过该接口查询单个私有或者公共镜像的详情
func (c *ImsClient) GlanceShowImage(request *model.GlanceShowImageRequest) (*model.GlanceShowImageResponse, error) {
	requestDef := GenReqDefForGlanceShowImage()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.GlanceShowImageResponse), nil
	}
}

//该接口主要用于镜像共享中查询某个镜像成员的详情。
func (c *ImsClient) GlanceShowImageMember(request *model.GlanceShowImageMemberRequest) (*model.GlanceShowImageMemberResponse, error) {
	requestDef := GenReqDefForGlanceShowImageMember()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.GlanceShowImageMemberResponse), nil
	}
}

//该接口主要用于查询镜像成员视图，通过视图，用户可以了解到镜像成员包含哪些属性，同时也可以了解每个属性的数据类型。
func (c *ImsClient) GlanceShowImageMemberSchemas(request *model.GlanceShowImageMemberSchemasRequest) (*model.GlanceShowImageMemberSchemasResponse, error) {
	requestDef := GenReqDefForGlanceShowImageMemberSchemas()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.GlanceShowImageMemberSchemasResponse), nil
	}
}

//该接口主要用于查询镜像视图，通过视图，用户可以了解到镜像包含哪些属性，同时也可以了解每个属性的数据类型等。
func (c *ImsClient) GlanceShowImageSchemas(request *model.GlanceShowImageSchemasRequest) (*model.GlanceShowImageSchemasResponse, error) {
	requestDef := GenReqDefForGlanceShowImageSchemas()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.GlanceShowImageSchemasResponse), nil
	}
}

//修改镜像信息
func (c *ImsClient) GlanceUpdateImage(request *model.GlanceUpdateImageRequest) (*model.GlanceUpdateImageResponse, error) {
	requestDef := GenReqDefForGlanceUpdateImage()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.GlanceUpdateImageResponse), nil
	}
}

//用户接受或者拒绝共享镜像时，使用该接口更新镜像成员的状态。
func (c *ImsClient) GlanceUpdateImageMember(request *model.GlanceUpdateImageMemberRequest) (*model.GlanceUpdateImageMemberResponse, error) {
	requestDef := GenReqDefForGlanceUpdateImageMember()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.GlanceUpdateImageMemberResponse), nil
	}
}
