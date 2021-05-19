package aliyuncms

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/time/rate"

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

func (s *CMS) run(ctx context.Context) {

	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			moduleLogger.Errorf("panic: %s", err)
			moduleLogger.Errorf("%s", string(buf[:n]))
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			moduleLogger.Infof("aliyuncloud api call info: %s", s.apiCallInfo)

			datakit.SleepContext(ctx, time.Hour)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
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
		return
	}

	if len(s.reqs) == 0 {
		moduleLogger.Warnf("no metric found")
		return
	}

	select {
	case <-ctx.Done():
		return
	default:
	}

	for {

		select {
		case <-ctx.Done():
			return
		default:
		}

		for _, req := range s.reqs {

			select {
			case <-ctx.Done():
				return
			default:
			}

			if err := s.fetchMetric(ctx, req); err != nil {
				if err != errSkipDueInterval && err != context.Canceled {
					moduleLogger.Warnf("%s", err)
				}
			}
		}

		datakit.SleepContext(ctx, time.Second*3)
	}

}

//构造所有请求
func (s *CMS) genReqs(ctx context.Context) error {

	for _, proj := range s.Project {

		if err := proj.checkProperties(); err != nil {
			return err
		}

		var reqs []*MetricsRequest

		var metrcNames []string
		if proj.MetricNames != "" {
			parts := strings.Split(proj.MetricNames, ",")
			for _, p := range parts {
				metrcNames = append(metrcNames, strings.TrimSpace(p))
			}
		} else {
			if proj.Metrics != nil {
				metrcNames = proj.Metrics.MetricNames
			}
		}

		if len(metrcNames) == 0 {
			metas, err := s.describeMetricMetaList(ctx, proj.namespace(), "")
			if err != nil {
				moduleLogger.Errorf("fail to DescribeMetricMetaList, %s", err)
				return err
			}
			moduleLogger.Debugf("get %d metrics of %s", len(metas), proj.namespace())

			for name, meta := range metas {
				select {
				case <-ctx.Done():
					return context.Canceled
				default:
				}

				r := proj.makeReqWrap(name)
				r.meta = meta
				reqs = append(reqs, r)
			}
		} else {
			for _, metricName := range metrcNames {

				select {
				case <-ctx.Done():
					return context.Canceled
				default:
				}

				r := proj.makeReqWrap(metricName)
				reqs = append(reqs, r)
			}
		}

		for _, req := range reqs {

			select {
			case <-ctx.Done():
				return context.Canceled
			default:
			}

			proj.applyProperty(req)

			if req.interval == 0 {
				req.interval = s.Interval.Duration
			}
		}

		s.reqs = append(s.reqs, reqs...)
	}

	moduleLogger.Infof("total metrics: %d", len(s.reqs))

	return nil
}

func (s *CMS) initializeAliyunCMS() error {

	cli, err := cms.NewClientWithAccessKey(s.RegionID, s.AccessKeyID, s.AccessKeySecret)
	if err != nil {
		return err
	}
	s.apiClient = cli

	return nil
}

func (s *CMS) describeMetricMetaList(ctx context.Context, namespace, metricname string) (map[string]*MetricMeta, error) {

	request := cms.CreateDescribeMetricMetaListRequest()
	request.Scheme = "https"
	request.Namespace = namespace
	request.MetricName = metricname
	request.PageSize = requests.NewInteger(1000)

	if s.SecurityToken != "" {
		//fmt.Printf("token: %s\n", s.cfg.SecurityToken)
		request.QueryParams["SecurityToken"] = s.SecurityToken
		request.FormParams["SecurityToken"] = s.SecurityToken

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
		response, err = s.apiClient.DescribeMetricMetaList(request)

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
			s.apiCallInfo.Inc(`DescribeMetricMetaList`, true)
		} else {
			s.apiCallInfo.Inc(`DescribeMetricMetaList`, false)
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
		//moduleLogger.Debugf("%s.%s: Periods=%s, Dimensions=%s, Statistics=%s, Unit=%s", namespace, res.MetricName, periodStrs, res.Dimensions, res.Statistics, res.Unit)
		metas[res.MetricName] = meta
	}

	return metas, nil
}

