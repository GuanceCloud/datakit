package huaweiyunces

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	ces "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1"
	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	cesregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/region"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/region"

	ecs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	ecsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
	ecsregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/region"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func (ag *agent) runOld() {

	ag.genAPIClient()

	if ag.cesClient == nil {
		return
	}

	if err := ag.genReqsOld(ag.ctx); err != nil {
		if ag.isTestOnce() {
			ag.testError = err
		}
		return
	}

	if len(ag.reqs) == 0 {
		moduleLogger.Warnf("no metric found")
		if ag.isTestOnce() {
			ag.testError = fmt.Errorf("no metric found")
		}
		return
	}
	moduleLogger.Debugf("%d reqs", len(ag.reqs))

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		for _, req := range ag.reqs {

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			resp, err := ag.showMetricData(req)
			if err != nil {
				if ag.isTestOnce() {
					ag.testError = err
					return
				}
				continue
			}

			if resp == nil || resp.Datapoints == nil {
				continue
			}

			metricSetName := formatMeasurement(req.namespace)

			for _, dp := range *resp.Datapoints {

				select {
				case <-ag.ctx.Done():
					return
				default:
				}

				tags := map[string]string{}
				extendTags(tags, req.tags, false)
				extendTags(tags, ag.Tags, false)
				for _, k := range req.dimsOld {
					tags[k.Name] = k.Value
				}
				tags["unit"] = *dp.Unit

				fields := map[string]interface{}{}

				var val float64
				switch req.filter {
				case "max":
					if dp.Max != nil {
						val = *dp.Max
					}
				case "min":
					if dp.Min != nil {
						val = *dp.Min
					}
				case "sum":
					if dp.Sum != nil {
						val = *dp.Sum
					}
				case "variance":
					if dp.Variance != nil {
						val = *dp.Variance
					}
				default:
					if dp.Average != nil {
						val = *dp.Average
					}
				}

				fields[fmt.Sprintf("%s_%s", req.metricname, req.filter)] = val

				tm := time.Unix(dp.Timestamp/1000, 0)

				if len(fields) == 0 {
					moduleLogger.Warnf("skip %s.%s datapoint for no fields, %s", req.namespace, req.metricname, dp.String())
					continue
				}

				if ag.isTestOnce() {
					// pass
				} else if ag.isDebug() {
					data, _ := io.MakeMetric(metricSetName, tags, fields, tm)
					fmt.Printf("%s\n", string(data))
				} else {
					io.NamedFeedEx(inputName, datakit.Metric, metricSetName, tags, fields, tm)
				}
			}
		}

		if ag.isTestOnce() {
			break
		}

		datakit.SleepContext(ag.ctx, time.Second*3)
	}

}

func (ag *agent) genAPIClient() {
	auth := basic.NewCredentialsBuilder().
		WithAk(ag.AccessKeyID).
		WithSk(ag.AccessKeySecret).
		WithProjectId(ag.ProjectID).
		Build()

	regid := ag.RegionID
	if regid == "" {
		regid = ag.EndPoint
	}

	var reg *region.Region

	if strings.HasPrefix(regid, "http://") || strings.HasPrefix(regid, "https://") {
		reg = region.NewRegion("", regid)
	} else {
		reg = getCesRegion(regid)
	}

	if reg == nil {
		moduleLogger.Errorf("invalid ces endpoint: %s", regid)
		return
	}

	ag.cesClient = ces.NewCesClient(
		ces.CesClientBuilder().
			WithRegion(reg).
			WithCredential(auth).
			Build())
}

