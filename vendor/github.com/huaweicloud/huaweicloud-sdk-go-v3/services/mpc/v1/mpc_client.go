package v1

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/mpc/v1/model"
)

type MpcClient struct {
	HcClient *http_client.HcHttpClient
}

func NewMpcClient(hcClient *http_client.HcHttpClient) *MpcClient {
	return &MpcClient{HcClient: hcClient}
}

func MpcClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//创建动图任务，用于将完整的视频文件或视频文件中的一部分转换为动态图文件，暂只支持输出GIF文件。 待转动图的视频文件需要存储在与媒体处理服务同区域的OBS桶中，且该OBS桶已授权。
func (c *MpcClient) CreateAnimatedGraphicsTask(request *model.CreateAnimatedGraphicsTaskRequest) (*model.CreateAnimatedGraphicsTaskResponse, error) {
	requestDef := GenReqDefForCreateAnimatedGraphicsTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateAnimatedGraphicsTaskResponse), nil
	}
}

//取消已下发的生成动图任务，仅支持取消正在排队中的任务。
func (c *MpcClient) DeleteAnimatedGraphicsTask(request *model.DeleteAnimatedGraphicsTaskRequest) (*model.DeleteAnimatedGraphicsTaskResponse, error) {
	requestDef := GenReqDefForDeleteAnimatedGraphicsTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteAnimatedGraphicsTaskResponse), nil
	}
}

//查询动图任务的状态。
func (c *MpcClient) ListAnimatedGraphicsTask(request *model.ListAnimatedGraphicsTaskRequest) (*model.ListAnimatedGraphicsTaskResponse, error) {
	requestDef := GenReqDefForListAnimatedGraphicsTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListAnimatedGraphicsTaskResponse), nil
	}
}

//支持独立加密，包括创建、查询、删除独立加密任务。  约束：   - 只支持转码后的文件进行加密。   - 加密的文件必须是m3u8或者mpd结尾的文件。
func (c *MpcClient) CreateEncryptTask(request *model.CreateEncryptTaskRequest) (*model.CreateEncryptTaskResponse, error) {
	requestDef := GenReqDefForCreateEncryptTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateEncryptTaskResponse), nil
	}
}

//取消独立加密任务。  约束：    只能取消正在任务队列中排队的任务。已开始加密或已完成的加密任务不能取消。
func (c *MpcClient) DeleteEncryptTask(request *model.DeleteEncryptTaskRequest) (*model.DeleteEncryptTaskResponse, error) {
	requestDef := GenReqDefForDeleteEncryptTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteEncryptTaskResponse), nil
	}
}

//查询独立加密任务状态。返回任务执行结果或当前状态。
func (c *MpcClient) ListEncryptTask(request *model.ListEncryptTaskRequest) (*model.ListEncryptTaskResponse, error) {
	requestDef := GenReqDefForListEncryptTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListEncryptTaskResponse), nil
	}
}

//创建视频解析任务，解析视频元数据。
func (c *MpcClient) CreateExtractTask(request *model.CreateExtractTaskRequest) (*model.CreateExtractTaskResponse, error) {
	requestDef := GenReqDefForCreateExtractTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateExtractTaskResponse), nil
	}
}

//取消已下发的视频解析任务，仅支持取消正在排队中的任务。
func (c *MpcClient) DeleteExtractTask(request *model.DeleteExtractTaskRequest) (*model.DeleteExtractTaskResponse, error) {
	requestDef := GenReqDefForDeleteExtractTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteExtractTaskResponse), nil
	}
}

//查询解析任务的状态和结果。
func (c *MpcClient) ListExtractTask(request *model.ListExtractTaskRequest) (*model.ListExtractTaskResponse, error) {
	requestDef := GenReqDefForListExtractTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListExtractTaskResponse), nil
	}
}

//## 典型场景 ##   合并音频多声道文件任务、重置音频文件声轨任务上报结果接口。 ## 接口功能 ##   合并音频多声道文件任务、重置音频文件声轨任务上报结果接口。 ## 接口约束 ##   无。
func (c *MpcClient) CreateMbTasksReport(request *model.CreateMbTasksReportRequest) (*model.CreateMbTasksReportResponse, error) {
	requestDef := GenReqDefForCreateMbTasksReport()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateMbTasksReportResponse), nil
	}
}

