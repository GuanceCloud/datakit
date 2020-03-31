package aliyuncms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/time/rate"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials/providers"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
)

type requestRateTrace struct {
	lastRequestTime      time.Time
	lastRequestTimeCount int64
	requestCount         int64
}

func (t *requestRateTrace) check(logger *models.Logger) {
	nowtm := time.Now()
	if nowtm.Sub(t.lastRequestTime) >= time.Second {
		logger.Debugf("request rate, %v requests in %v", t.requestCount-t.lastRequestTimeCount, nowtm.Sub(t.lastRequestTime))
		t.lastRequestTime = nowtm
		t.lastRequestTimeCount = t.requestCount
	}
}

var (
	errGetMetricMeta = fmt.Errorf("fail to get metric meta")

	reqRateTrace requestRateTrace

	retryCount = 5
)

func (s *runningCMS) run(ctx context.Context) error {

	defer func() {
		if err := recover(); err != nil {
			//s.agent.inputStat.SetLastErr(fmt.Errorf("%v", err))
		}
		atomic.AddInt32(&s.agent.inputStat.Stat, -1)
	}()

	if err := s.initializeAliyunCMS(); err != nil {
		s.logger.Errorf("create cms client failed, %s", err)
		return err
	}

	if err := s.genReqs(); err != nil {
		return err
	}

	if len(s.reqs) == 0 {
		s.logger.Warnf("no metric is configed for %s.%s", s.cfg.AccessKeyID, s.cfg.RegionID)
		return nil
	}

	select {
	case <-ctx.Done():
		return context.Canceled
	default:
	}

	limit := rate.Every(50 * time.Millisecond)
	s.limiter = rate.NewLimiter(limit, 1)

	for {

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		tm := time.Now()

		metricOK := 0
		metricFail := 0
		for _, req := range s.reqs {

			select {
			case <-ctx.Done():
				return context.Canceled
			default:
			}

			if err := s.fetchMetric(ctx, req); err != nil {
				metricFail++
				//s.agent.inputStat.SetLastErr(err)
				//s.logger.Errorf(`fail to get metric "%s.%s", %s`, req.q.Namespace, req.q.MetricName, err)
			} else {
				metricOK++
			}
		}

		useage := time.Now().Sub(tm)
		s.logger.Debugf(`use %v in this loop, success=%d, fail=%d`, useage, metricOK, metricFail)
		if useage < batchInterval {
			remain := batchInterval - useage
			internal.SleepContext(ctx, remain)
		}
	}
}

func (s *runningCMS) genReqs() error {

	//生成所有请求
	for _, proj := range s.cfg.Project {
		for _, metricName := range proj.Metrics.MetricNames {

			req := cms.CreateDescribeMetricListRequest()
			req.Scheme = "https"
			req.RegionId = s.cfg.RegionID
			req.Period = proj.getPeriod(metricName)
			req.MetricName = metricName
			req.Namespace = proj.Name
			if ds, err := proj.genDimension(metricName, s.logger); err == nil {
				req.Dimensions = ds
			}

			s.reqs = append(s.reqs, &MetricsRequest{
				q: req,
			})
		}
	}

	return nil
}

