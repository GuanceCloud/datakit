package aliyuncms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers/influx"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials/providers"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/log"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/uploader"
)

func init() {
	config.AddConfig("aliyuncms", &Cfg)
	service.Add("aliyuncms", func(logger log.Logger) service.Service {
		if Cfg.Disable {
			return nil
		}
		return &AliyuncmsSvr{
			logger: logger,
		}
	})
}

var (
	batchInterval = time.Duration(5) * time.Minute
	metricPeriod  = time.Duration(5 * time.Minute)
	rateLimit     = 20
)

type (
	aliyuncmsClient interface {
		DescribeMetricList(request *cms.DescribeMetricListRequest) (*cms.DescribeMetricListResponse, error)
	}

	AliyunCMS struct {
		client aliyuncmsClient

		cfg *CmsCfg

		timer *time.Timer

		wg sync.WaitGroup

		uploader uploader.IUploader
		logger   log.Logger
	}

	AliyuncmsSvr struct {
		cmss   []*AliyunCMS
		logger log.Logger
	}
)

func (m *AliyuncmsSvr) Start(ctx context.Context, up uploader.IUploader) error {

	if Cfg.Disable {
		return nil
	}

	m.cmss = []*AliyunCMS{}

	for _, c := range Cfg.CmsCfg {
		a := &AliyunCMS{
			cfg:      c,
			uploader: up,
			logger:   m.logger,
		}
		m.cmss = append(m.cmss, a)
	}

	var wg sync.WaitGroup

	m.logger.Info("Starting AliyuncmsSvr...")

	for _, c := range m.cmss {
		wg.Add(1)
		go func(ac *AliyunCMS) {
			defer wg.Done()

			if err := ac.Run(ctx); err != nil && err != context.Canceled {
				m.logger.Errorf("%s", err)
			}
		}(c)
	}

	wg.Wait()

	m.logger.Info("AliyuncmsSvr done")
	return nil
}

func (s *AliyunCMS) Run(ctx context.Context) error {

	if err := s.initializeAliyunCMS(); err != nil {
		return err
	}

	s.logger.Info("retrieve aliyun credential success")

	select {
	case <-ctx.Done():
		return context.Canceled
	default:
	}

	lmtr := NewRateLimiter(rateLimit, time.Second)
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
			if err := fetchMetric(req, s.client, s.uploader); err != nil {
				s.logger.Errorf("fetchMetric error: %s", err)
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

func (s *AliyunCMS) initializeAliyunCMS() error {
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

func fetchMetric(req *cms.DescribeMetricListRequest, client aliyuncmsClient, up uploader.IUploader) error {
	tags := make(map[string]string)

	if req.Dimensions != "" {
		ms := []map[string]string{}
		if err := json.Unmarshal([]byte(req.Dimensions), &ms); err == nil {
			for _, m := range ms {
				for k, v := range m {
					tags[k] = v
				}
			}
		}
	}

	tags["regionId"] = req.RegionId

	//const timeInterval = int64(5 * time.Minute * 1000)

	var st, et int64
	//if req.StartTime == "" {
	et = (time.Now().Unix() - 30) * 1000
	st = et - int64(metricPeriod.Seconds())*1000
	//} else {
	//st, _ = strconv.ParseInt(req.EndTime, 10, 64)
	//et = st + int64(metricPeriod.Seconds())*1000
	//}

	req.EndTime = strconv.FormatInt(et, 10)
	req.StartTime = strconv.FormatInt(st, 10)

	//log.Printf("req: %#v", req)

	for more := true; more; {
		resp, err := client.DescribeMetricList(req)
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

		//log.Printf("datapoints count: %v", len(datapoints))

		for _, datapoint := range datapoints {

			fields := make(map[string]interface{})

			if average, ok := datapoint["Average"]; ok {
				fields[formatField(req.MetricName, "Average")] = average
			}
			if minimum, ok := datapoint["Minimum"]; ok {
				fields[formatField(req.MetricName, "Minimum")] = minimum
			}
			if maximum, ok := datapoint["Maximum"]; ok {
				fields[formatField(req.MetricName, "Maximum")] = maximum
			}
			if value, ok := datapoint["Value"]; ok {
				fields[formatField(req.MetricName, "Value")] = value
			}
			tags["userId"] = datapoint["userId"].(string)

			datapointTime := int64(datapoint["timestamp"].(float64)) / 1000

			m, _ := metric.New(formatMeasurement(req.Namespace), tags, fields, time.Unix(datapointTime, 0))

			serializer := influx.NewSerializer()
			output, err := serializer.Serialize(m)
			if err == nil {
				if up != nil {
					up.AddLog(&uploader.LogItem{
						Log: string(output),
					})
				}
			} else {
				//log.Printf("[warn] Serialize to influx protocol line fail: %s", err)
			}
		}

		req.NextToken = resp.NextToken
		more = req.NextToken != ""
	}

	return nil
}

func formatField(metricName string, statistic string) string {
	return fmt.Sprintf("%s_%s", snakeCase(metricName), snakeCase(statistic))
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