//创建声道合并任务，合并声道任务支持将每个声道各放一个文件中的片源，合并为单个音频文件。 执行合并声道的源音频文件需要存储在与媒体处理服务同区域的OBS桶中，且该OBS桶已授权。
func (c *MpcClient) CreateMergeChannelsTask(request *model.CreateMergeChannelsTaskRequest) (*model.CreateMergeChannelsTaskResponse, error) {
	requestDef := GenReqDefForCreateMergeChannelsTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateMergeChannelsTaskResponse), nil
	}
}

//创建音轨重置任务，重置音轨任务支持按人工指定关系声道layout，语言标签，转封装片源，使其满足转码输入。 执行音轨重置的源音频文件需要存储在与媒体处理服务同区域的OBS桶中，且该OBS桶已授权。
func (c *MpcClient) CreateResetTracksTask(request *model.CreateResetTracksTaskRequest) (*model.CreateResetTracksTaskResponse, error) {
	requestDef := GenReqDefForCreateResetTracksTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateResetTracksTaskResponse), nil
	}
}

//取消合并音频多声道文件。
func (c *MpcClient) DeleteMergeChannelsTask(request *model.DeleteMergeChannelsTaskRequest) (*model.DeleteMergeChannelsTaskResponse, error) {
	requestDef := GenReqDefForDeleteMergeChannelsTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteMergeChannelsTaskResponse), nil
	}
}

//取消重置音频文件声轨任务。
func (c *MpcClient) DeleteResetTracksTask(request *model.DeleteResetTracksTaskRequest) (*model.DeleteResetTracksTaskResponse, error) {
	requestDef := GenReqDefForDeleteResetTracksTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteResetTracksTaskResponse), nil
	}
}

//查询声道合并任务的状态。
func (c *MpcClient) ListMergeChannelsTask(request *model.ListMergeChannelsTaskRequest) (*model.ListMergeChannelsTaskResponse, error) {
	requestDef := GenReqDefForListMergeChannelsTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListMergeChannelsTaskResponse), nil
	}
}

//查询音轨重置任务的状态。
func (c *MpcClient) ListResetTracksTask(request *model.ListResetTracksTaskRequest) (*model.ListResetTracksTaskResponse, error) {
	requestDef := GenReqDefForListResetTracksTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListResetTracksTaskResponse), nil
	}
}

//## 典型场景 ##   创建视频增强任务。  ## 接口功能 ##   创建视频增强任务。  ## 接口约束 ##   无。
func (c *MpcClient) CreateMediaProcessTask(request *model.CreateMediaProcessTaskRequest) (*model.CreateMediaProcessTaskResponse, error) {
	requestDef := GenReqDefForCreateMediaProcessTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateMediaProcessTaskResponse), nil
	}
}

//## 典型场景 ##   取消视频增强任务。  ## 接口功能 ##   取消视频增强任务。  ## 接口约束 ##   仅可删除正在排队中的任务。
func (c *MpcClient) DeleteMediaProcessTask(request *model.DeleteMediaProcessTaskRequest) (*model.DeleteMediaProcessTaskResponse, error) {
	requestDef := GenReqDefForDeleteMediaProcessTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteMediaProcessTaskResponse), nil
	}
}

//## 典型场景 ##   查询视频增强任务。  ## 接口功能 ##   查询视频增强任务。  ## 接口约束 ##   无。
func (c *MpcClient) ListMediaProcessTask(request *model.ListMediaProcessTaskRequest) (*model.ListMediaProcessTaskResponse, error) {
	requestDef := GenReqDefForListMediaProcessTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListMediaProcessTaskResponse), nil
	}
}

//## 典型场景 ##   mpe通知mpc。 ## 接口功能 ##   mpe调用此接口通知mpc转封装等结果。 ## 接口约束 ##   无。
func (c *MpcClient) CreateMpeCallBack(request *model.CreateMpeCallBackRequest) (*model.CreateMpeCallBackResponse, error) {
	requestDef := GenReqDefForCreateMpeCallBack()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateMpeCallBackResponse), nil
	}
}

//创建视频增强模板
func (c *MpcClient) CreateQualityEnhanceTemplate(request *model.CreateQualityEnhanceTemplateRequest) (*model.CreateQualityEnhanceTemplateResponse, error) {
	requestDef := GenReqDefForCreateQualityEnhanceTemplate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateQualityEnhanceTemplateResponse), nil
	}
}

//删除用户视频增强模板。
func (c *MpcClient) DeleteQualityEnhanceTemplate(request *model.DeleteQualityEnhanceTemplateRequest) (*model.DeleteQualityEnhanceTemplateResponse, error) {
	requestDef := GenReqDefForDeleteQualityEnhanceTemplate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteQualityEnhanceTemplateResponse), nil
	}
}

