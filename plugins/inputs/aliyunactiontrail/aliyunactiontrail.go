package aliyunactiontrail

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/actiontrail"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type (
	AliyunActiontrail struct {
		Actiontrail []*ActiontrailInstance

		ctx       context.Context
		cancelFun context.CancelFunc

		wg sync.WaitGroup

		logger *models.Logger
	}

	runningInstance struct {
		cfg *ActiontrailInstance

		agent *AliyunActiontrail

		logger *models.Logger

		client *actiontrail.Client

		metricName string

		rateLimiter *rate.Limiter
	}
)

func (_ *AliyunActiontrail) Catalog() string {
	return "aliyun"
}

func (_ *AliyunActiontrail) SampleConfig() string {
	return configSample
}

func (a *AliyunActiontrail) Run() {

	a.logger = &models.Logger{
		Name: `aliyunactiontrail`,
	}

	if len(a.Actiontrail) == 0 {
		a.logger.Warnf("no configuration found")
		return
	}

	go func() {
		<-datakit.Exit.Wait()
		a.cancelFun()
	}()

	for _, instCfg := range a.Actiontrail {
		a.wg.Add(1)
		go func(instCfg *ActiontrailInstance) {
			defer a.wg.Done()

			r := &runningInstance{
				cfg:    instCfg,
				agent:  a,
				logger: a.logger,
			}
			r.metricName = instCfg.MetricName
			if r.metricName == "" {
				r.metricName = "aliyun_actiontrail"
			}

			if r.cfg.Interval.Duration == 0 {
				r.cfg.Interval.Duration = time.Minute * 10
			}

			limit := rate.Every(40 * time.Millisecond)
			r.rateLimiter = rate.NewLimiter(limit, 1)

			r.run()

		}(instCfg)

	}

	a.wg.Wait()
}

func (r *runningInstance) getHistory() error {
	if r.cfg.From == "" {
		return nil
	}

	endTm := time.Now().Truncate(time.Minute).Add(-r.cfg.Interval.Duration)
	request := actiontrail.CreateLookupEventsRequest()
	request.Scheme = "https"
	request.StartTime = r.cfg.From
	request.EndTime = unixTimeStrISO8601(endTm)

	response, err := r.lookupEvents(request, r.client.LookupEvents)
	if err != nil {
		r.logger.Errorf("(history)LookupEvents between %s - %s failed", request.StartTime, request.EndTime)
		return err
	}

	r.handleResponse(response)

	return nil
}

func (r *runningInstance) lookupEvents(request *actiontrail.LookupEventsRequest,
	originFn func(*actiontrail.LookupEventsRequest) (*actiontrail.LookupEventsResponse, error)) (*actiontrail.LookupEventsResponse, error) {

	var response *actiontrail.LookupEventsResponse
	var err error
	var tempDelay time.Duration

	for i := 0; i < 5; i++ {
		r.rateLimiter.Wait(r.agent.ctx)
		response, err = r.client.LookupEvents(request)

		if tempDelay == 0 {
			tempDelay = time.Millisecond * 50
		} else {
			tempDelay *= 2
		}

		if max := time.Second; tempDelay > max {
			tempDelay = max
		}

		if err != nil {
			r.logger.Warnf("%s", err)
			time.Sleep(tempDelay)
		} else {
			if i != 0 {
				r.logger.Debugf("retry successed, %d", i)
			}
			break
		}
	}

	return response, err
}

func (r *runningInstance) run() error {

	defer func() {
		if e := recover(); e != nil {

		}
	}()

	cli, err := actiontrail.NewClientWithAccessKey(r.cfg.Region, r.cfg.AccessID, r.cfg.AccessKey)
	if err != nil {
		r.logger.Errorf("create client failed, %s", err)
		return err
	}
	r.client = cli

	go r.getHistory()

	startTm := time.Now().Truncate(time.Minute).Add(-r.cfg.Interval.Duration)

	for {
		select {
		case <-datakit.Exit.Wait():
			return nil
		default:
		}

		request := actiontrail.CreateLookupEventsRequest()
		request.Scheme = "https"
		request.StartTime = unixTimeStrISO8601(startTm)
		request.EndTime = unixTimeStrISO8601(startTm.Add(r.cfg.Interval.Duration))

		response, err := r.lookupEvents(request, r.client.LookupEvents)

		if err != nil {
			r.logger.Errorf("LookupEvents between %s - %s failed", request.StartTime, request.EndTime)
		}

		r.handleResponse(response)

		internal.SleepContext(r.agent.ctx, r.cfg.Interval.Duration)
		startTm = startTm.Add(r.cfg.Interval.Duration)
	}
}

func (r *runningInstance) handleResponse(response *actiontrail.LookupEventsResponse) error {

	if response == nil {
		return nil
	}

	r.logger.Debugf("%s-%s, count=%d", response.StartTime, response.EndTime, len(response.Events))

	for _, ev := range response.Events {

		select {
		case <-datakit.Exit.Wait():
			return nil
		default:
		}

		tags := map[string]string{}
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

		fields["eventId"] = ev["eventId"]
		fields["eventSource"] = ev["eventSource"]
		if ev["sourceIpAddress"] != nil {
			fields["sourceIpAddress"] = ev["sourceIpAddress"]
		}
		fields["userAgent"] = ev["userAgent"]
		fields["eventVersion"] = ev["eventVersion"]

		if userIdentity, ok := ev["userIdentity"].(map[string]interface{}); ok {
			fields["userIdentity_accountId"] = userIdentity["accountId"]
			fields["userIdentity_type"] = userIdentity["type"]
			fields["userIdentity_principalId"] = userIdentity["principalId"]

			if userName, ok := userIdentity["userName"].(string); ok {
				tags["userIdentity_userName"] = userName
			}

			if accessKeyID, ok := userIdentity["accessKeyId"].(string); ok {
				tags["userIdentity_accessKeyId"] = accessKeyID
			}
		}

		if additionalEventData, ok := ev["additionalEventData"].(map[string]interface{}); ok {
			fields["loginAccount"] = additionalEventData["loginAccount"]
			fields["isMFAChecked"] = additionalEventData["isMFAChecked"]
		}

		eventTime := ev["eventTime"].(string) //utc
		evtm, err := time.Parse(`2006-01-02T15:04:05Z`, eventTime)
		if err != nil {
			r.logger.Warnf("%s", err)
		}

		io.FeedEx(io.Metric, r.metricName, tags, fields, evtm)
	}

	return nil
}

func unixTimeStrISO8601(t time.Time) string {
	_, zoff := t.Zone()
	nt := t.Add(-(time.Duration(zoff) * time.Second))
	s := nt.Format(`2006-01-02T15:04:05Z`)
	return s
}

func init() {
	inputs.Add("aliyunactiontrail", func() inputs.Input {
		ip := &AliyunActiontrail{}
		ip.ctx, ip.cancelFun = context.WithCancel(context.Background())
		return ip
	})
}
