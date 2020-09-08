package awshealth

import (
	"context"
	"time"

	"golang.org/x/time/rate"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/health"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName    = "awshealth"
	moduleLogger *logger.Logger
)

func (*agent) Catalog() string {
	return "aws"
}

func (*agent) SampleConfig() string {
	return sampleConfig
}

func createAwsClient(ak, sk, region string) (*health.Health, error) {

	cred := credentials.NewStaticCredentials(ak, sk, "")

	cfg := aws.NewConfig()
	cfg.WithCredentials(cred).WithRegion(region)

	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigDisable,
		Config:            *cfg,
	})

	if err != nil {
		return nil, err
	}

	healthCli := health.New(sess, aws.NewConfig().WithRegion(region))

	return healthCli, nil
}

func (ag *agent) Run() {

	if ag.MetricName == "" {
		ag.MetricName = "awshealth"
	}

	if ag.Interval.Duration == 0 {
		ag.Interval.Duration = time.Minute
	}

	moduleLogger = logger.SLogger(inputName)

	go func() {
		<-datakit.Exit.Wait()
		ag.cancelFun()
	}()

	limit := rate.Every(40 * time.Millisecond)
	ag.rateLimiter = rate.NewLimiter(limit, 1)

	for {
		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err := createAwsClient(ag.AccessKeyID, ag.AccessKeySecret, ag.Region)
		if err == nil {
			ag.client = cli
			break
		}
		moduleLogger.Errorf("%s", err)
		datakit.SleepContext(ag.ctx, time.Second)
	}

	input := &health.DescribeEventsInput{
		Filter:    &health.EventFilter{},
		NextToken: nil,
	}

	// for _, s := range ag.Services {
	// 	input.Filter.Services = append(input.Filter.Services, aws.String(s))
	// }

	// for _, s := range ag.EventStatus {
	// 	input.Filter.EventStatusCodes = append(input.Filter.EventStatusCodes, aws.String(s))
	// }

	// for _, s := range ag.EventType {
	// 	input.Filter.EventTypeCategories = append(input.Filter.EventTypeCategories, aws.String(s))
	// }

	var lastTime time.Time

	for {
		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		nt := time.Now().UTC()

		if lastTime.IsZero() {
			lastTime = nt.Add(-ag.Interval.Duration)
		} else {
			lastTime = lastTime.Add(-time.Minute)
		}

		input.Filter.StartTimes = []*health.DateTimeRange{
			{
				From: aws.Time(lastTime),
			},
		}

		for {
			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			ag.rateLimiter.Wait(ag.ctx)
			outputs, err := ag.client.DescribeEvents(input)

			if err != nil {
				moduleLogger.Errorf("%s", err)
				break
			}

			detailsInput := &health.DescribeEventDetailsInput{}
			for _, e := range outputs.Events {
				detailsInput.EventArns = append(detailsInput.EventArns, e.Arn)
			}
			details, err := ag.client.DescribeEventDetails(detailsInput)
			if err != nil {
				moduleLogger.Errorf("fail to get event details, %s", err)
				details = nil
			}

			ag.handleOutputs(outputs, details)

			if outputs.NextToken == nil {
				break
			}
			input.NextToken = outputs.NextToken
		}

		lastTime = nt

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (ag *agent) handleOutputs(outputs *health.DescribeEventsOutput, details *health.DescribeEventDetailsOutput) {

	for _, evt := range outputs.Events {
		tags := map[string]string{
			"EventTypeCategory": aws.StringValue(evt.EventTypeCategory),
			"StatusCode":        aws.StringValue(evt.StatusCode),
			"EventTypeCode":     aws.StringValue(evt.EventTypeCode),
			"Service":           aws.StringValue(evt.Service),
			"Region":            aws.StringValue(evt.Region),
		}

		for k, v := range ag.Tags {
			tags[k] = v
		}

		fields := map[string]interface{}{
			"Arn":       aws.StringValue(evt.Arn),
			"StartTime": aws.TimeValue(evt.StartTime).String(),
			"EndTime":   aws.TimeValue(evt.EndTime).String(),
		}

		for _, d := range details.SuccessfulSet {
			if aws.StringValue(d.Event.Arn) == aws.StringValue(evt.Arn) {
				fields["Description"] = aws.StringValue(d.EventDescription.LatestDescription)
			}
		}

		io.NamedFeedEx(inputName, io.Metric, ag.MetricName, tags, fields, *evt.StartTime)
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		ip := &agent{}
		ip.ctx, ip.cancelFun = context.WithCancel(context.Background())
		return ip
	})
}
