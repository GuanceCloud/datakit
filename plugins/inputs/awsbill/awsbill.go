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

	"github.com/influxdata/telegraf"
	uuid "github.com/satori/go.uuid"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type (
	AwsBillAgent struct {
		AwsInstances []*AwsInstance `toml:"aws_billing"`

		ctx       context.Context
		cancelFun context.CancelFunc

		accumulator telegraf.Accumulator

		logger *models.Logger

		runningInstances []*runningInstance
	}

	runningInstance struct {
		cfg *AwsInstance

		cloudwatchClient *cloudwatch.CloudWatch

		agent *AwsBillAgent

		logger *models.Logger

		metricName string

		rateLimiter *rate.Limiter

		billingMetrics map[string]*cloudwatch.Metric
	}
)

func (_ *AwsBillAgent) SampleConfig() string {
	return sampleConfig
}

func (_ *AwsBillAgent) Description() string {
	return ""
}

func (_ *AwsBillAgent) Gather(telegraf.Accumulator) error {
	return nil
}

func (_ *AwsBillAgent) Init() error {
	return nil
}

func (a *AwsBillAgent) Start(acc telegraf.Accumulator) error {

	a.logger = &models.Logger{
		Name: inputName,
	}

	if len(a.AwsInstances) == 0 {
		a.logger.Warnf("no configuration found")
		return nil
	}

	a.logger.Infof("starting...")

	a.accumulator = acc

	for _, instCfg := range a.AwsInstances {
		r := &runningInstance{
			cfg:            instCfg,
			agent:          a,
			logger:         a.logger,
			billingMetrics: make(map[string]*cloudwatch.Metric),
		}
		r.metricName = instCfg.MetricName
		if r.metricName == "" {
			r.metricName = "aws_billing"
		}

		if r.cfg.Interval.Duration == 0 {
			r.cfg.Interval.Duration = time.Hour * 6
		}

		limit := rate.Every(50 * time.Millisecond)
		r.rateLimiter = rate.NewLimiter(limit, 1)

		a.runningInstances = append(a.runningInstances, r)

		go r.run(a.ctx)
	}

	return nil
}

func (r *runningInstance) initClient() error {
	cred := credentials.NewStaticCredentials(r.cfg.AccessKey, r.cfg.AccessSecret, r.cfg.AccessToken)

	cfg := aws.NewConfig()
	cfg.WithCredentials(cred).WithRegion(r.cfg.RegionID) //.WithRegion(`cn-north-1`)

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

func (r *runningInstance) getSupportMetrics() error {

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
		r.logger.Errorf("fail to get billing metrics info, %s", err)
		return err
	}

	for i, ms := range result.Metrics {
		r.billingMetrics[fmt.Sprintf("dk%d", i)] = ms
	}

	return nil
}

func (r *runningInstance) getMetric(ctx context.Context, start, end time.Time) (*cloudwatch.GetMetricDataOutput, error) {

	params := &cloudwatch.GetMetricDataInput{
		EndTime:   aws.Time(end),
		StartTime: aws.Time(start),
	}

	for id, ms := range r.billingMetrics {
		query := &cloudwatch.MetricDataQuery{
			MetricStat: &cloudwatch.MetricStat{
				Metric: ms,
				Period: aws.Int64(int64(r.cfg.Interval.Duration / time.Second)),
				Stat:   aws.String(`Maximum`),
			},
			Id:         aws.String(id),
			ReturnData: aws.Bool(true),
		}
		params.MetricDataQueries = append(params.MetricDataQueries, query) //max 100
	}

	reqUid, _ := uuid.NewV4()

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
			r.logger.Warnf("%s", err)
			time.Sleep(tempDelay)
		} else {
			if i != 0 {
				r.logger.Debugf("retry %s successed, %d", reqUid.String(), i)
			}
			break
		}
	}

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (r *runningInstance) run(ctx context.Context) error {

	defer func() {
		if e := recover(); e != nil {
			r.logger.Errorf("panic, %v", e)
		}
	}()

	if err := r.initClient(); err != nil {
		r.logger.Errorf("fail to init client, %s", err)
		return err
	}

	if err := r.getSupportMetrics(); err != nil {
		return err
	}

	metricName := r.metricName

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		end_time := time.Now().UTC().Truncate(time.Minute)
		start_time := end_time.Add(-r.cfg.Interval.Duration)

		response, err := r.getMetric(ctx, start_time, end_time)

		if err != nil {
			r.logger.Errorf("fail to GetMetricData, %s", err)
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
						r.agent.accumulator.AddFields(metricName, fields, tags, *tm)
					}
				}
			}

		}

		internal.SleepContext(ctx, r.cfg.Interval.Duration)
	}

}

func (a *AwsBillAgent) Stop() {
	a.cancelFun()
}

func init() {
	inputs.Add(inputName, func() telegraf.Input {
		ac := &AwsBillAgent{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
