package v2

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/live/v2/model"
)

type LiveClient struct {
	HcClient *http_client.HcHttpClient
}

func NewLiveClient(hcClient *http_client.HcHttpClient) *LiveClient {
	return &LiveClient{HcClient: hcClient}
}

func LiveClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//查询播放域名带宽数据。  最大查询跨度31天，最大查询周期90天。
func (c *LiveClient) ListBandwidthDetail(request *model.ListBandwidthDetailRequest) (*model.ListBandwidthDetailResponse, error) {
	requestDef := GenReqDefForListBandwidthDetail()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListBandwidthDetailResponse), nil
	}
}

//查询指定时间范围内播放带宽峰值。  最大查询跨度31天，最大查询周期90天。
func (c *LiveClient) ListDomainBandwidthPeak(request *model.ListDomainBandwidthPeakRequest) (*model.ListDomainBandwidthPeakResponse, error) {
	requestDef := GenReqDefForListDomainBandwidthPeak()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListDomainBandwidthPeakResponse), nil
	}
}

//查询播放域名流量数据。  最大查询跨度31天，最大查询周期90天。
func (c *LiveClient) ListDomainTrafficDetail(request *model.ListDomainTrafficDetailRequest) (*model.ListDomainTrafficDetailResponse, error) {
	requestDef := GenReqDefForListDomainTrafficDetail()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListDomainTrafficDetailResponse), nil
	}
}

//查询指定时间范围内流量汇总量。  最大查询跨度31天，最大查询周期90天。
func (c *LiveClient) ListDomainTrafficSummary(request *model.ListDomainTrafficSummaryRequest) (*model.ListDomainTrafficSummaryResponse, error) {
	requestDef := GenReqDefForListDomainTrafficSummary()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListDomainTrafficSummaryResponse), nil
	}
}

//查询历史推流列表。  最大查询跨度1天，最大查询周期7天。
func (c *LiveClient) ListHistoryStreams(request *model.ListHistoryStreamsRequest) (*model.ListHistoryStreamsResponse, error) {
	requestDef := GenReqDefForListHistoryStreams()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListHistoryStreamsResponse), nil
	}
}

//查询直播拉流HTTP状态码接口。  获取加速域名1分钟粒度的HTTP返回码  最大查询跨度不能超过24小时，最大查询周期7天。
func (c *LiveClient) ListQueryHttpCode(request *model.ListQueryHttpCodeRequest) (*model.ListQueryHttpCodeResponse, error) {
	requestDef := GenReqDefForListQueryHttpCode()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListQueryHttpCodeResponse), nil
	}
}

//查询直播租户每小时录制的最大并发数，计算1小时内每分钟的并发总路数，取最大值做为统计值。  最大查询跨度31天，最大查询周期90天。
func (c *LiveClient) ListRecordData(request *model.ListRecordDataRequest) (*model.ListRecordDataResponse, error) {
	requestDef := GenReqDefForListRecordData()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListRecordDataResponse), nil
	}
}

//查询直播域名每小时的截图数量。  最大查询跨度31天，最大查询周期90天。
func (c *LiveClient) ListSnapshotData(request *model.ListSnapshotDataRequest) (*model.ListSnapshotDataResponse, error) {
	requestDef := GenReqDefForListSnapshotData()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListSnapshotDataResponse), nil
	}
}

//查询直播域名每小时的转码时长数据。  最大查询跨度31天，最大查询周期90天。
func (c *LiveClient) ListTranscodeData(request *model.ListTranscodeDataRequest) (*model.ListTranscodeDataResponse, error) {
	requestDef := GenReqDefForListTranscodeData()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListTranscodeDataResponse), nil
	}
}

//查询观众趋势。  最大查询跨度7天，最大查询周期90天。
func (c *LiveClient) ListUsersOfStream(request *model.ListUsersOfStreamRequest) (*model.ListUsersOfStreamResponse, error) {
	requestDef := GenReqDefForListUsersOfStream()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListUsersOfStreamResponse), nil
	}
}

//查询域名维度推流路数接口。  最大查询跨度31天，最大查询周期90天。
func (c *LiveClient) ShowStreamCount(request *model.ShowStreamCountRequest) (*model.ShowStreamCountResponse, error) {
	requestDef := GenReqDefForShowStreamCount()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowStreamCountResponse), nil
	}
}

//查询播放画像信息。  最大查询跨度1天，最大查询周期31天。
func (c *LiveClient) ShowStreamPortrait(request *model.ShowStreamPortraitRequest) (*model.ShowStreamPortraitResponse, error) {
	requestDef := GenReqDefForShowStreamPortrait()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowStreamPortraitResponse), nil
	}
}

//查询上行带宽数据。  最大查询跨度31天，最大查询周期90天。
func (c *LiveClient) ShowUpBandwidth(request *model.ShowUpBandwidthRequest) (*model.ShowUpBandwidthResponse, error) {
	requestDef := GenReqDefForShowUpBandwidth()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowUpBandwidthResponse), nil
	}
}

//查询推流监控码率数据接口。  最大查询跨度6小时，最大查询周期7天。
func (c *LiveClient) ListSingleStreamBitrate(request *model.ListSingleStreamBitrateRequest) (*model.ListSingleStreamBitrateResponse, error) {
	requestDef := GenReqDefForListSingleStreamBitrate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListSingleStreamBitrateResponse), nil
	}
}

//查询推流帧率数据接口。  最大查询跨度6小时，最大查询周期7天。
func (c *LiveClient) ListSingleStreamFramerate(request *model.ListSingleStreamFramerateRequest) (*model.ListSingleStreamFramerateResponse, error) {
	requestDef := GenReqDefForListSingleStreamFramerate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListSingleStreamFramerateResponse), nil
	}
}
