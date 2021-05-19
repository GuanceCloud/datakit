package huaweiyunces

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	ces "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1"
	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	cesregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/region"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/region"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	iam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	iammodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	iamregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func (ag *agent) genCesClient(projectid, projectname string) (*ces.CesClient, string) {

	var reg *region.Region
	func() {
		defer func() {
			recover()
		}()
		reg = cesregion.ValueOf(projectname)
	}()

	if reg == nil {
		if r, ok := ag.ProjectRegions[projectid]; ok {
			reg = region.NewRegion(r, fmt.Sprintf("https://ces.%s.myhuaweicloud.com", r))
		} else {
			staticRegions := map[string]*region.Region{
				"af-south-1":     cesregion.AF_SOUTH_1,
				"cn-north-4":     cesregion.CN_NORTH_4,
				"cn-north-1":     cesregion.CN_NORTH_1,
				"cn-east-2":      cesregion.CN_EAST_2,
				"cn-east-3":      cesregion.CN_EAST_3,
				"cn-south-1":     cesregion.CN_SOUTH_1,
				"cn-southwest-2": cesregion.CN_SOUTHWEST_2,
				"ap-southeast-2": cesregion.AP_SOUTHEAST_2,
				"ap-southeast-1": cesregion.AP_SOUTHEAST_1,
				"ap-southeast-3": cesregion.AP_SOUTHEAST_3,
			}

			for k, v := range staticRegions {
				if strings.Index(projectname, k) != -1 {
					reg = v
					break
				}
			}
		}

		if reg == nil {
			moduleLogger.Warnf("no endpoint for %s(projectid)", projectname, projectid)
			return nil, ""
		}
	}

	auth := basic.NewCredentialsBuilder().
		WithAk(ag.AccessKeyID).
		WithSk(ag.AccessKeySecret).
		WithProjectId(projectid).
		Build()

	return ces.NewCesClient(
		ces.CesClientBuilder().
			WithRegion(reg).
			WithCredential(auth).
			Build()), reg.Endpoint
}

func (ag *agent) keystoneListAuthProjects() (map[string]iammodel.AuthProjectResult, error) {

	projects := map[string]iammodel.AuthProjectResult{}
	var err error
	var response *iammodel.KeystoneListAuthProjectsResponse
	wt := time.Second * 5
	for i := 0; i < 5; i++ {
		auth := global.NewCredentialsBuilder().
			WithAk(ag.AccessKeyID).
			WithSk(ag.AccessKeySecret).
			Build()

		client := iam.NewIamClient(
			iam.IamClientBuilder().
				WithRegion(iamregion.ValueOf("cn-north-4")).
				WithCredential(auth).
				Build())

		request := &iammodel.KeystoneListAuthProjectsRequest{}
		response, err = client.KeystoneListAuthProjects(request)
		if err == nil {
			if response.Projects != nil {
				for _, p := range *response.Projects {
					//moduleLogger.Debugf("project: %s(%s)", p.Name, p.Id)
					projects[p.Id] = p
				}
			}
			break
		}
		moduleLogger.Errorf("%s", err)
		time.Sleep(wt)
		wt += time.Second * 5
	}

	return projects, err
}

// 指标维度
type MetricsDimension struct {
	// 资源维度，如：弹性云服务器，则维度为instance_id；目前最大支持4个维度，各服务资源的指标维度名称可查看：“[服务指标维度](https://support.huaweicloud.com/usermanual-ces/zh-cn_topic_0202622212.html)”。
	Name *string `json:"name,omitempty"`
	// 资源维度值，为资源的实例ID，如：4270ff17-aba3-4138-89fa-820594c39755。
	Value *string `json:"value,omitempty"`

	projectID string
}

// 查询结果元数据信息，包括分页信息等。
type MetaData struct {
	// 当前返回结果条数。
	Count int32 `json:"count"`
	// 总条数。
	Total int32 `json:"total"`
	// 下一个开始的标记，用于分页。
	Marker string `json:"marker"`
}

type ExtraInfo struct {
	OriginMetricName string `json:"origin_metric_name"`
	MetricPrefix     string `json:"metric_prefix"`
}

// 指标信息
type MetricInfoList struct {
	// 指标维度
	Dimensions []MetricsDimension `json:"dimensions"`
	// 指标名称，必须以字母开头，只能包含0-9/a-z/A-Z/_，长度最短为1，最大为64；各服务的指标名称可查看：“[服务指标名称](https://support.huaweicloud.com/usermanual-ces/zh-cn_topic_0202622212.html)”。
	MetricName string `json:"metric_name"`
	// 指标命名空间，例如弹性云服务器命名空间SYS.ECS；格式为service.item；service和item必须是字符串，必须以字母开头，只能包含0-9/a-z/A-Z/_，总长度最短为3，最大为32。说明： 当alarm_type为（EVENT.SYS| EVENT.CUSTOM）时允许为空；各服务的命名空间可查看：“[服务命名空间](https://support.huaweicloud.com/usermanual-ces/zh-cn_topic_0202622212.html)”。
	Namespace string `json:"namespace"`
	// 指标单位。
	Unit string `json:"unit"`

	ExtraInfo *ExtraInfo `json:"extra_info,omitempty"`
}

// Response Object
type ListMetricsResponse struct {
	// 指标信息列表
	Metrics        *[]MetricInfoList `json:"metrics,omitempty"`
	MetaData       *MetaData         `json:"meta_data,omitempty"`
	HttpStatusCode int               `json:"-"`
}

