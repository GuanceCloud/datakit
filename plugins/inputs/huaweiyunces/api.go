package huaweiyunces

import (
	"fmt"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	ces "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1"
	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	cesregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/region"
	ecs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	ecsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
	ecsregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/region"
)

func (ag *agent) listServersDetails() []string {

	auth := basic.NewCredentialsBuilder().
		WithAk(ag.AccessKeyID).
		WithSk(ag.AccessKeySecret).
		WithProjectId(ag.ProjectID).
		Build()

	reg := ag.RegionID
	if reg == "" {
		reg = ag.EndPoint
	}

	cli := ecs.EcsClientBuilder().WithRegion(ecsregion.ValueOf(reg)).WithCredential(auth).Build()
	esccli := ecs.NewEcsClient(cli)

	for i := 0; i < 3; i++ {
		req := &ecsmodel.ListServersDetailsRequest{}
		resp, err := esccli.ListServersDetails(req)
		if err != nil {
			moduleLogger.Error(err)
			time.Sleep(time.Second * 3)
			continue
		}
		instids := []string{}
		for _, r := range *resp.Servers {
			instids = append(instids, r.Id)
		}
		return instids
	}

	return nil
}

func (ag *agent) genHWClient() {
	auth := basic.NewCredentialsBuilder().
		WithAk(ag.AccessKeyID).
		WithSk(ag.AccessKeySecret).
		WithProjectId(ag.ProjectID).
		Build()

	reg := ag.RegionID
	if reg == "" {
		reg = ag.EndPoint
	}

	ag.client = ces.NewCesClient(
		ces.CesClientBuilder().
			WithRegion(cesregion.ValueOf(reg)).
			//WithRegion(region.CN_NORTH_4).
			WithCredential(auth).
			Build())
}

func (ag *agent) tryShowMetricData(req *metricsRequest) (*cesmodel.ShowMetricDataResponse, error) {
	var resp *cesmodel.ShowMetricDataResponse
	var err error
	wt := time.Second * 5
	for i := 0; i < 5; i++ {
		resp, err = ag.showMetricData(req)
		if err == nil {
			return resp, nil
		}
		time.Sleep(wt)
		wt += time.Second * 5
	}
	return nil, err
}

func (ag *agent) showMetricData(req *metricsRequest) (*cesmodel.ShowMetricDataResponse, error) {

	nt := time.Now().Truncate(time.Second)
	endTime := nt.Unix() * 1000
	var startTime int64
	if req.lastTime.IsZero() {
		startTime = nt.Add(-5*time.Minute).Unix() * 1000
	} else {
		if nt.Sub(req.lastTime) < req.interval {
			return nil, nil
		}
		startTime = req.lastTime.Add(-(ag.Delay.Duration)).Unix() * 1000
	}

	logEndtime := time.Unix(endTime/int64(1000), 0)
	logStarttime := time.Unix(startTime/int64(1000), 0)

	req.to = endTime
	req.from = startTime

	request := &cesmodel.ShowMetricDataRequest{}
	request.Namespace = req.namespace
	request.MetricName = req.metricname

	numdim := len(req.dimensoions)
	if numdim > 4 {
		moduleLogger.Warnf("only support up to 4 dimensions, get %d dimensions", numdim)
		numdim = 4
	}

	if numdim > 0 {
		request.Dim0 = fmt.Sprintf("%s,%s", req.dimensoions[0].Name, req.dimensoions[0].Value)
	}
	if numdim > 1 {
		dim1 := fmt.Sprintf("%s,%s", req.dimensoions[1].Name, req.dimensoions[1].Value)
		request.Dim1 = &dim1
	}
	if numdim > 2 {
		dim2 := fmt.Sprintf("%s,%s", req.dimensoions[2].Name, req.dimensoions[2].Value)
		request.Dim2 = &dim2
	}
	if numdim > 3 {
		dim3 := fmt.Sprintf("%s,%s", req.dimensoions[3].Name, req.dimensoions[3].Value)
		request.Dim3 = &dim3
	}

	//request.Dim0 = "instance_id,d735ce9a-342e-476e-920d-5243032133d4"
	request.Filter = cesmodel.GetShowMetricDataRequestFilterEnum().AVERAGE
	request.Period = int32(req.period)
	request.From = req.from
	request.To = req.to
	response, err := ag.client.ShowMetricData(request)
	if err != nil {
		moduleLogger.Errorf("fail to get metric: Namespace = %s, MetricName = %s, Period = %v, StartTime = %v(%s), EndTime = %v(%s), Dimensions = %s, err: %s", req.namespace, req.metricname, req.period, req.from, logStarttime, req.to, logEndtime, req.dimensoions, err)
		if response != nil {
			moduleLogger.Errorf("%s", response.String())
		}
		return nil, err
	}

	req.lastTime = nt

	nc := 0
	if response != nil && response.Datapoints != nil {
		nc = len(*response.Datapoints)
	}

	moduleLogger.Debugf("get %d datapoints: Namespace = %s, MetricName = %s, Filter = %s, Period = %v, InterVal = %v, StartTime = %v(%s), EndTime = %v(%s), Dimensions = %s", nc, req.namespace, req.metricname, req.filter, req.period, req.interval, req.from, logStarttime, req.to, logEndtime, req.dimensoions)

	return response, nil
}