func (s *CMS) fetchMetric(ctx context.Context, req *MetricsRequest) error {

	if req.tryGetMeta > 0 && req.meta == nil {
		metas, _ := s.describeMetricMetaList(ctx, req.q.Namespace, req.q.MetricName)
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
		startTime = nt.Add(-(s.Delay.Duration)).Unix() * 1000
	} else {
		if nt.Sub(req.lastTime) < req.interval {
			return errSkipDueInterval
		}
		startTime = req.lastTime.Add(-(s.Delay.Duration)).Unix() * 1000
	}

	logEndtime := time.Unix(endTime/int64(1000), 0)
	logStarttime := time.Unix(startTime/int64(1000), 0)

	req.q.EndTime = strconv.FormatInt(endTime, 10)
	req.q.StartTime = strconv.FormatInt(startTime, 10)
	req.q.NextToken = ""

	if s.SecurityToken != "" {
		//fmt.Printf("token: %s\n", s.cfg.SecurityToken)
		req.q.QueryParams["SecurityToken"] = s.SecurityToken
		req.q.FormParams["SecurityToken"] = s.SecurityToken

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

			resp, err = s.apiClient.DescribeMetricList(req.q)

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
				//moduleLogger.Warnf("DescribeMetricList: %s", err)
				s.apiCallInfo.Inc("DescribeMetricList", true)
				time.Sleep(tempDelay)
			} else {
				if i != 0 {
					moduleLogger.Debugf("retry successed, %d", i)
				}
				s.apiCallInfo.Inc("DescribeMetricList", false)
				break
			}
		}

		if err != nil {
			moduleLogger.Debugf("params: Namespace: %s, MetricName: %s, Period: %s, StartTime: %s, EndTime: %s, Dimensions: %s, RegionId: %s, NextToken: %s", req.q.Namespace, req.q.MetricName, req.q.Period, req.q.StartTime, req.q.EndTime, req.q.Dimensions, req.q.RegionId, resp.NextToken)
			moduleLogger.Errorf("bad response, err: %s", err)
			break
		}

		req.q.NextToken = resp.NextToken
		more = (req.q.NextToken != "")

		dps := []map[string]interface{}{}
		if resp.Datapoints != "" {
			if err = json.Unmarshal([]byte(resp.Datapoints), &dps); err != nil {
				moduleLogger.Errorf("%s.%s failed to decode response datapoints:%s, err:%s", req.q.Namespace, req.q.MetricName, resp.Datapoints, err)
			}
		}

		moduleLogger.Debugf("get %v datapoints: Namespace = %s, MetricName = %s, Period = %s, StartTime = %s(%s), EndTime = %s(%s), Dimensions = %s, RegionId = %s, NextToken = %s", len(dps), req.q.Namespace, req.q.MetricName, req.q.Period, req.q.StartTime, logStarttime, req.q.EndTime, logEndtime, req.q.Dimensions, req.q.RegionId, resp.NextToken)

		if err != nil {
			break
		}

		datapoints = append(datapoints, dps...)
	}

	req.lastTime = nt

	metricSetName := req.measurementName
	if metricSetName == "" {
		metricSetName = formatMeasurement(req.q.Namespace)
	}

	metricName := req.q.MetricName

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
			if s.Tags != nil {
				for k, v := range s.Tags {
					tags[k] = v
				}
			}
		}

		fields := make(map[string]interface{})

		for _, k := range valueKeys {
			v := datapoint[k]
			if v == nil {
				continue
			}
			if sv, ok := v.(string); ok && sv == "" {
				fields[formatField(metricName, k)] = float64(0)
			} else {
				fields[formatField(metricName, k)] = v
			}
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

		if len(fields) > 0 {

			if s.mode == "debug" {
				//data, _ := io.MakeMetric(metricSetName, tags, fields, tm)
				//fmt.Printf("%s\n", string(data))
			} else {
				io.HighFreqFeedEx(inputName, datakit.Metric, metricSetName, tags, fields, tm)
			}
		} else {
			moduleLogger.Warnf("skip %s.%s datapoint for no value, %s", req.q.Namespace, metricName, datapoint)
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

func snakeCase(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	s := strings.Replace(string(out), "__", "_", -1)

	return s
}
