package aliyuncms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers/influx"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials/providers"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/uploader"
)

var (
	ErrServerShuttingDown = errors.New("server is shutting down")
	ErrServerRunning      = errors.New("server is running")
	ErrServerCanceled     = errors.New("server is canceled")

	batchInterval = time.Duration(5) * time.Minute
	metricPeriod  = time.Duration(5 * time.Minute)
	rateLimit     = 20
)

type (
	AliyunCMS struct {
		client aliyuncmsClient

		cfg        *CmsCfg
		inShutdown int32
		inRunning  int32

		exitCtx context.Context
		exitFun context.CancelFunc

		timer *time.Timer

		wg sync.WaitGroup

		uploader uploader.IUploader
	}

	aliyuncmsClient interface {
		DescribeMetricList(request *cms.DescribeMetricListRequest) (*cms.DescribeMetricListResponse, error)
	}

	AliyunCMSManager struct {
		cmss []*AliyunCMS

		uploader uploader.IUploader
	}
)

func NewAliyunCMSManager(up uploader.IUploader) *AliyunCMSManager {

	m := &AliyunCMSManager{
		cmss:     []*AliyunCMS{},
		uploader: up,
	}

	for _, c := range Cfg.CmsCfg {
		a := &AliyunCMS{
			cfg:      c,
			uploader: up,
		}
		m.cmss = append(m.cmss, a)
	}

	return m
}

func (m *AliyunCMSManager) Start() error {
	var wg sync.WaitGroup

	for _, c := range m.cmss {
		wg.Add(1)
		go func(ac *AliyunCMS) {
			defer wg.Done()
			if err := ac.Start(); err != nil {
				log.Println(err)
			} else {
				log.Println("cms instance finish")
			}
		}(c)
	}
	wg.Wait()
	log.Println("AliyunCMSManager finish")
	return nil
}

func (m *AliyunCMSManager) Stop() {
	for _, c := range m.cmss {
		c.Stop()
	}
}

func (s *AliyunCMS) isShuttingDown() bool {
	return atomic.LoadInt32(&s.inShutdown) != 0
}

func (s *AliyunCMS) isRunning() bool {
	return atomic.LoadInt32(&s.inRunning) != 0
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

func (s *AliyunCMS) Start() error {
	if s.isShuttingDown() {
		return ErrServerShuttingDown
	}

	if s.isRunning() {
		return ErrServerRunning
	}

	s.exitCtx, s.exitFun = context.WithCancel(context.Background())

	atomic.AddInt32(&s.inRunning, 1)
	defer atomic.AddInt32(&s.inRunning, -1)

	if err := s.initializeAliyunCMS(); err != nil {
		return err
	}

	log.Println("check credential ok")

	select {
	case <-s.exitCtx.Done():
		return ErrServerCanceled
	default:
	}

	lmtr := NewRateLimiter(rateLimit, time.Second)
	defer lmtr.Stop()

	s.wg.Add(1)
	defer s.wg.Done()

	for {

		select {
		case <-s.exitCtx.Done():
			return ErrServerCanceled
		default:
		}

		t := time.Now()
		for _, req := range MetricsRequests {

			select {
			case <-s.exitCtx.Done():
				return ErrServerCanceled
			default:
			}

			<-lmtr.C
			if err := fetchMetric(req, s.client, s.uploader); err != nil {
				log.Printf("fetchMetric error: %s", err)
			}
		}

		useage := time.Now().Sub(t)
		if useage < batchInterval {
			remain := batchInterval - useage
			log.Printf("start wait, useage: %v, towait: %v", useage, remain)

			if s.timer == nil {
				s.timer = time.NewTimer(remain)
			} else {
				s.timer.Reset(remain)
			}
			select {
			case <-s.exitCtx.Done():
				log.Println("cancel done")
				if s.timer != nil {
					s.timer.Stop()
					s.timer = nil
				}
				return ErrServerCanceled
			case <-s.timer.C:
				log.Println("wait done")
			}
		}
	}
}

func (s *AliyunCMS) Stop() {
	if s.isShuttingDown() || !s.isRunning() {
		return
	}
	atomic.AddInt32(&s.inShutdown, 1)
	s.exitFun()
	s.wg.Wait()
	s.client = nil
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
	atomic.AddInt32(&s.inShutdown, -1)
}

func (s *AliyunCMS) Restart() {
	s.Stop()
	s.Start()
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

	log.Printf("req: %#v", req)

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

		log.Printf("datapoints count: %v", len(datapoints))

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
				log.Printf("[warn] Serialize to influx protocol line fail: %s", err)
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
