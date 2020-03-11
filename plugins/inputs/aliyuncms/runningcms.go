package aliyuncms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials/providers"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/limiter"
)

var (
	errGetMetricMeta = fmt.Errorf("fail to get metric meta")
)

func (s *runningCMS) run(ctx context.Context) error {

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		if err := s.initializeAliyunCMS(); err != nil {
			s.logger.Errorf("initialize error, %s", err)
			internal.SleepContext(ctx, time.Second*15)
		} else {
			break
		}
	}

	select {
	case <-ctx.Done():
		return context.Canceled
	default:
	}

	lmtr := limiter.NewRateLimiter(rateLimit, time.Second)
	defer lmtr.Stop()

	s.wg.Add(1)
	defer s.wg.Done()

	for {

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		t := time.Now()
		for _, req := range MetricsRequests {

			select {
			case <-ctx.Done():
				return context.Canceled
			default:
			}

			<-lmtr.C
			if err := s.fetchMetric(req); err != nil {
				s.logger.Errorf(`fail to get metric "%s.%s", %s`, req.q.Namespace, req.q.MetricName, err)
			}
		}

		useage := time.Now().Sub(t)
		if useage < batchInterval {
			remain := batchInterval - useage

			if s.timer == nil {
				s.timer = time.NewTimer(remain)
			} else {
				s.timer.Reset(remain)
			}
			select {
			case <-ctx.Done():
				if s.timer != nil {
					s.timer.Stop()
					s.timer = nil
				}
				return context.Canceled
			case <-s.timer.C:
			}
		}
	}
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

	s.client = cli

	return nil
}

func (s *runningCMS) fetchMetricMeta(namespace, metricname string) (*MetricMeta, error) {

	request := cms.CreateDescribeMetricMetaListRequest()
	request.Scheme = "https"

	request.Namespace = namespace
	request.MetricName = metricname

	response, err := s.client.DescribeMetricMetaList(request)
	if err != nil {
		s.logger.Warnf("fail to get metric meta for '%s.%s', %s", namespace, metricname, err)
		return nil, errGetMetricMeta
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
		s.logger.Debugf("%s.%s: Periods=%s, Dimensions=%s, Statistics=%s, Unit=%s", namespace, metricname, periodStrs, res.Dimensions, res.Statistics, res.Unit)
		return meta, nil
	}

	return nil, nil
}

func (s *runningCMS) fetchMetric(req *MetricsRequest) error {

	if !req.haveGetMeta {
		if req.meta == nil {
			req.meta, _ = s.fetchMetricMeta(req.q.Namespace, req.q.MetricName)
		}
		req.haveGetMeta = true //有些指标阿里云更新不及时，所以拿不到就忽略
	}

	if req.meta != nil {

		//验证period
		if !req.tunePeriod {
			pv, _ := strconv.ParseInt(req.q.Period, 10, 64)
			bValidPeriod := false
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

	nt := time.Now().Truncate(time.Minute)
	et := nt.Unix() * 1000
	st := nt.Add(-(5 * time.Minute)).Unix() * 1000

	req.q.EndTime = strconv.FormatInt(et, 10)
	req.q.StartTime = strconv.FormatInt(st, 10)

	req.q.NextToken = ""

	s.logger.Debugf("request: Namespace:%s, MetricName:%s, Period:%s, StartTime:%s, EndTime:%s, Dimensions:%s", req.q.Namespace, req.q.MetricName, req.q.Period, req.q.StartTime, req.q.EndTime, req.q.Dimensions)

	for more := true; more; {
		resp, err := s.client.DescribeMetricList(req.q)
		if err != nil {
			return fmt.Errorf("failed to query metric list: %v", err)
		} else if resp.Code != "200" {
			return fmt.Errorf("failed to query metric list: %v", resp.Message)
		}

		if len(resp.Datapoints) == 0 {
			break
		}

		var datapoints []map[string]interface{}
		if err = json.Unmarshal([]byte(resp.Datapoints), &datapoints); err != nil {
			return fmt.Errorf("failed to decode response datapoints: %v", err)
		}

		for _, datapoint := range datapoints {

			tags := map[string]string{
				"regionId": req.q.RegionId,
			}

			fields := make(map[string]interface{})

			if average, ok := datapoint["Average"]; ok {
				fields[formatField(req.q.MetricName, "Average")] = average
			}
			if minimum, ok := datapoint["Minimum"]; ok {
				fields[formatField(req.q.MetricName, "Minimum")] = minimum
			}
			if maximum, ok := datapoint["Maximum"]; ok {
				fields[formatField(req.q.MetricName, "Maximum")] = maximum
			}
			if value, ok := datapoint["Value"]; ok {
				fields[formatField(req.q.MetricName, "Value")] = value
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

			if s.agent.accumulator != nil && len(fields) > 0 {
				s.agent.accumulator.AddFields(formatMeasurement(req.q.Namespace), fields, tags, tm)
			}

		}

		req.q.NextToken = resp.NextToken
		more = (req.q.NextToken != "")
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
