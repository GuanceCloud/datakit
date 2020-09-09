package huaweiyunces

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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/sdk/huaweicloud"
)

var (
	moduleLogger *logger.Logger
	inputName    = "huaweiyunces"
)

func (*agent) SampleConfig() string {
	return sampleConfig
}

func (*agent) Catalog() string {
	return `huaweiyun`
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

	ag.client = huaweicloud.NewHWClient(ag.AccessKeyID, ag.AccessKeySecret, ag.EndPoint, ag.ProjectID, moduleLogger)

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

func (ag *agent) fetchMetric(ctx context.Context, req *metricsRequest) {

	nt := time.Now().Truncate(time.Second)
	endTime := nt.Unix() * 1000
	var startTime int64
	if req.lastTime.IsZero() {
		startTime = nt.Add(-5*time.Minute).Unix() * 1000
	} else {
		if nt.Sub(req.lastTime) < req.interval {
			return
		}
		startTime = req.lastTime.Add(-(ag.Delay.Duration)).Unix() * 1000
	}

	logEndtime := time.Unix(endTime/int64(1000), 0)
	logStarttime := time.Unix(startTime/int64(1000), 0)

	req.to = endTime
	req.from = startTime

	dms := []string{}
	for _, d := range req.dimensoions {
		dms = append(dms, fmt.Sprintf("%s,%s", d.Name, d.Value))
	}

	resData, err := ag.client.CESGetMetric(req.namespace, req.metricname, req.filter, req.period, req.from, req.to, dms)
	if err != nil {
		moduleLogger.Errorf("fail to get metric: Namespace=%s, MetricName=%s, Period=%v, StartTime=%v(%s), EndTime=%v(%s), Dimensions=%s", req.namespace, req.metricname, req.period, req.from, logStarttime, req.to, logEndtime, req.dimensoions)
		return
	}

	resp := parseMetricResponse(resData, req.filter)

	moduleLogger.Debugf("get %d datapoints: Namespace=%s, MetricName=%s, Filter=%s, Period=%v, InterVal=%v, StartTime=%v(%s), EndTime=%v(%s), Dimensions=%s", len(resp.datapoints), req.namespace, req.metricname, req.filter, req.period, req.interval, req.from, logStarttime, req.to, logEndtime, req.dimensoions)

	metricSetName := req.metricSetName
	if metricSetName == "" {
		metricSetName = formatMeasurement(req.namespace)
	}

	req.lastTime = nt

	for _, datapoint := range resp.datapoints {

		select {
		case <-ctx.Done():
			return
		default:
		}

		tags := map[string]string{}
		extendTags(tags, req.tags, false)
		extendTags(tags, ag.Tags, false)
		for _, k := range req.dimensoions {
			tags[k.Name] = k.Value
		}
		tags["unit"] = datapoint.unit

		fields := map[string]interface{}{}
		fields[formatField(req.metricname, req.filter)] = datapoint.value

		tm := time.Now()
		if datapoint.tm != 0 {
			tm = time.Unix(datapoint.tm/1000, 0)
		}

		if len(fields) == 0 {
			moduleLogger.Warnf("skip %s.%s datapoint for no fields, %s", req.namespace, req.metricname, datapoint)
			continue
		}

		/*data, err := io.MakeMetric(metricSetName, tags, fields)
		if err != nil {
			moduleLogger.Errorf("MakeMetric failed, %s", err)
		} else {
			fmt.Printf("**** %s ****\n", string(data))
		}*/

		io.NamedFeedEx(inputName, io.Metric, metricSetName, tags, fields, tm)

	}

}

func (ag *agent) genReqs(ctx context.Context) error {

	//生成所有请求
	for _, proj := range ag.Namespace {

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