//查询视频增强预置模板，返回所有结果。
func (c *MpcClient) ListQualityEnhanceDefaultTemplate(request *model.ListQualityEnhanceDefaultTemplateRequest) (*model.ListQualityEnhanceDefaultTemplateResponse, error) {
	requestDef := GenReqDefForListQualityEnhanceDefaultTemplate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListQualityEnhanceDefaultTemplateResponse), nil
	}
}

//更新视频增强模板。
func (c *MpcClient) UpdateQualityEnhanceTemplate(request *model.UpdateQualityEnhanceTemplateRequest) (*model.UpdateQualityEnhanceTemplateResponse, error) {
	requestDef := GenReqDefForUpdateQualityEnhanceTemplate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateQualityEnhanceTemplateResponse), nil
	}
}

//查询媒资转码详情
func (c *MpcClient) ListTranscodeDetail(request *model.ListTranscodeDetailRequest) (*model.ListTranscodeDetailResponse, error) {
	requestDef := GenReqDefForListTranscodeDetail()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListTranscodeDetailResponse), nil
	}
}

//取消已下发的转封装任务，仅支持取消正在排队中的任务。。
func (c *MpcClient) CancelRemuxTask(request *model.CancelRemuxTaskRequest) (*model.CancelRemuxTaskResponse, error) {
	requestDef := GenReqDefForCancelRemuxTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CancelRemuxTaskResponse), nil
	}
}

//创建转封装任务，转换音视频文件的格式，但不改变其分辨率和码率。 待转封装的媒资文件需要存储在与媒体处理服务同区域的OBS桶中，且该OBS桶已授权。
func (c *MpcClient) CreateRemuxTask(request *model.CreateRemuxTaskRequest) (*model.CreateRemuxTaskResponse, error) {
	requestDef := GenReqDefForCreateRemuxTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateRemuxTaskResponse), nil
	}
}

//对失败的转封装任务进行重试。
func (c *MpcClient) CreateRetryRemuxTask(request *model.CreateRetryRemuxTaskRequest) (*model.CreateRetryRemuxTaskResponse, error) {
	requestDef := GenReqDefForCreateRetryRemuxTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateRetryRemuxTaskResponse), nil
	}
}

//删除转封装任务
func (c *MpcClient) DeleteRemuxTask(request *model.DeleteRemuxTaskRequest) (*model.DeleteRemuxTaskResponse, error) {
	requestDef := GenReqDefForDeleteRemuxTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteRemuxTaskResponse), nil
	}
}

//查询转封装任务状态。
func (c *MpcClient) ListRemuxTask(request *model.ListRemuxTaskRequest) (*model.ListRemuxTaskResponse, error) {
	requestDef := GenReqDefForListRemuxTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListRemuxTaskResponse), nil
	}
}

//新建转码模板组，最多支持一进六出。
func (c *MpcClient) CreateTemplateGroup(request *model.CreateTemplateGroupRequest) (*model.CreateTemplateGroupResponse, error) {
	requestDef := GenReqDefForCreateTemplateGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateTemplateGroupResponse), nil
	}
}

//删除转码模板组。
func (c *MpcClient) DeleteTemplateGroup(request *model.DeleteTemplateGroupRequest) (*model.DeleteTemplateGroupResponse, error) {
	requestDef := GenReqDefForDeleteTemplateGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteTemplateGroupResponse), nil
	}
}

//查询转码模板组列表。
func (c *MpcClient) ListTemplateGroup(request *model.ListTemplateGroupRequest) (*model.ListTemplateGroupResponse, error) {
	requestDef := GenReqDefForListTemplateGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListTemplateGroupResponse), nil
	}
}

//修改模板组接口。
func (c *MpcClient) UpdateTemplateGroup(request *model.UpdateTemplateGroupRequest) (*model.UpdateTemplateGroupResponse, error) {
	requestDef := GenReqDefForUpdateTemplateGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateTemplateGroupResponse), nil
	}
}

//新建截图任务，视频截图将从首帧开始，按设置的时间间隔截图，最后截取末帧。 待截图的视频文件需要存储在与媒体处理服务同区域的OBS桶中，且该OBS桶已授权。  约束：   暂只支持生成JPG格式的图片文件。
func (c *MpcClient) CreateThumbnailsTask(request *model.CreateThumbnailsTaskRequest) (*model.CreateThumbnailsTaskResponse, error) {
	requestDef := GenReqDefForCreateThumbnailsTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateThumbnailsTaskResponse), nil
	}
}