func GenReqDefForListMetrics() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/V1.0/{project_id}/metrics").
		WithResponse(new(ListMetricsResponse)).
		WithContentType("application/json")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Dim0").
		WithJsonTag("dim.0").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Dim1").
		WithJsonTag("dim.1").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Dim2").
		WithJsonTag("dim.2").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("MetricName").
		WithJsonTag("metric_name").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Namespace").
		WithJsonTag("namespace").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Order").
		WithJsonTag("order").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Start").
		WithJsonTag("start").
		WithLocationType(def.Query))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ContentType").
		WithJsonTag("Content-Type").
		WithLocationType(def.Header))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func (cli *cesCli) listMetrics(ctx context.Context, request *cesmodel.ListMetricsRequest) (*ListMetricsResponse, error) {
	requestDef := GenReqDefForListMetrics()
	var err error
	var resp interface{}
	wt := time.Second * 5
	for i := 0; i < 5; i++ {

		select {
		case <-ctx.Done():
			return nil, nil
		default:
		}

		resp, err = cli.cli.HcClient.Sync(request, requestDef)
		if err == nil {
			return resp.(*ListMetricsResponse), nil
		}
		datakit.SleepContext(ctx, wt)
		wt += time.Second * 5
	}
	return nil, err
}

func (cli *cesCli) listNamespaceMetrics(namespace string, ag *agent) (metrics []MetricInfoList, err error) {

	var response *ListMetricsResponse

	metrics = []MetricInfoList{}

	start := ""
	var limit int32 = 1000
	more := true

	for more {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		request := &cesmodel.ListMetricsRequest{}
		if start != "" {
			request.Start = &start
		}
		if namespace != "" {
			request.Namespace = &namespace
		}
		request.Limit = &limit

		ag.limiter.Wait(ag.ctx)
		response, err = cli.listMetrics(ag.ctx, request)

		if err != nil {
			moduleLogger.Errorf("listMetrics error:%s, namespace=%s", err, namespace)
			return
		}

		if response != nil {
			if response.Metrics != nil {
				metrics = append(metrics, *response.Metrics...)
			}
			if response.MetaData != nil {
				start = response.MetaData.Marker
			}
		}
		if start == "" {
			more = false
		}
	}
	return
}

func (cli *cesCli) getMetricInfos(ag *agent) ([]MetricInfoList, error) {

	if len(ag.includeMetrics) > 0 {
		var result []MetricInfoList
		for namespace := range ag.includeMetrics {

			select {
			case <-ag.ctx.Done():
				return nil, nil
			default:
			}

			list, err := cli.listNamespaceMetrics(namespace, ag)
			if err != nil {
				return nil, err
			}
			if list != nil {
				result = append(result, list...)
			}
		}
		return result, nil
	}
	return cli.listNamespaceMetrics("", ag)
}

func (cli *cesCli) batchListMetricData(ag *agent, req *metricsRequest) ([]cesmodel.BatchMetricData, error) {

	if len(req.dimensoions) == 0 {
		err := fmt.Errorf("no dimension found, namespace=%s, metricname=%s", req.namespace, req.metricname)
		moduleLogger.Errorf("%s", err)
		return nil, err
	}

	var results []cesmodel.BatchMetricData

	nt := time.Now().Truncate(time.Second)
	endTime := nt.Unix() * 1000
	var startTime int64
	if req.lastTime.IsZero() {
		startTime = nt.Add(-req.interval).Unix() * 1000
	} else {
		if nt.Sub(req.lastTime) < req.interval {
			return nil, nil
		}
		startTime = req.lastTime.Add(-(req.delay)).Unix() * 1000
	}

	req.to = endTime
	req.from = startTime

	var metricInfos []cesmodel.MetricInfo

	//moduleLogger.Debugf("%d dimensions", len(req.dimensoions))

	for _, d := range req.dimensoions {
		mi := cesmodel.MetricInfo{
			Namespace:  req.namespace,
			MetricName: req.metricname,
		}
		mi.Dimensions = []cesmodel.MetricsDimension{
			cesmodel.MetricsDimension{
				Name:  d.Name,
				Value: d.Value,
			},
		}
		metricInfos = append(metricInfos, mi)
	}

	index := 0
	//批量接口一次最多支持10个
	for {

		last := index + 10
		if last > len(metricInfos) {
			last = len(metricInfos)
		}

		request := &cesmodel.BatchListMetricDataRequest{}
		request.Body = &cesmodel.BatchListMetricDataRequestBody{
			Metrics: metricInfos[index:last],
			Period:  fmt.Sprintf("%d", req.period),
			Filter:  req.filter,
			From:    req.from,
			To:      req.to,
		}

		var response *cesmodel.BatchListMetricDataResponse
		var err error
		wt := time.Second * 5
		for i := 0; i < 5; i++ {
			ag.limiter.Wait(ag.ctx)
			response, err = cli.cli.BatchListMetricData(request)
			if err == nil {
				break
			}
			time.Sleep(wt)
			wt += time.Second * 5
		}

		if err != nil {
			moduleLogger.Errorf("fail to batchListMetricData, request:%s, err: %s", request.String(), err)
			if response != nil {
				moduleLogger.Errorf("%s", response.String())
			}
			return nil, err
		}

		if response.Metrics != nil {
			cnt := 0
			for _, m := range *response.Metrics {
				cnt += len(m.Datapoints)
			}
			results = append(results, *response.Metrics...)
			moduleLogger.Debugf("%s(%s.%s) get %d datapoints", cli.proj.Name, req.namespace, req.metricname, cnt)
			if cnt == 0 {
				moduleLogger.Debugf("request: %s", request.String())
			}
		}

		index += 10

		if index >= len(metricInfos) {
			break
		}
	}

	req.lastTime = nt

	return results, nil
}
