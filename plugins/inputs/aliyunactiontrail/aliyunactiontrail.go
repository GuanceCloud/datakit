package aliyunactiontrail

import (
	"context"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/actiontrail"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type (
	AliyunActiontrail struct {
		Actiontrail []*ActiontrailInstance

		ctx       context.Context
		cancelFun context.CancelFunc

		accumulator telegraf.Accumulator

		logger *models.Logger

		runningInstances []*runningInstance
	}

	runningInstance struct {
		cfg *ActiontrailInstance

		agent *AliyunActiontrail

		logger *models.Logger

		client *actiontrail.Client

		metricName string
	}
)

func (_ *AliyunActiontrail) SampleConfig() string {
	return configSample
}

func (_ *AliyunActiontrail) Description() string {
	return ""
}

func (_ *AliyunActiontrail) Gather(telegraf.Accumulator) error {
	return nil
}

func (_ *AliyunActiontrail) Init() error {
	return nil
}

func (a *AliyunActiontrail) Start(acc telegraf.Accumulator) error {

	a.logger = &models.Logger{
		Errs: selfstat.Register("gather", "errors", nil),
		Name: `aliyunactiontrail`,
	}

	if len(a.Actiontrail) == 0 {
		a.logger.Warnf("W! no configuration found")
		return nil
	}

	a.logger.Infof("start")

	a.accumulator = acc

	for _, instCfg := range a.Actiontrail {
		r := &runningInstance{
			cfg:    instCfg,
			agent:  a,
			logger: a.logger,
		}
		r.metricName = instCfg.MetricName
		if r.metricName == "" {
			r.metricName = "aliyun_actiontrail"
		}

		a.runningInstances = append(a.runningInstances, r)

		go r.run(a.ctx)
	}
	return nil
}

func (a *AliyunActiontrail) Stop() {
	a.cancelFun()
}

func (r *runningInstance) getHistory(ctx context.Context) error {
	if r.cfg.From == "" {
		return nil
	}

	endTm := time.Now().Truncate(time.Minute).Add(-r.cfg.Interval.Duration)

	request := actiontrail.CreateLookupEventsRequest()
	request.Scheme = "https"
	request.StartTime = r.cfg.From
	request.EndTime = unixTimeStrISO8601(endTm)

	response, err := r.client.LookupEvents(request)
	if err != nil {
		r.logger.Errorf("(history)LookupEvents failed, %s", err)
		return err
	}

	r.handleResponse(ctx, response)

	return nil
}

func (r *runningInstance) run(ctx context.Context) error {

	cli, err := actiontrail.NewClientWithAccessKey(r.cfg.Region, r.cfg.AccessID, r.cfg.AccessKey)
	if err != nil {
		r.logger.Errorf("create client failed, %s", err)
		return err
	}
	r.client = cli

	go r.getHistory(ctx)

	startTm := time.Now().Truncate(time.Minute).Add(-r.cfg.Interval.Duration)

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		request := actiontrail.CreateLookupEventsRequest()
		request.Scheme = "https"
		request.StartTime = unixTimeStrISO8601(startTm)
		request.EndTime = unixTimeStrISO8601(startTm.Add(r.cfg.Interval.Duration))

		response, err := r.client.LookupEvents(request)
		if err != nil {
			r.logger.Errorf("LookupEvents failed, %s", err)
		}

		r.handleResponse(ctx, response)

		internal.SleepContext(ctx, r.cfg.Interval.Duration)
		startTm = startTm.Add(r.cfg.Interval.Duration)
	}
}

func (r *runningInstance) handleResponse(ctx context.Context, response *actiontrail.LookupEventsResponse) error {

	if response == nil {
		return nil
	}

	r.logger.Debugf("%s-%s, count=%d", response.StartTime, response.EndTime, len(response.Events))

	for _, ev := range response.Events {

		select {
		case <-ctx.Done():
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

			if accessKeyId, ok := userIdentity["accessKeyId"].(string); ok {
				tags["userIdentity_accessKeyId"] = accessKeyId
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

		r.agent.accumulator.AddFields(r.metricName, fields, tags, evtm)
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
	inputs.Add("aliyunactiontrail", func() telegraf.Input {
		ac := &AliyunActiontrail{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
