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
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/limiter"
)

func (s *RunningCMS) run() error {

	if err := s.initializeAliyunCMS(); err != nil {
		return err
	}

	s.logger.Infof("retrieve aliyun credential success")

	select {
	case <-s.ctx.Done():
		return context.Canceled
	default:
	}

	lmtr := limiter.NewRateLimiter(rateLimit, time.Second)
	defer lmtr.Stop()

	s.wg.Add(1)
	defer s.wg.Done()

	for {

		select {
		case <-s.ctx.Done():
			return context.Canceled
		default:
		}

		t := time.Now()
		for _, req := range MetricsRequests {

			select {
			case <-s.ctx.Done():
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
			case <-s.ctx.Done():
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

func (s *RunningCMS) initializeAliyunCMS() error {
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
		return errors.New("failed to retrieve credential")
	}
	cli, err := cms.NewClientWithOptions(s.cfg.RegionID, sdk.NewConfig(), credential)
	if err != nil {
		return fmt.Errorf("failed to create cms client: %v", err)
	}

	s.client = cli

	return nil
}

func (s *RunningCMS) fetchMetricInfo(namespace, metricname string) (*MetricInfo, error) {

	request := cms.CreateDescribeMetricMetaListRequest()
	request.Scheme = "https"

	request.Namespace = namespace
	request.MetricName = metricname
	request.PageSize = requests.NewInteger(100)

	response, err := s.client.DescribeMetricMetaList(request)
	if err != nil {
		s.logger.Warnf("fail to get metric(%s.%s) info: %s", namespace, metricname, err)
		return nil, fmt.Errorf("fail to get metric(%s.%s) info: %s", namespace, metricname, err)
	}

	if len(response.Resources.Resource) == 0 {
		return nil, fmt.Errorf("metric \"%s\" not support in %s", metricname, namespace)
	}

	for _, res := range response.Resources.Resource {
		periodStrs := strings.Split(res.Periods, ",")
		periods := []int64{}
		for _, p := range periodStrs {
			np, err := strconv.ParseInt(p, 10, 64)
			if err == nil {
				periods = append(periods, np)
			}
		}
		info := &MetricInfo{
			Periods:    periods,
			Statistics: strings.Split(res.Statistics, ","),
			Dimensions: res.Dimensions,
		}
		s.logger.Debugf("%s.%s: Periods=%s, Statistics=%s", namespace, metricname, periodStrs, strings.Split(res.Statistics, ","))
		return info, nil
	}

	return nil, nil
}

func (s *RunningCMS) fetchMetric(req *MetricsRequest) error {

	var err error
	if !req.checkPeriod {
		if req.info == nil {
			req.info, err = s.fetchMetricInfo(req.q.Namespace, req.q.MetricName)
			if err != nil {
				req.info = nil
				req.checkPeriod = true
			}
		}
		if req.info != nil {
			pv, _ := strconv.ParseInt(req.q.Period, 10, 64)
			bok := false
			for _, n := range req.info.Periods {
				if pv == n {
					bok = true
					break
				}
			}

			if !bok {
				s.logger.Warnf("period of %v for %s.%s not support, avariable periods:%#v", pv, req.q.Namespace, req.q.MetricName, req.info.Periods)
				//按照监控项默认的最小周期来查询数据
				req.q.Period = ""
			}

			req.checkPeriod = true
		}
	}

	// if req.q.Dimensions != "" {
	// 	ms := []map[string]string{}
	// 	if err := json.Unmarshal([]byte(req.q.Dimensions), &ms); err == nil {
	// 		for _, m := range ms {
	// 			for k, v := range m {
	// 				tags[k] = v
	// 			}
	// 		}
	// 	} else {
	// 		s.logger.Errorf("dimesion err: %s", err)
	// 	}
	// }

	nt := time.Now().Truncate(time.Minute)
	et := nt.Unix() * 1000
	st := nt.Add(-(5 * time.Minute)).Unix() * 1000 //-6因为是[)

	req.q.EndTime = strconv.FormatInt(et, 10)
	req.q.StartTime = strconv.FormatInt(st, 10)

	//req.q.EndTime = `1574918880000`   // strconv.FormatInt(et, 10)
	//req.q.StartTime = `1574918580000` //strconv.FormatInt(st, 10)
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

			tags := make(map[string]string)

			if s.aliCMS.tags != nil {
				for k, v := range s.aliCMS.tags {
					tags[k] = v
				}
			}
			tags["regionId"] = req.q.RegionId

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

			if s.aliCMS.accumulator != nil && len(fields) > 0 {
				s.aliCMS.accumulator.AddFields(formatMeasurement(req.q.Namespace), fields, tags)
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
