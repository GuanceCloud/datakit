package aliyunactiontrail

import (
	"context"
	"encoding/json"
	"time"

	"golang.org/x/time/rate"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/actiontrail"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
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

func (a *AliyunActiontrail) Run() {

	moduleLogger = logger.SLogger(inputName)

	go func() {
		<-datakit.Exit.Wait()
		a.cancelFun()
	}()

	limit := rate.Every(40 * time.Millisecond)
	a.rateLimiter = rate.NewLimiter(limit, 1)

	if a.metricName == "" {
		a.metricName = "aliyun_actiontrail"
	}

	if a.Interval.Duration == 0 {
		a.Interval.Duration = time.Minute * 10
	}

	a.run()
}

func (a *AliyunActiontrail) getHistory() error {
	if a.From == "" {
		return nil
	}

	endTm := time.Now().Truncate(time.Minute).Add(-a.Interval.Duration)
	request := actiontrail.CreateLookupEventsRequest()
	request.Scheme = "https"
	request.StartTime = a.From
	request.EndTime = unixTimeStrISO8601(endTm)

	response, err := a.lookupEvents(request, a.client.LookupEvents)
	if err != nil {
		moduleLogger.Errorf("(history)LookupEvents between %s - %s failed", request.StartTime, request.EndTime)
		return err
	}

	a.handleResponse(response)

	return nil
}

func (a *AliyunActiontrail) lookupEvents(request *actiontrail.LookupEventsRequest,
	originFn func(*actiontrail.LookupEventsRequest) (*actiontrail.LookupEventsResponse, error)) (*actiontrail.LookupEventsResponse, error) {

	var response *actiontrail.LookupEventsResponse
	var err error
	var tempDelay time.Duration

	for i := 0; i < 5; i++ {
		a.rateLimiter.Wait(a.ctx)
		response, err = a.client.LookupEvents(request)

		if tempDelay == 0 {
			tempDelay = time.Millisecond * 50
		} else {
			tempDelay *= 2
		}

		if max := time.Second; tempDelay > max {
			tempDelay = max
		}

		if err != nil {
			moduleLogger.Warnf("%s", err)
			time.Sleep(tempDelay)
		} else {
			if i != 0 {
				moduleLogger.Debugf("retry successed, %d", i)
			}
			break
		}
	}

	return response, err
}

func (r *AliyunActiontrail) run() error {

	defer func() {
		if e := recover(); e != nil {
			moduleLogger.Errorf("panic error, %v", e)
		}
	}()

	for {

		select {
		case <-datakit.Exit.Wait():
			return nil
		default:
		}

		cli, err := actiontrail.NewClientWithAccessKey(r.Region, r.AccessID, r.AccessKey)
		if err != nil {
			moduleLogger.Errorf("create client failed, %s", err)
			time.Sleep(time.Second)
		} else {
			r.client = cli
			break
		}

	}

	go r.getHistory()

	startTm := time.Now().Truncate(time.Minute).Add(-r.Interval.Duration)

	for {
		select {
		case <-datakit.Exit.Wait():
			return nil
		default:
		}

		request := actiontrail.CreateLookupEventsRequest()
		request.Scheme = "https"
		request.StartTime = unixTimeStrISO8601(startTm)
		request.EndTime = unixTimeStrISO8601(startTm.Add(r.Interval.Duration))

		response, err := r.lookupEvents(request, r.client.LookupEvents)

		if err != nil {
			moduleLogger.Errorf("LookupEvents between %s - %s failed", request.StartTime, request.EndTime)
		}

		r.handleResponse(response)

		datakit.SleepContext(r.ctx, r.Interval.Duration)
		startTm = startTm.Add(r.Interval.Duration)
	}
}

func (r *AliyunActiontrail) handleResponse(response *actiontrail.LookupEventsResponse) error {

	if response == nil {
		return nil
	}

	moduleLogger.Debugf("%s-%s, count=%d", response.StartTime, response.EndTime, len(response.Events))

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
			moduleLogger.Warnf("%s", err)
		}

		evdata, _ := json.Marshal(&ev)
		fields["__content"] = string(evdata)

		io.NamedFeedEx(inputName, io.Logging, r.metricName, tags, fields, evtm)
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
	inputs.Add(inputName, func() inputs.Input {
		ip := &AliyunActiontrail{}
		ip.ctx, ip.cancelFun = context.WithCancel(context.Background())
		return ip
	})
}
