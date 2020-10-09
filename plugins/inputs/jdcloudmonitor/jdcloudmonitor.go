package jdcloudmonitor

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"golang.org/x/time/rate"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	jdcore "github.com/jdcloud-api/jdcloud-sdk-go/core"
	jcwapis "github.com/jdcloud-api/jdcloud-sdk-go/services/monitor/apis"
	jcwclient "github.com/jdcloud-api/jdcloud-sdk-go/services/monitor/client"
)

var (
	inputName    = `jdcloudmonitor`
	moduleLogger *logger.Logger
)

func (_ *agent) Catalog() string {
	return `jdcloud`
}

func (_ *agent) SampleConfig() string {
	return sampleConfig
}

func (ag *agent) Run() {
	moduleLogger = logger.SLogger(inputName)

	go func() {
		<-datakit.Exit.Wait()
		ag.cancelFun()
	}()

	if ag.Delay.Duration == 0 {
		ag.Delay.Duration = time.Minute * 1
	}

	if ag.Interval.Duration == 0 {
		ag.Interval.Duration = time.Minute * 5
	}

	credentials := jdcore.NewCredentials(ag.AccessKeyID, ag.AccessKeySecret)
	ag.client = jcwclient.NewMonitorClient(credentials)
	ag.client.Logger = jdcore.NewDefaultLogger(jdcore.LogError)

	//每秒最多20个请求
	limit := rate.Every(50 * time.Millisecond)
	ag.limiter = rate.NewLimiter(limit, 1)

	if err := ag.genReqs(ag.ctx); err != nil {
		return
	}

	if len(ag.reqs) == 0 {
		moduleLogger.Warnf("no metric found")
		return
	}

	select {
	case <-ag.ctx.Done():
		return
	default:
	}

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

			ag.fetchMetric(ag.ctx, req)
		}

		datakit.SleepContext(ag.ctx, time.Second*3)
	}

}

func (ag *agent) genReqs(ctx context.Context) error {

	//生成所有请求
	for _, proj := range ag.Services {

		for _, metricName := range proj.MetricNames {

			if metricName == "*" {
				continue
			}

			select {
			case <-ctx.Done():
				return context.Canceled
			default:
			}

			req, err := proj.genMetricReq(metricName)
			if err != nil {
				moduleLogger.Errorf("%s", err)
				return err
			}

			if req.interval == 0 {
				req.interval = ag.Interval.Duration
			}

			ag.reqs = append(ag.reqs, req)
		}
	}

	return nil
}

func (ag *agent) fetchMetric(ctx context.Context, req *metricsRequest) {

	now := time.Now().Truncate(time.Minute)
	end := now.Format("2006-01-02T15:04:05Z0700")
	var startTime time.Time
	if req.lastTime.IsZero() {
		startTime = now.Add(-5 * time.Minute)
	} else {
		if now.Sub(req.lastTime) < req.interval {
			return
		}
		startTime = req.lastTime.Add(-(ag.Delay.Duration))
	}

	req.to = end
	req.from = startTime.Format("2006-01-02T15:04:05Z0700")

	request := jcwapis.NewDescribeMetricDataRequest(ag.RegionID, req.metricname, req.resourceid)
	request.SetServiceCode(req.servicename)
	request.SetAggrType(req.aggreType)
	request.SetDownSampleType(req.aggreType)
	request.SetStartTime(req.from)
	request.SetEndTime(req.to)
	resp, err := ag.client.DescribeMetricData(request)

	if err != nil || resp.Error.Code != 0 {
		moduleLogger.Errorf("fail to get metric: Service=%s, ResourceID=%s, MetricName=%s, AggregateType=%s, StartTime=%v, EndTime=%v, %v", req.servicename, req.resourceid, req.metricname, req.aggreType, req.from, req.to, resp.Error)
		return
	}

	req.lastTime = now

	measurement := formatMeasurement(req.servicename)

	for _, m := range resp.Result.MetricDatas {
		moduleLogger.Debugf("get %d datapoints: Service=%s, ResourceID=%s, MetricName=%s, AggregateType=%s, StartTime=%v, EndTime=%v, Period=%s, Unit=%s", len(m.Data), req.servicename, req.resourceid, m.Metric.Metric, m.Metric.Aggregator, req.from, req.to, m.Metric.Period, m.Metric.CalculateUnit)

		for _, datapoint := range m.Data {

			select {
			case <-ctx.Done():
				return
			default:
			}

			tags := map[string]string{}
			extendTags(tags, req.tags, false)
			extendTags(tags, ag.Tags, false)
			for _, k := range m.Tags {
				tags[k.TagKey] = k.TagValue
			}
			tags["unit"] = m.Metric.CalculateUnit
			tags["period"] = m.Metric.Period

			fields := map[string]interface{}{}
			fields[formatField(req.metricname, m.Metric.Aggregator)] = datapoint.Value

			tm := time.Now()
			if datapoint.Timestamp != 0 {
				tm = time.Unix(datapoint.Timestamp/1000, 0)
			}

			if len(fields) == 0 {
				moduleLogger.Warnf("skip %s.%s datapoint for no fields, %s", req.servicename, req.metricname, datapoint)
				continue
			}

			if ag.debugMode {
				data, err := io.MakeMetric(measurement, tags, fields)
				if err != nil {
					moduleLogger.Errorf("MakeMetric failed, %s", err)
				} else {
					fmt.Printf("**** %s ****\n", string(data))
				}
			} else {
				io.NamedFeedEx(inputName, io.Metric, measurement, tags, fields, tm)
			}

		}
	}

}

func extendTags(to map[string]string, from map[string]string, override bool) {
	if from == nil {
		return
	}
	for k, v := range from {
		if !override {
			if _, exist := to[k]; exist {
				continue
			}
		}
		to[k] = v
	}
}

func formatMeasurement(project string) string {
	project = strings.Replace(project, "/", "_", -1)
	project = snakeCase(project)
	return fmt.Sprintf("%s_%s", inputName, project)
}

func formatField(metricName string, statistic string) string {
	return fmt.Sprintf("%s_%s", metricName, statistic)
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

func newAgent() *agent {
	ac := &agent{}
	ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
	return ac
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newAgent()
	})
}