//取消已下发截图任务。 只能取消已接受尚在队列中等待处理的任务，已完成或正在执行阶段的任务不能取消。
func (c *MpcClient) DeleteThumbnailsTask(request *model.DeleteThumbnailsTaskRequest) (*model.DeleteThumbnailsTaskResponse, error) {
	requestDef := GenReqDefForDeleteThumbnailsTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteThumbnailsTaskResponse), nil
	}
}

//查询截图任务状态。返回任务执行结果，包括状态、输入、输出等信息。
func (c *MpcClient) ListThumbnailsTask(request *model.ListThumbnailsTaskRequest) (*model.ListThumbnailsTaskResponse, error) {
	requestDef := GenReqDefForListThumbnailsTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListThumbnailsTaskResponse), nil
	}
}

//新建转码任务可以将视频进行转码，并在转码过程中压制水印、视频截图等。视频转码前需要配置转码模板。 待转码的音视频需要存储在与媒体处理服务同区域的OBS桶中，且该OBS桶已授权。
func (c *MpcClient) CreateTranscodingTask(request *model.CreateTranscodingTaskRequest) (*model.CreateTranscodingTaskResponse, error) {
	requestDef := GenReqDefForCreateTranscodingTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateTranscodingTaskResponse), nil
	}
}

//取消已下发转码任务。 只能取消正在转码任务队列中排队的转码任务。已开始转码或已完成的转码任务不能取消。
func (c *MpcClient) DeleteTranscodingTask(request *model.DeleteTranscodingTaskRequest) (*model.DeleteTranscodingTaskResponse, error) {
	requestDef := GenReqDefForDeleteTranscodingTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteTranscodingTaskResponse), nil
	}
}

//查询转码任务状态。
func (c *MpcClient) ListTranscodingTask(request *model.ListTranscodingTaskRequest) (*model.ListTranscodingTaskResponse, error) {
	requestDef := GenReqDefForListTranscodingTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListTranscodingTaskResponse), nil
	}
}

//新建转码模板，采用自定义的模板转码。
func (c *MpcClient) CreateTransTemplate(request *model.CreateTransTemplateRequest) (*model.CreateTransTemplateResponse, error) {
	requestDef := GenReqDefForCreateTransTemplate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateTransTemplateResponse), nil
	}
}

//删除转码模板。
func (c *MpcClient) DeleteTemplate(request *model.DeleteTemplateRequest) (*model.DeleteTemplateResponse, error) {
	requestDef := GenReqDefForDeleteTemplate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteTemplateResponse), nil
	}
}

//查询用户自定义转码配置模板。 支持指定模板ID查询，或分页全量查询。转码配置模板ID，最多10个。
func (c *MpcClient) ListTemplate(request *model.ListTemplateRequest) (*model.ListTemplateResponse, error) {
	requestDef := GenReqDefForListTemplate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListTemplateResponse), nil
	}
}

//更新转码模板。
func (c *MpcClient) UpdateTransTemplate(request *model.UpdateTransTemplateRequest) (*model.UpdateTransTemplateResponse, error) {
	requestDef := GenReqDefForUpdateTransTemplate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateTransTemplateResponse), nil
	}
}

//自定义水印模板。
func (c *MpcClient) CreateWatermarkTemplate(request *model.CreateWatermarkTemplateRequest) (*model.CreateWatermarkTemplateResponse, error) {
	requestDef := GenReqDefForCreateWatermarkTemplate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateWatermarkTemplateResponse), nil
	}
}

//删除自定义水印模板。
func (c *MpcClient) DeleteWatermarkTemplate(request *model.DeleteWatermarkTemplateRequest) (*model.DeleteWatermarkTemplateResponse, error) {
	requestDef := GenReqDefForDeleteWatermarkTemplate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteWatermarkTemplateResponse), nil
	}
}

//查询自定义水印模板。支持指定模板ID查询，或分页全量查询。
func (c *MpcClient) ListWatermarkTemplate(request *model.ListWatermarkTemplateRequest) (*model.ListWatermarkTemplateResponse, error) {
	requestDef := GenReqDefForListWatermarkTemplate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListWatermarkTemplateResponse), nil
	}
}

//更新自定义水印模板。
func (c *MpcClient) UpdateWatermarkTemplate(request *model.UpdateWatermarkTemplateRequest) (*model.UpdateWatermarkTemplateResponse, error) {
	requestDef := GenReqDefForUpdateWatermarkTemplate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateWatermarkTemplateResponse), nil
	}
}
