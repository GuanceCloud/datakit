package awsbill

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/time/rate"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName    = "aws_billing"
	moduleLogger *logger.Logger
)

func (_ *AwsInstance) Catalog() string {
	return "aws"
}

func (_ *AwsInstance) SampleConfig() string {
	return sampleConfig
}

// func (_ *AwsBillAgent) Description() string {
// 	return ""
// }

func (a *AwsInstance) Run() {

	moduleLogger = logger.SLogger(inputName)

	go func() {
		<-datakit.Exit.Wait()
		a.cancelFun()
	}()

	a.billingMetrics = make(map[string]*cloudwatch.Metric)

	if a.MetricName == "" {
		a.MetricName = "aws_billing"
	}

	if a.Interval.Duration == 0 {
		a.Interval.Duration = time.Hour * 6
	}

	limit := rate.Every(50 * time.Millisecond)
	a.rateLimiter = rate.NewLimiter(limit, 1)

	a.run(a.ctx)
}

func (r *AwsInstance) initClient() error {
	cred := credentials.NewStaticCredentials(r.AccessKey, r.AccessSecret, r.AccessToken)

	cfg := aws.NewConfig()
	cfg.WithCredentials(cred).WithRegion(r.RegionID) //.WithRegion(`cn-north-1`)

	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigDisable,
		Config:            *cfg,
	})

	if err != nil {
		return err
	}

	r.cloudwatchClient = cloudwatch.New(sess, aws.NewConfig().WithRegion(`us-east-1`))

	return nil
}

func (r *AwsInstance) getSupportMetrics() error {

	namespace := `AWS/Billing`

	var token *string
	params := &cloudwatch.ListMetricsInput{
		Namespace:  aws.String(namespace),
		NextToken:  token,
		MetricName: aws.String(`EstimatedCharges`),
	}

	var err error
	var result *cloudwatch.ListMetricsOutput
	for i := 0; i < 3; i++ {
		result, err = r.cloudwatchClient.ListMetrics(params)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		moduleLogger.Errorf("fail to get billing metrics info, %s", err)
		return err
	}

	for i, ms := range result.Metrics {
		r.billingMetrics[fmt.Sprintf("dk%d", i)] = ms
	}

	return nil
}

func (r *AwsInstance) getMetric(ctx context.Context, start, end time.Time) (*cloudwatch.GetMetricDataOutput, error) {

	params := &cloudwatch.GetMetricDataInput{
		EndTime:   aws.Time(end),
		StartTime: aws.Time(start),
	}

	for id, ms := range r.billingMetrics {
		query := &cloudwatch.MetricDataQuery{
			MetricStat: &cloudwatch.MetricStat{
				Metric: ms,
				Period: aws.Int64(int64(r.Interval.Duration / time.Second)),
				Stat:   aws.String(`Maximum`),
			},
			Id:         aws.String(id),
			ReturnData: aws.Bool(true),
		}
		params.MetricDataQueries = append(params.MetricDataQueries, query) //max 100
	}

	var tempDelay time.Duration
	var response *cloudwatch.GetMetricDataOutput
	var err error
	for i := 0; i < 5; i++ {
		r.rateLimiter.Wait(ctx)
		response, err = r.cloudwatchClient.GetMetricData(params)

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

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (r *AwsInstance) run(ctx context.Context) error {

	defer func() {
		if e := recover(); e != nil {
			moduleLogger.Errorf("panic, %v", e)
		}
	}()

	for {
		select {
		case <-datakit.Exit.Wait():
			return nil
		default:
		}

		if err := r.initClient(); err != nil {
			moduleLogger.Errorf("fail to init client, %s", err)
			time.Sleep(time.Second)
		} else {
			break
		}

	}

	if err := r.getSupportMetrics(); err != nil {
		return err
	}

	metricName := r.MetricName

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		end_time := time.Now().UTC().Truncate(time.Minute)
		start_time := end_time.Add(-r.Interval.Duration)

		response, err := r.getMetric(ctx, start_time, end_time)

		if err != nil {
			moduleLogger.Errorf("fail to GetMetricData, %s", err)
		} else {
			for _, res := range response.MetricDataResults {
				if *res.StatusCode != cloudwatch.StatusCodeComplete {
					continue
				}

				if res.Id == nil {
					continue
				}

				if ms, ok := r.billingMetrics[*res.Id]; ok {
					for idx, tm := range res.Timestamps {
						tags := map[string]string{}
						fields := map[string]interface{}{
							"estimated_charges": *res.Values[idx],
						}

						for _, dm := range ms.Dimensions {
							tags[*dm.Name] = *dm.Value
						}

						io.NamedFeedEx(inputName, io.Metric, metricName, tags, fields, *tm)
					}
				}
			}

		}

		internal.SleepContext(ctx, r.Interval.Duration)
	}

}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		ac := &AwsInstance{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
