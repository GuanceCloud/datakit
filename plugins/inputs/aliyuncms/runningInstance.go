package aliyuncms

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/time/rate"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials/providers"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

var (
	errGetMetricMeta   = fmt.Errorf("fail to get metric meta")
	errSkipDueInterval = fmt.Errorf("skip this round due to interval")

	retryCount = 5
)

func (s *runningInstance) run(ctx context.Context) error {

	defer func() {
		if err := recover(); err != nil {
			moduleLogger.Errorf("panic, %v", err)
		}
	}()

	for {
		select {
		case <-datakit.Exit.Wait():
			return nil
		default:
		}

		if err := s.initializeAliyunCMS(); err != nil {
			moduleLogger.Errorf("initialize failed, %s", err)
		} else {
			break
		}

		time.Sleep(time.Second)
	}

	//每秒最多20个请求
	limit := rate.Every(50 * time.Millisecond)
	s.limiter = rate.NewLimiter(limit, 1)

	if err := s.genReqs(ctx); err != nil {
		return err
	}

	if len(s.reqs) == 0 {
		moduleLogger.Warnf("no metric found")
		return nil
	}

	select {
	case <-ctx.Done():
		return context.Canceled
	default:
	}

	for {

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		for _, req := range s.reqs {

			select {
			case <-ctx.Done():
				return context.Canceled
			default:
			}

			if err := s.fetchMetric(ctx, req); err != nil {
				if err == errSkipDueInterval {

				}
			}
		}

		datakit.SleepContext(ctx, time.Second*3)
	}
}

func (s *runningInstance) genReqs(ctx context.Context) error {

	//生成所有请求

	for _, proj := range s.cfg.Project {

		/*projMetricMetas, err := s.fetchMetricMeta(ctx, proj.Name, "")
		if err != nil {
			return err
		}

		//暂不支持*, 指标过多
		if len(proj.Metrics.MetricNames) > 0 && proj.Metrics.MetricNames[0] == "*" {
			names := []string{}
			for _, meta := range projMetricMetas {
				names = append(names, meta.metricName)
			}
			proj.Metrics.MetricNames = names
		}*/

		for _, metricName := range proj.Metrics.MetricNames {

			if metricName == "*" {
				continue
			}

			select {
			case <-ctx.Done():
				return context.Canceled
			default:
			}

			req, err := proj.genMetricReq(metricName, s.cfg.RegionID)
			if err != nil {
				moduleLogger.Errorf("%s", err)
				return err
			}

			if req.interval == 0 {
				req.interval = s.cfg.Interval.Duration
			}

			/*if meta, ok := projMetricMetas[metricName]; ok {
				req.meta = meta
			}*/

			s.reqs = append(s.reqs, req)
		}
	}

	return nil
}

func (s *runningInstance) initializeAliyunCMS() error {
	// if s.cfg.RegionID == "" {
	// 	return errors.New("region id is not set")
	// }

	configuration := &providers.Configuration{
		AccessKeyID:     s.cfg.AccessKeyID,
		AccessKeySecret: s.cfg.AccessKeySecret,
	}
	credentialProviders := []providers.Provider{
		providers.NewConfigurationCredentialProvider(configuration),
		providers.NewEnvCredentialProvider(),
		providers.NewInstanceMetadataProvider(),
	}
	credential, err := providers.NewChainProvider(credentialProviders).Retrieve()
	if err != nil {
		return fmt.Errorf("failed to retrieve credential")
	}
	cli, err := cms.NewClientWithOptions(s.cfg.RegionID, sdk.NewConfig(), credential)
	if err != nil {
		return fmt.Errorf("failed to create cms client: %v", err)
	}

	s.cmsClient = cli

	return nil
}