func (ag *agent) genReqsOld(ctx context.Context) error {

	ag.ecsInstanceIDs = ag.listServersDetails()
	moduleLogger.Debugf("%d ecs instances", len(ag.ecsInstanceIDs))

	//生成所有请求
	for _, proj := range ag.Namespace {

		if err := proj.checkProperties(); err != nil {
			return err
		}

		for _, metricName := range proj.MetricNames {

			select {
			case <-ctx.Done():
				return context.Canceled
			default:
			}

			req := proj.genMetricReq(metricName)
			if req.interval == 0 {
				req.interval = ag.Interval.Duration
			}

			adds := proj.applyProperty(req, ag.ecsInstanceIDs)

			ag.reqs = append(ag.reqs, req)
			if len(adds) > 0 {
				ag.reqs = append(ag.reqs, adds...)
			}
		}
	}

	return nil
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
		startTime = nt.Add(-req.interval).Unix() * 1000
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

	numdim := len(req.dimsOld)
	if numdim > 4 {
		moduleLogger.Warnf("only support up to 4 dimensions, get %d dimensions", numdim)
		numdim = 4
	}

	if numdim > 0 {
		request.Dim0 = fmt.Sprintf("%s,%s", req.dimsOld[0].Name, req.dimsOld[0].Value)
	}
	if numdim > 1 {
		dim1 := fmt.Sprintf("%s,%s", req.dimsOld[1].Name, req.dimsOld[1].Value)
		request.Dim1 = &dim1
	}
	if numdim > 2 {
		dim2 := fmt.Sprintf("%s,%s", req.dimsOld[2].Name, req.dimsOld[2].Value)
		request.Dim2 = &dim2
	}
	if numdim > 3 {
		dim3 := fmt.Sprintf("%s,%s", req.dimsOld[3].Name, req.dimsOld[3].Value)
		request.Dim3 = &dim3
	}

	//request.Dim0 = "instance_id,d735ce9a-342e-476e-920d-5243032133d4"
	request.Filter = cesmodel.GetShowMetricDataRequestFilterEnum().AVERAGE
	request.Period = int32(req.period)
	request.From = req.from
	request.To = req.to
	response, err := ag.cesClient.ShowMetricData(request)
	if err != nil {
		moduleLogger.Errorf("fail to get metric: Namespace = %s, MetricName = %s, Period = %v, StartTime = %v(%s), EndTime = %v(%s), Dimensions = %s, err: %s", req.namespace, req.metricname, req.period, req.from, logStarttime, req.to, logEndtime, req.dimsOld, err)
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

	//moduleLogger.Debugf("get %d datapoints: Namespace = %s, MetricName = %s, Filter = %s, Period = %v, InterVal = %v, StartTime = %v(%s), EndTime = %v(%s), Dimensions = %s", nc, req.namespace, req.metricname, req.filter, req.period, req.interval, req.from, logStarttime, req.to, logEndtime, req.dimsOld)

	moduleLogger.Debugf("get %d datapoints: %s", nc, request.String())

	return response, nil
}

func (ag *agent) listServersDetails() []string {

	regid := ag.RegionID
	if regid == "" {
		regid = ag.EndPoint
	}

	reg := getEcsRegion(regid)
	if reg == nil {
		moduleLogger.Warnf("invalid ecs endpoint")
		return nil
	}

	auth := basic.NewCredentialsBuilder().
		WithAk(ag.AccessKeyID).
		WithSk(ag.AccessKeySecret).
		WithProjectId(ag.ProjectID).
		Build()

	cli := ecs.EcsClientBuilder().WithRegion(reg).WithCredential(auth).Build()
	esccli := ecs.NewEcsClient(cli)

	for i := 0; i < 3; i++ {
		req := &ecsmodel.ListServersDetailsRequest{}
		resp, err := esccli.ListServersDetails(req)
		if err != nil {
			moduleLogger.Errorf("%s, request: %s", err, req.String())
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

func getCesRegion(regid string) (reg *region.Region) {
	defer func() {
		if err := recover(); err != nil {
			reg = nil
		}
	}()

	reg = cesregion.ValueOf(regid)
	return
}

func getEcsRegion(regid string) (reg *region.Region) {
	defer func() {
		if err := recover(); err != nil {
			reg = nil
		}
	}()

	reg = ecsregion.ValueOf(regid)
	return
}
