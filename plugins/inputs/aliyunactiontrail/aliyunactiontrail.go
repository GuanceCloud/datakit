package aliyunactiontrail

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/actiontrail"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName    = `aliyunactiontrail`
	moduleLogger *logger.Logger
)

func (_ *AliyunActiontrail) Catalog() string {
	return "aliyun"
}

func (_ *AliyunActiontrail) SampleConfig() string {
	return configSample
}

func (a *AliyunActiontrail) PipelineConfig() map[string]string {
	return map[string]string{
		inputName: pipelineSample,
	}
}

// TODO
func (*AliyunActiontrail) RunPipeline() {
}

func (a *AliyunActiontrail) Run() {

	moduleLogger = logger.SLogger(inputName)

	if !a.isTest() {
		go func() {
			<-datakit.Exit.Wait()
			a.cancelFun()
		}()
	}

	limit := rate.Every(40 * time.Millisecond)
	a.rateLimiter = rate.NewLimiter(limit, 1)

	if a.MetricName == "" {
		a.MetricName = "aliyun_actiontrail"
	}

	if a.Interval.Duration == 0 {
		a.Interval.Duration = time.Minute * 10
	}

	a.regions = strings.Split(a.Region, ",")

	a.run()
}

func (a *AliyunActiontrail) getHistory() {
	if a.From == "" {
		return
	}
	endTm := time.Now().UTC().Truncate(time.Minute).Add(-a.Interval.Duration)
	endTime := timeStrISO8601(endTm)
	for _, region := range a.regions {
		a.lookupEvents(a.From, endTime, region, true)
	}
}

func (a *AliyunActiontrail) lookupEvents(from, to string, region string, history bool) error {

	request := actiontrail.CreateLookupEventsRequest()
	request.Scheme = "https"
	request.StartTime = from
	request.EndTime = to
	request.RegionId = region

	var response *actiontrail.LookupEventsResponse
	var err error

	prefix := ""
	if history {
		prefix = "(history)"
	}

	for {
		var tempDelay time.Duration

		select {
		case <-a.ctx.Done():
			return nil
		default:
		}

		if history {
			for atomic.LoadInt32(&a.historyFlag) == 1 {
				select {
				case <-a.ctx.Done():
					return nil
				default:
				}
				datakit.SleepContext(a.ctx, 3*time.Second)
			}
		}

		for i := 0; i < 5; i++ {

			select {
			case <-a.ctx.Done():
				return nil
			default:
			}

			if history {
				for atomic.LoadInt32(&a.historyFlag) == 1 {
					select {
					case <-a.ctx.Done():
						return nil
					default:
					}
					datakit.SleepContext(a.ctx, 3*time.Second)
				}
			}

			a.rateLimiter.Wait(a.ctx)
			response, err = a.client.LookupEvents(request)

			if err == nil {
				break
			}
			moduleLogger.Errorf("%sFail to LookupEvents(%s) %s - %s, %s", prefix, region, request.StartTime, request.EndTime, err)
			if a.isTest() {
				break
			}

			if tempDelay == 0 {
				tempDelay = time.Millisecond * 200
			} else {
				tempDelay *= 2
			}

			if max := time.Second; tempDelay > max {
				tempDelay = max
			}

			datakit.SleepContext(a.ctx, tempDelay)
		}

		if err == nil {
			moduleLogger.Debugf("%sLookupEvents(%s) %s - %s: count=%d, NextToken=%s", prefix, region, request.StartTime, request.EndTime, len(response.Events), response.NextToken)
			a.handleResponse(response, history)
		}

		if err != nil || response.NextToken == "" {
			break
		}
		request.NextToken = response.NextToken
	}

	return err
}

func (r *AliyunActiontrail) run() error {

	defer func() {
		if e := recover(); e != nil {
			moduleLogger.Errorf("panic error, %v", e)
		}
	}()

	var err error
	r.pipelineScript, err = pipeline.NewPipeline(r.Pipeline)
	if err != nil {
		moduleLogger.Errorf("%s", err)
	}

	for {

		select {
		case <-datakit.Exit.Wait():
			return nil
		default:
		}

		region := ""
		if len(r.regions) > 0 {
			region = r.regions[0]
		}
		cli, err := actiontrail.NewClientWithAccessKey(region, r.AccessID, r.AccessKey)
		if err != nil {
			moduleLogger.Errorf("create client failed, %s", err)
			if r.isTest() {
				r.testError = err
				return err
			}
			time.Sleep(time.Second)
		} else {
			r.client = cli
			break
		}

	}

	if !r.isTest() {
		go r.getHistory()
	}

	var lastTime time.Time

	for {
		select {
		case <-datakit.Exit.Wait():
			return nil
		default:
		}

		//暂停历史数据抓取
		atomic.AddInt32(&r.historyFlag, 1)

		now := time.Now().UTC().Truncate(time.Minute)
		if lastTime.IsZero() {
			lastTime = now.Add(-r.Interval.Duration)
		}
		from := timeStrISO8601(lastTime)
		to := timeStrISO8601(now)

		apiStart := time.Now()
		for _, region := range r.regions {
			err := r.lookupEvents(from, to, region, false)
			if r.isTest() && err != nil {
				r.testError = err
				return err
			}
		}

		if r.isTest() {
			break
		}

		used := time.Now().Sub(apiStart)
		toSleep := r.Interval.Duration
		if toSleep > used {
			atomic.AddInt32(&r.historyFlag, -1)
			toSleep = toSleep - used
			datakit.SleepContext(r.ctx, toSleep)
		}

		lastTime = now
	}

	return nil
}

func (r *AliyunActiontrail) handleResponse(response *actiontrail.LookupEventsResponse, history bool) error {

	if response == nil {
		return nil
	}

	for _, ev := range response.Events {

		select {
		case <-datakit.Exit.Wait():
			return nil
		default:
		}

		tags := map[string]string{}
		tags["source"] = "aliyun_actiontrail"

		fields := map[string]interface{}{}

		if eventType, ok := ev["eventType"].(string); ok {
			tags["eventType"] = eventType
		}

		if acsRegion, ok := ev["acsRegion"].(string); ok {
			tags["region"] = acsRegion
		}

		if serviceName, ok := ev["serviceName"].(string); ok {
			tags["serviceName"] = serviceName
		}

		eventTime := ev["eventTime"].(string) //utc
		evtm, err := time.Parse(`2006-01-02T15:04:05Z`, eventTime)
		if err != nil {
			moduleLogger.Warnf("%s", err)
		}

		evdata, err := json.Marshal(&ev)
		if err != nil {
			moduleLogger.Errorf("%s", err)
			continue
		}

		if r.pipelineScript != nil {
			if result, err := r.pipelineScript.Run(string(evdata)).Result(); err != nil {
				moduleLogger.Warnf("%s", err)
			} else {
				fields = result
			}
		}

		fields["message"] = string(evdata)

		if r.isTest() {
			// pass
		} else if r.isDebug() {
			data, _ := io.MakeMetric(r.MetricName, tags, fields, evtm)
			fmt.Printf("-----%s\n", string(data))
		} else {
			io.NamedFeedEx(inputName, datakit.Logging, r.MetricName, tags, fields, evtm)
		}
	}

	return nil
}

func timeStrISO8601(t time.Time) string {
	return t.Format(`2006-01-02T15:04:05Z`)
}

func newAgent() *AliyunActiontrail {
	ag := &AliyunActiontrail{}
	ag.ctx, ag.cancelFun = context.WithCancel(context.Background())
	return ag
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newAgent()
	})
}