func (s *runningInstance) fetchMetricMeta(ctx context.Context, namespace, metricname string) (map[string]*MetricMeta, error) {

	request := cms.CreateDescribeMetricMetaListRequest()
	request.Scheme = "https"
	request.Namespace = namespace
	request.MetricName = metricname
	request.PageSize = requests.NewInteger(100)

	if s.cfg.SecurityToken != "" {
		//fmt.Printf("token: %s\n", s.cfg.SecurityToken)
		request.QueryParams["SecurityToken"] = s.cfg.SecurityToken
		request.FormParams["SecurityToken"] = s.cfg.SecurityToken

	}

	var err error
	var response *cms.DescribeMetricMetaListResponse

	var tempDelay time.Duration

	for i := 0; i < retryCount; i++ {

		select {
		case <-ctx.Done():
			return nil, context.Canceled
		default:
		}

		s.limiter.Wait(ctx)
		response, err = s.cmsClient.DescribeMetricMetaList(request)

		if tempDelay == 0 {
			tempDelay = time.Millisecond * 50
		} else {
			tempDelay *= 2
		}

		if max := time.Second; tempDelay > max {
			tempDelay = max
		}

		if err == nil && !response.Success {
			err = fmt.Errorf("%s", response.String())
		}

		if err != nil {
			moduleLogger.Warnf("%s", err)
			datakit.SleepContext(ctx, tempDelay)
		} else {
			if i != 0 {
				moduleLogger.Debugf("retry successed, %d", i)
			}
			break
		}
	}

	if err != nil {
		moduleLogger.Errorf("fail to get metric meta for '%s.%s', %s", namespace, metricname, err)
		return nil, errGetMetricMeta
	}

	if len(response.Resources.Resource) == 0 {
		moduleLogger.Warnf("empty metric meta of '%s.%s'", namespace, metricname)
		return nil, errGetMetricMeta
	}

	metas := map[string]*MetricMeta{}

	for _, res := range response.Resources.Resource {
		periodStrs := strings.Split(res.Periods, ",")
		periods := []int64{}
		for _, p := range periodStrs {
			np, err := strconv.ParseInt(p, 10, 64)
			if err == nil {
				periods = append(periods, np)
			} else {
				moduleLogger.Warnf("%s.%s: unknown period '%s', %s", namespace, res.MetricName, p, err)
			}
		}
		meta := &MetricMeta{
			Periods:     periods,
			Statistics:  strings.Split(res.Statistics, ","),
			Dimensions:  strings.Split(res.Dimensions, ","),
			Description: res.Description,
			Unit:        res.Unit,
			metricName:  res.MetricName,
		}
		moduleLogger.Debugf("%s.%s: Periods=%s, Dimensions=%s, Statistics=%s, Unit=%s", namespace, res.MetricName, periodStrs, res.Dimensions, res.Statistics, res.Unit)
		metas[res.MetricName] = meta
	}

	return metas, nil
}