func (s *runningCMS) initializeAliyunCMS() error {
	if s.cfg.RegionID == "" {
		return errors.New("region id is not set")
	}

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

func (s *runningCMS) fetchMetricMeta(ctx context.Context, namespace, metricname string) (*MetricMeta, error) {

	request := cms.CreateDescribeMetricMetaListRequest()
	request.Scheme = "https"

	request.Namespace = namespace
	request.MetricName = metricname

	var err error
	var response *cms.DescribeMetricMetaListResponse

	s.limiter.Wait(ctx)
	var tempDelay time.Duration
	if reqRateTrace.lastRequestTime.IsZero() {
		reqRateTrace.lastRequestTime = time.Now()
	}

	reqUid, _ := uuid.NewV4()

	for i := 0; i < retryCount; i++ {

		response, err = s.cmsClient.DescribeMetricMetaList(request)
		reqRateTrace.requestCount++
		reqRateTrace.check(s.logger)

		if tempDelay == 0 {
			tempDelay = time.Millisecond * 50
		} else {
			tempDelay *= 2
		}

		if max := time.Second; tempDelay > max {
			tempDelay = max
		}

		if err == nil && response.Code != "200" {
			err = fmt.Errorf("%s", response.String())
		}

		if err != nil {
			s.logger.Warnf("fail to get metric meta for '%s.%s' (%s), %s", namespace, metricname, reqUid.String(), err)
			time.Sleep(tempDelay)
		} else {
			if i != 0 {
				s.logger.Debugf("retry %s successed, %d", reqUid.String(), i)
			}
			break
		}
	}

	if err != nil {
		s.agent.faildRequest++

		ectx := errCtxMetricMeta(request)
		ectx["RequestId"] = response.RequestId
		ctxStr, _ := json.Marshal(ectx)

		e := internal.ContextErr{
			ID:      reqUid.String(),
			Context: string(ctxStr),
			Content: err.Error(),
		}
		s.logger.Errorf("%s", e.Error())

		s.agent.inputStat.AddErrorID(reqUid.String())
		return nil, errGetMetricMeta
	} else {
		s.agent.succedRequest++
	}

	if len(response.Resources.Resource) == 0 {
		s.logger.Warnf("empty metric meta of '%s.%s'", namespace, metricname)
		return nil, errGetMetricMeta
	}

	for _, res := range response.Resources.Resource {
		periodStrs := strings.Split(res.Periods, ",")
		periods := []int64{}
		for _, p := range periodStrs {
			np, err := strconv.ParseInt(p, 10, 64)
			if err == nil {
				periods = append(periods, np)
			} else {
				s.logger.Warnf("unknown period '%s', %s", p, err)
			}
		}
		meta := &MetricMeta{
			Periods:     periods,
			Statistics:  strings.Split(res.Statistics, ","),
			Dimensions:  strings.Split(res.Dimensions, ","),
			Description: res.Description,
			Unit:        res.Unit,
		}
		//s.logger.Debugf("%s.%s: Periods=%s, Dimensions=%s, Statistics=%s, Unit=%s", namespace, metricname, periodStrs, res.Dimensions, res.Statistics, res.Unit)
		return meta, nil
	}

	return nil, nil
}

func (s *runningCMS) fetchMetric(ctx context.Context, req *MetricsRequest) error {

	if !req.haveGetMeta && req.meta == nil {
		req.meta, _ = s.fetchMetricMeta(ctx, req.q.Namespace, req.q.MetricName)
		req.haveGetMeta = true //有些指标阿里云更新不及时，所以拿不到就忽略
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
				s.logger.Warnf("period '%v' for %s.%s not support, valid periods: %v", pv, req.q.Namespace, req.q.MetricName, req.meta.Periods)
				req.q.Period = "" //按照监控项默认的最小周期来查询数据
			}

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
								s.logger.Warnf("%s.%s not support dimension '%s'", req.q.Namespace, req.q.MetricName, k)
							}
						}
					}
					if btuned {
						if jd, err := json.Marshal(ms); err == nil {
							req.q.Dimensions = string(jd)
						}
					}
				}
			}
			req.tuneDimension = true
		}

	}

	nt := time.Now().Truncate(time.Minute) //.Add(-time.Second * 30)
	endTime := nt.Unix() * 1000
	var startTime int64
	//if req.lastTime.IsZero() {
	startTime = nt.Add(-(10 * time.Minute)).Unix() * 1000
	//} else {
	//	startTime = req.lastTime.Unix() * 1000
	//}

	logEndtime := time.Unix(endTime/int64(1000), 0)
	logStarttime := time.Unix(startTime/int64(1000), 0)

	req.q.EndTime = strconv.FormatInt(endTime, 10)
	req.q.StartTime = strconv.FormatInt(startTime, 10)

	req.q.NextToken = ""

	for more := true; more; {
		var err error
		var resp *cms.DescribeMetricListResponse
		s.limiter.Wait(ctx)
		var tempDelay time.Duration
		if reqRateTrace.lastRequestTime.IsZero() {
			reqRateTrace.lastRequestTime = time.Now()
		}

		reqUid, _ := uuid.NewV4()

		for i := 0; i < retryCount; i++ {
			resp, err = s.cmsClient.DescribeMetricList(req.q)
			reqRateTrace.requestCount++
			reqRateTrace.check(s.logger)

			if tempDelay == 0 {
				tempDelay = time.Millisecond * 50
			} else {
				tempDelay *= 2
			}

			if max := time.Second; tempDelay > max {
				tempDelay = max
			}

			if err == nil && resp.Code != "200" {
				err = fmt.Errorf("%s", resp.String())
			}

			if err != nil {
				s.logger.Errorf("fail to query metric list (%s): %s", reqUid.String(), err)
				time.Sleep(tempDelay)
			} else {
				if i != 0 {
					s.logger.Debugf("retry %s successed, %d", reqUid.String(), i)
				}
				break
			}
		}

		if err != nil {

			ectx := errCtxMetricList(req)
			ectx["RequestId"] = resp.RequestId
			ctxStr, _ := json.Marshal(ectx)

			e := internal.ContextErr{
				ID:      reqUid.String(),
				Context: string(ctxStr),
				Content: err.Error(),
			}
			s.logger.Errorf("%s", e.Error())

			s.agent.faildRequest++
			s.agent.inputStat.AddErrorID(reqUid.String())
			return err
		} else {
			s.agent.succedRequest++
		}

		if len(resp.Datapoints) == 0 {
			break
		}

		var datapoints []map[string]interface{}
		if err = json.Unmarshal([]byte(resp.Datapoints), &datapoints); err != nil {
			return fmt.Errorf("failed to decode response datapoints: %v", err)
		}

		s.logger.Debugf("get %v datapoints: Namespace=%s, MetricName=%s, Period=%s, StartTime=%s(%s), EndTime=%s(%s), Dimensions=%s, RegionId=%s, NextToken=%s", len(datapoints), req.q.Namespace, req.q.MetricName, req.q.Period, req.q.StartTime, logStarttime, req.q.EndTime, logEndtime, req.q.Dimensions, req.q.RegionId, resp.NextToken)

		metricName := req.q.MetricName

		for _, datapoint := range datapoints {

			tags := map[string]string{
				"regionId": req.q.RegionId,
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

			for _, k := range dms {
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
				s.logger.Debugf("skip %s.%s datapoint for no value, %s", req.q.Namespace, metricName, datapoint)
			}

			if s.agent.accumulator != nil && len(fields) > 0 {
				s.agent.inputStat.IncrTotal()
				s.agent.accumulator.AddFields(formatMeasurement(req.q.Namespace), fields, tags, tm)
			}
		}

		req.q.NextToken = resp.NextToken
		more = (req.q.NextToken != "")
	}

	//req.lastTime = nt.Add(-1 * time.Minute)

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
