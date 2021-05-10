package awscloudtrail

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudtrail"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName    = "awscloudtrail"
	moduleLogger *logger.Logger
)

func (*AwsInstance) Catalog() string {
	return "aws"
}

func (*AwsInstance) SampleConfig() string {
	return sampleConfig
}

func (a *AwsInstance) Run() {

	moduleLogger = logger.SLogger(inputName)

	go func() {
		<-datakit.Exit.Wait()
		a.cancelFun()
	}()

	if a.Interval.Duration == 0 {
		a.Interval.Duration = time.Minute * 5
	}

	limit := rate.Every(500 * time.Millisecond)
	a.rateLimiter = rate.NewLimiter(limit, 1)

	a.run(a.ctx)
}

func (r *AwsInstance) initClient() error {
	cred := credentials.NewStaticCredentials(r.AccessKey, r.AccessSecret, "")

	cfg := aws.NewConfig()
	cfg.WithCredentials(cred).WithRegion(r.RegionID)

	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigDisable,
		Config:            *cfg,
	})

	if err != nil {
		return err
	}

	r.awsClient = cloudtrail.New(sess, aws.NewConfig())

	return nil
}

func (r *AwsInstance) run(ctx context.Context) {

	defer func() {
		if e := recover(); e != nil {
			moduleLogger.Errorf("panic, %v", e)
		}
	}()

	for {
		select {
		case <-datakit.Exit.Wait():
			return
		default:
		}

		if err := r.initClient(); err != nil {
			moduleLogger.Errorf("fail to init client, %s", err)
			if r.isTest() {
				r.testError = err
				return
			}
			time.Sleep(time.Second)
		} else {
			break
		}

	}

	measurement := `awscloudtrail`

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		startTime := time.Now().UTC().Truncate(time.Minute).Add(-r.Interval.Duration - time.Minute)

		input := &cloudtrail.LookupEventsInput{
			StartTime: aws.Time(startTime),
		}
		response, err := r.awsClient.LookupEvents(input)

		select {
		case <-ctx.Done():
			return
		default:
		}

		if err != nil {
			moduleLogger.Errorf("fail to LookupEvents, %s", err)
			if r.isTest() {
				r.testError = err
				return
			}
		} else {
			for _, evt := range response.Events {

				tags := map[string]string{}
				fields := map[string]interface{}{}

				if evt.EventSource != nil {
					tags["EventSource"] = aws.StringValue(evt.EventSource)
				}
				if evt.Username != nil {
					tags["Username"] = aws.StringValue(evt.Username)
				}
				if evt.ReadOnly != nil {
					tags["ReadOnly"] = aws.StringValue(evt.ReadOnly)
				}

				if evt.EventName != nil {
					fields["EventName"] = aws.StringValue(evt.EventName)
				}
				if evt.EventId != nil {
					fields["EventId"] = aws.StringValue(evt.EventId)
				}
				if evt.AccessKeyId != nil {
					fields["AccessKeyId"] = aws.StringValue(evt.AccessKeyId)
				}
				if evt.CloudTrailEvent != nil {
					fields["CloudTrailEvent"] = aws.StringValue(evt.CloudTrailEvent)
				}
				resources := []string{}
				for _, res := range evt.Resources {
					if res.ResourceName != nil {
						r := aws.StringValue(res.ResourceName)
						if res.ResourceType != nil {
							r += "(" + aws.StringValue(res.ResourceType) + ")"
						}
						resources = append(resources, r)
					}
				}
				fields["Resources"] = strings.Join(resources, ",")

				if r.isTest() {
					// pass
				} else if r.isDebug() {
					data, _ := io.MakeMetric(measurement, tags, fields, *evt.EventTime)
					fmt.Printf("%s\n", string(data))
				} else {
					io.NamedFeedEx(inputName, datakit.Metric, measurement, tags, fields, *evt.EventTime)
				}
			}

		}

		if r.isTest() {
			break
		}

		datakit.SleepContext(ctx, r.Interval.Duration)
	}

}

func newInstance() *AwsInstance {
	ac := &AwsInstance{}
	ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
	return ac
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newInstance()
	})
}
