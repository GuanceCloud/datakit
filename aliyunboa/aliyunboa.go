package aliyunboa

import (
	"context"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"

	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers/influx"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/log"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/uploader"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/utils"
)

func init() {
	config.AddConfig("aliyuncost", &Cfg)
	service.Add("aliyuncost", func(logger log.Logger) service.Service {
		if len(Cfg.Boas) == 0 {
			return nil
		}

		return &AliyunBoaSvr{
			logger: logger,
		}
	})
}

var (
	batchInterval = time.Duration(5) * time.Minute
	metricPeriod  = time.Duration(5 * time.Minute)
	rateLimit     = 20
)

type (
	RunningBoa struct {
		cfg *Boa

		wg sync.WaitGroup

		uploader uploader.IUploader
		logger   log.Logger

		client *bssopenapi.Client

		modules []BoaModule

		lmtr *utils.RateLimiter

		wgSuspend sync.WaitGroup
		mutex     sync.Mutex
	}

	AliyunBoaSvr struct {
		boas   []*RunningBoa
		logger log.Logger
	}

	BoaModule interface {
		getInterval() time.Duration
		getName() string
		run(context.Context) error
	}
)

func (s *AliyunBoaSvr) Start(ctx context.Context, up uploader.IUploader) error {

	if len(Cfg.Boas) == 0 {
		return nil
	}

	s.boas = []*RunningBoa{}

	for _, c := range Cfg.Boas {
		a := &RunningBoa{
			cfg:      c,
			uploader: up,
			logger:   s.logger,
		}

		a.modules = []BoaModule{
			&BoaAccount{
				name:     "aliyun_cost_account",
				interval: c.AccountInterval.Duration,
				boa:      a,
			},
			&BoaBill{
				name:     "aliyun_cost_bill",
				interval: c.AccountInterval.Duration,
				boa:      a,
			},
			&BoaOrder{
				name:     "aliyun_cost_order",
				interval: c.AccountInterval.Duration,
				boa:      a,
			},
		}

		s.boas = append(s.boas, a)
	}

	var wg sync.WaitGroup

	s.logger.Info("Starting AliyunBoaSvr...")

	for _, c := range s.boas {
		wg.Add(1)
		go func(ac *RunningBoa) {
			defer wg.Done()

			if err := ac.Run(ctx); err != nil && err != context.Canceled {
				s.logger.Errorf("%s", err)
			}
		}(c)
	}

	wg.Wait()

	s.logger.Info("AliyunBoaSvr done")
	return nil
}

func (s *RunningBoa) suspendLastyearFetch() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.wgSuspend.Add(1)
}

func (s *RunningBoa) resumeLastyearFetch() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.wgSuspend.Done()
}

func (s *RunningBoa) wait() {
	s.wgSuspend.Wait()
}

func (s *RunningBoa) Run(ctx context.Context) error {

	s.lmtr = utils.NewRateLimiter(10, time.Minute)
	defer s.lmtr.Stop()

	var err error
	s.client, err = bssopenapi.NewClientWithAccessKey(s.cfg.RegionID, s.cfg.AccessKeyID, s.cfg.AccessKeySecret)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return context.Canceled
	default:
	}

	for _, boaModule := range s.modules {
		s.wg.Add(1)
		go func(m BoaModule, ctx context.Context) {
			defer s.wg.Done()

			m.run(ctx)

		}(boaModule, ctx)

	}

	s.wg.Wait()

	return nil
}

func addLine(metricName string, tags map[string]string, fields map[string]interface{}, tm time.Time, up uploader.IUploader, l log.Logger) error {

	serializer := influx.NewSerializer()

	m, _ := metric.New(metricName, tags, fields, tm)

	output, err := serializer.Serialize(m)
	//l.Debug(string(output))
	if err == nil {
		if up != nil {
			up.AddLog(&uploader.LogItem{
				Log: string(output),
			})
		}
	} else {
		l.Warnf("[warn] Serialize to influx protocol line fail(%s): %s; tags:%#v, fields:%#v", metricName, err, tags, fields)
	}

	return err
}