func (s *runningInstance) fetchMetric(ctx context.Context, req *MetricsRequest) error {

	if req.tryGetMeta > 0 && req.meta == nil {
		metas, _ := s.fetchMetricMeta(ctx, req.q.Namespace, req.q.MetricName)
		if len(metas) > 0 {
			req.meta = metas[req.q.MetricName]
		}
		req.tryGetMeta-- //有时接口 DescribeMetricMetaList 更新不及时，所以重试几次后拿不到就忽略
	}

	select {
	case <-ctx.Done():
		return context.Canceled
	default:
	}

	if req.meta != nil {

		//验证period
		if !req.tunePeriod {
			pv, _ := strconv.ParseInt(req.q.Period, 10, 64)
			bValidPeriod := false
			//检查设置的period是否被支持
			for _, n := range req.meta.Periods {
				if pv == n {
					bValidPeriod = true
					break
				}
			}

			if !bValidPeriod {
				moduleLogger.Warnf("period '%v' for %s.%s not support, valid periods: %v", pv, req.q.Namespace, req.q.MetricName, req.meta.Periods)
				req.q.Period = "" //不传，按照监控项默认的最小周期来查询数据
			}

			//只检查一次
			req.tunePeriod = true
		}

		//check dimension
		if !req.tuneDimension {
			if req.q.Dimensions != "" && len(req.meta.Dimensions) > 0 {
				ms := []map[string]string{}
				if err := json.Unmarshal([]byte(req.q.Dimensions), &ms); err == nil {
					btuned := false
					for _, m := range ms {
						for k := range m {
							bSupport := false
							for _, ds := range req.meta.Dimensions {
								if ds == k {
									bSupport = true
									break
								}
							}
							if !bSupport {
								delete(m, k)
								btuned = true
								moduleLogger.Warnf("%s.%s not support dimension '%s'", req.q.Namespace, req.q.MetricName, k)
							}
						}
					}
					if btuned {
						if jd, err := json.Marshal(ms); err == nil {
							moduleLogger.Debugf("dimension after tuned: %s", string(jd))
							req.q.Dimensions = string(jd)
						}
					}
				}
			}
			req.tuneDimension = true
		}

	}

	nt := time.Now().Truncate(time.Minute)
	endTime := nt.Unix() * 1000
	var startTime int64
	if req.lastTime.IsZero() {
		startTime = nt.Add(-(s.cfg.Delay.Duration)).Unix() * 1000
	} else {
		if nt.Sub(req.lastTime) < req.interval {
			return errSkipDueInterval
		}
		startTime = req.lastTime.Add(-(s.cfg.Delay.Duration)).Unix() * 1000
	}

	logEndtime := time.Unix(endTime/int64(1000), 0)
	logStarttime := time.Unix(startTime/int64(1000), 0)

	req.q.EndTime = strconv.FormatInt(endTime, 10)
	req.q.StartTime = strconv.FormatInt(startTime, 10)
	req.q.NextToken = ""

	if s.cfg.SecurityToken != "" {
		//fmt.Printf("token: %s\n", s.cfg.SecurityToken)
		req.q.QueryParams["SecurityToken"] = s.cfg.SecurityToken
		req.q.FormParams["SecurityToken"] = s.cfg.SecurityToken

	}

	datapoints := []map[string]interface{}{}

	for more := true; more; {
		var err error
		var resp *cms.DescribeMetricListResponse
		var tempDelay time.Duration

		for i := 0; i < retryCount; i++ {

			select {
			case <-ctx.Done():
				return context.Canceled
			default:
			}

			s.limiter.Wait(ctx)
			//fmt.Printf("querys: %s", req.q.GetQueryParams())
			resp, err = s.cmsClient.DescribeMetricList(req.q)

			if tempDelay == 0 {
				tempDelay = time.Millisecond * 50
			} else {
				tempDelay *= 2
			}

			if max := time.Second; tempDelay > max {
				tempDelay = max
			}

			if err == nil && !resp.Success {
				err = fmt.Errorf("%s", resp.String())
			}

			if err != nil {
				moduleLogger.Warnf("DescribeMetricList: %s", err)
				time.Sleep(tempDelay)
			} else {
				if i != 0 {
					moduleLogger.Debugf("retry successed, %d", i)
				}
				break
			}
		}

		if err != nil {
			moduleLogger.Errorf("fail to get %s.%s, %s", req.q.Namespace, req.q.MetricName, err)
			return err
		}

		req.q.NextToken = resp.NextToken
		more = (req.q.NextToken != "")

		// if len(resp.Datapoints) == 0 {
		// 	break
		// }

		dps := []map[string]interface{}{}
		if resp.Datapoints != "" {
			if err = json.Unmarshal([]byte(resp.Datapoints), &dps); err != nil {
				moduleLogger.Errorf("%s.%s failed to decode response datapoints:%s, err:%s", req.q.Namespace, req.q.MetricName, resp.Datapoints, err)
			}
		}

		moduleLogger.Debugf("get %v datapoints: Namespace=%s, MetricName=%s, Period=%s, StartTime=%s(%s), EndTime=%s(%s), Dimensions=%s, RegionId=%s, NextToken=%s", len(dps), req.q.Namespace, req.q.MetricName, req.q.Period, req.q.StartTime, logStarttime, req.q.EndTime, logEndtime, req.q.Dimensions, req.q.RegionId, resp.NextToken)

		if err != nil {
			break
		}

		datapoints = append(datapoints, dps...)
	}

	req.lastTime = nt

	metricName := req.q.MetricName

	metricSetName := req.metricSetName
	if metricSetName == "" {
		metricSetName = formatMeasurement(req.q.Namespace)
	}

	for _, datapoint := range datapoints {

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		tags := map[string]string{}

		if req.tags != nil {
			for k, v := range req.tags {
				tags[k] = v
			}
		} else {
			if s.cfg.Tags != nil {
				for k, v := range s.cfg.Tags {
					tags[k] = v
				}
			}
		}

		fields := make(map[string]interface{})

		if average, ok := datapoint["Average"]; ok {
			fields[formatField(metricName, "Average")] = average
		}
		if minimum, ok := datapoint["Minimum"]; ok {
			fields[formatField(metricName, "Minimum")] = minimum
		}
		if maximum, ok := datapoint["Maximum"]; ok {
			fields[formatField(metricName, "Maximum")] = maximum
		}
		if value, ok := datapoint["Value"]; ok {
			fields[formatField(metricName, "Value")] = value
		}
		if value, ok := datapoint["Sum"]; ok {
			fields[formatField(metricName, "Sum")] = value
		}
		if value, ok := datapoint["SumPerMinute"]; ok {
			fields[formatField(metricName, "SumPerMinute")] = value
		}

		for _, k := range supportedDimensions {
			if kv, ok := datapoint[k]; ok {
				if kvstr, bok := kv.(string); bok {
					tags[k] = kvstr
				} else {
					tags[k] = fmt.Sprintf("%v", kv)
				}
			}
		}

		tm := time.Now()
		switch ft := datapoint["timestamp"].(type) {
		case float64:
			tm = time.Unix((int64(ft))/1000, 0)
		}

		if len(fields) == 0 {
			moduleLogger.Warnf("skip %s.%s datapoint for no value, %s", req.q.Namespace, metricName, datapoint)
		}

		if len(fields) > 0 {
			io.NamedFeedEx(inputName, io.Metric, metricSetName, tags, fields, tm)
		}
	}

	return nil
}

func formatField(metricName string, statistic string) string {
	return fmt.Sprintf("%s_%s", metricName, statistic)
}

func formatMeasurement(project string) string {
	project = strings.Replace(project, "/", "_", -1)
	project = snakeCase(project)
	return fmt.Sprintf("aliyuncms_%s", project)
}

func snakeCase(s string) string {
	s = SnakeCase(s)
	s = strings.Replace(s, "__", "_", -1)
	return s
}

func SnakeCase(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}
