package aliyuncdn

import (
	"context"
	"log"
	"reflect"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cdn"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type AliyunCDN struct {
	CDN              []*CDN `toml:"cdn"`
	runningInstances []*RunningInstance
	ctx              context.Context
	cancelFun        context.CancelFunc
	accumulator      telegraf.Accumulator
	logger           *models.Logger
}

type RunningInstance struct {
	cfg             *CDN
	agent           *AliyunCDN
	logger          *models.Logger
	runningProjects []*RunningProject
	metricName      string
}

func (_ *AliyunCDN) SampleConfig() string {
	return aliyunCDNConfigSample
}

func (_ *AliyunCDN) Description() string {
	return ""
}

func (_ *AliyunCDN) Gather(telegraf.Accumulator) error {
	return nil
}

func (cdn *AliyunCDN) Start(acc telegraf.Accumulator) error {
	if len(cdn.CDN) == 0 {
		log.Printf("W! [aliyuncdn] no configuration found")
		return nil
	}

	cdn.logger = &models.Logger{
		Errs: selfstat.Register("gather", "errors", nil),
		Name: `aliyuncdn`,
	}

	log.Printf("aliyun cdn start")

	cdn.accumulator = acc

	for _, instCfg := range cdn.CDN {
		r := &RunningInstance{
			cfg:    instCfg,
			agent:  cdn,
			logger: cdn.logger,
		}
		cdn.runningInstances = append(cdn.runningInstances, r)

		go r.run(cdn.ctx)
	}

	return nil
}

func (cdn *AliyunCDN) Stop() {
	cdn.cancelFun()
}

func (r *RunningInstance) run(ctx context.Context) error {
	for _, c := range r.cfg.Actions {
		p := &RunningProject{
			cfg:     c,
			inst:    r,
			logger:  r.logger,
			mainCfg: r.cfg,
		}
		r.runningProjects = append(r.runningProjects, p)

		cli, err := cdn.NewClientWithAccessKey(r.cfg.RegionID, r.cfg.AccessKeyId, r.cfg.AccessKeySecret)

		if err != nil {
			r.logger.Error(err)
		}
		p.Client = cli

		go p.run(ctx)
	}
	return nil
}

func (r *RunningProject) run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		fv := reflect.ValueOf(&r).Elem()
		fv.MethodByName(r.cfg.ActionName).Call(nil)

		internal.SleepContext(ctx, r.mainCfg.Interval.Duration)
	}
}

func init() {
	inputs.Add("aliyuncdn", func() telegraf.Input {
		ac := &AliyunCDN{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
