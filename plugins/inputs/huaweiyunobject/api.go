package huaweiyunobject

import (
	"fmt"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	rms "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1"
	rmsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1/model"
	region "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1/region"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func (ag *agent) genAPIClient() {
	ak := ag.AccessKeyID
	sk := ag.AccessKeySecret

	auth := global.NewCredentialsBuilder().
		WithAk(ak).
		WithSk(sk).
		Build()

	ag.apiClient = rms.NewRmsClient(
		rms.RmsClientBuilder().
			WithRegion(region.ValueOf("cn-north-4")).
			WithCredential(auth).
			Build())

}

func (ag *agent) listAllResources() {
	ak := ag.AccessKeyID
	sk := ag.AccessKeySecret

	auth := global.NewCredentialsBuilder().
		WithAk(ak).
		WithSk(sk).
		Build()

	client := rms.NewRmsClient(
		rms.RmsClientBuilder().
			WithRegion(region.ValueOf("cn-north-4")).
			WithCredential(auth).
			Build())

	request := &rmsmodel.ListAllResourcesRequest{}
	response, err := client.ListAllResources(request)
	if err == nil {
		_ = response
	} else {
		fmt.Println(err)
	}
}

func (ag *agent) listResources(provider, typ string) ([]rmsmodel.ResourceEntity, error) {

	var results []rmsmodel.ResourceEntity

	var response *rmsmodel.ListResourcesResponse
	var err error

	request := &rmsmodel.ListResourcesRequest{}
	request.Provider = provider
	request.Type = typ

	marker := ""

	for true {
		if marker != "" {
			request.Marker = &marker
		}
		for i := 0; i < 5; i++ {
			ag.limiter.Wait(ag.ctx)
			response, err = ag.apiClient.ListResources(request)
			if err == nil {
				break
			}
			select {
			case <-ag.ctx.Done():
				break
			default:
			}
			datakit.SleepContext(ag.ctx, time.Second*5)
		}
		if err != nil {
			moduleLogger.Errorf("%s, request: %s", err, request.String())
			return nil, err
		}

		if response != nil && response.Resources != nil {
			results = append(results, *response.Resources...)
		}

		marker = ""
		if response != nil && response.PageInfo != nil && response.PageInfo.NextMarker != nil {
			marker = *response.PageInfo.NextMarker
		}

		if marker == "" {
			break
		}
	}

	return results, nil
}

func (ag *agent) listProviders() ([]rmsmodel.ResourceProviderResponse, error) {

	for i := 0; i < 5; i++ {
		request := &rmsmodel.ListProvidersRequest{}
		response, err := ag.apiClient.ListProviders(request)
		if err == nil {
			if response.ResourceProviders != nil {
				return *response.ResourceProviders, nil
			}
			break
		}
		moduleLogger.Errorf("%s", err)

		select {
		case <-ag.ctx.Done():
			break
		default:
		}

		datakit.SleepContext(ag.ctx, time.Second*5)
	}

	return nil, nil
}
