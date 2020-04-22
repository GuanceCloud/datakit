package awsbill

import (
	"context"
	"time"

	"golang.org/x/time/rate"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/influxdata/telegraf"
	uuid "github.com/satori/go.uuid"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
)

type (
	AwsBillAgent struct {
		AwsInstances []*AwsInstance

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
	}
)

func (_ *AwsBillAgent) SampleConfig() string {
	return ``
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
		Name: `awsbilling`,
	}

	if len(a.AwsInstances) == 0 {
		a.logger.Warnf("no configuration found")
		return nil
	}

	a.logger.Infof("starting...")

	a.accumulator = acc

	for _, instCfg := range a.AwsInstances {
		r := &runningInstance{
			cfg:    instCfg,
			agent:  a,
			logger: a.logger,
		}
		r.metricName = instCfg.MetricName
		if r.metricName == "" {
			r.metricName = "aws_billing"
		}

		if r.cfg.Interval.Duration == 0 {
			r.cfg.Interval.Duration = time.Minute * 5
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
	cfg.WithCredentials(cred) //.WithRegion(`cn-north-1`)

	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigDisable,
		Config:            *cfg,
	})

	if err != nil {
		return err
	}

	r.cloudwatchClient = cloudwatch.New(sess, aws.NewConfig().WithRegion(r.cfg.RegionID))

	return nil
}

func (r *runningInstance) run(ctx context.Context) error {

	defer func() {
		if e := recover(); e != nil {

		}
	}()

	if err := r.initClient(); err != nil {
		r.logger.Errorf("fail to init client, %s", err)
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		query1 := &cloudwatch.MetricDataQuery{
			MetricStat: &cloudwatch.MetricStat{
				Metric: &cloudwatch.Metric{
					MetricName: aws.String(`EstimatedCharges`),
					Namespace:  aws.String(`AWS/Billing`),
					Dimensions: []*cloudwatch.Dimension{
						&cloudwatch.Dimension{
							Name:  aws.String(`Currency`),
							Value: aws.String(`USD`),
						},
					},
				},
				Period: aws.Int64(3600),
				Stat:   aws.String(`Maximum`),
			},
			Id:         aws.String("a1"),
			ReturnData: aws.Bool(true),
		}

		params := &cloudwatch.GetMetricDataInput{
			EndTime:           aws.Time(time.Now().UTC().Truncate(time.Minute)),
			StartTime:         aws.Time(time.Now().UTC().Truncate(time.Minute).Add(-24 * time.Hour)),
			MetricDataQueries: []*cloudwatch.MetricDataQuery{query1}, //max 100
		}

		reqUid, _ := uuid.NewV4()

		var tempDelay time.Duration
		var resp *cloudwatch.GetMetricDataOutput
		var err error
		for i := 0; i < 5; i++ {
			r.rateLimiter.Wait(ctx)
			resp, err = r.cloudwatchClient.GetMetricData(params)

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
			r.logger.Errorf("fail to GetMetricData, %s", err)
		} else {

			_ = resp
		}

	}

}

func (a *AwsBillAgent) Stop() {
	a.cancelFun()
}
