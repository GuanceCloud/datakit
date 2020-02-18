package aliyuncost

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"

	"github.com/influxdata/telegraf"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/limiter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	batchInterval = time.Duration(5) * time.Minute
	metricPeriod  = time.Duration(5 * time.Minute)
	rateLimit     = 20

	historyCacheDir = `/etc/datakit/aliyuncost`
)

type (
	AliyunCost struct {
		Costs []*CostCfg `toml:"boa"`

		runningInst []*RunningInstance

		tags map[string]string

		ctx       context.Context
		cancelFun context.CancelFunc

		accumulator telegraf.Accumulator
	}

	RunningInstance struct {
		cfg *CostCfg

		wg sync.WaitGroup

		client *bssopenapi.Client

		modules []costModule

		lmtr *limiter.RateLimiter

		wgSuspend sync.WaitGroup
		mutex     sync.Mutex

		cost *AliyunCost

		ctx context.Context

		accountName string
		accountID   string
	}

	costModule interface {
		getInterval() time.Duration
		getName() string
		run(context.Context) error
	}
)

func (_ *AliyunCost) SampleConfig() string {
	return aliyuncostConfigSample
}

func (_ *AliyunCost) Description() string {
	return ""
}

func (_ *AliyunCost) Gather(telegraf.Accumulator) error {
	return nil
}

func (ac *AliyunCost) Init() error {

	for _, cfg := range ac.Costs {
		if cfg.AccountInterval.Duration == 0 {
			cfg.AccountInterval.Duration = 24 * time.Hour
		}

		if cfg.BiilInterval.Duration == 0 {
			cfg.BiilInterval.Duration = time.Hour
		}

		if cfg.OrdertInterval.Duration == 0 {
			cfg.OrdertInterval.Duration = time.Hour
		}
	}

	return nil
}

func (ac *AliyunCost) Start(acc telegraf.Accumulator) error {

	if len(ac.Costs) == 0 {
		log.Printf("W! [aliyuncost] no configuration found")
		return nil
	}

	log.Printf("aliyun cost start")

	ac.accumulator = acc

	for _, cfg := range ac.Costs {

		ri := &RunningInstance{
			cfg:  cfg,
			cost: ac,
			ctx:  ac.ctx,
		}

		// if cfg.AccountInterval.Duration > 0 {
		// 	ri.modules = append(ri.modules, NewCostAccount(cfg, ri))
		// }

		// if cfg.BiilInterval.Duration > 0 {
		// 	ri.modules = append(ri.modules, NewCostBill(cfg, ri))
		// }

		if cfg.OrdertInterval.Duration > 0 {
			ri.modules = append(ri.modules, NewCostOrder(cfg, ri))
		}

		ac.runningInst = append(ac.runningInst, ri)
	}

	for _, inst := range ac.runningInst {

		go func(ri *RunningInstance) {

			if err := ri.run(); err != nil && err != context.Canceled {
				log.Printf("E! [aliyuncost] %s", err)
			}

			log.Printf("[aliyuncost] instance done")

		}(inst)
	}

	return nil
}

func (ac *AliyunCost) Stop() {
	ac.cancelFun()
}

func (s *RunningInstance) suspendHistoryFetch() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.wgSuspend.Add(1)
}

func (s *RunningInstance) resumeHistoryFetch() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.wgSuspend.Done()
}

func (s *RunningInstance) wait() {
	s.wgSuspend.Wait()
}

func (s *RunningInstance) cacheFileKey(subname string) string {
	m := md5.New()
	m.Write([]byte(s.cfg.AccessKeyID))
	m.Write([]byte(s.cfg.AccessKeySecret))
	m.Write([]byte(s.cfg.RegionID))
	m.Write([]byte(subname))
	return hex.EncodeToString(m.Sum(nil))
}

func (s *RunningInstance) getAccountInfo() {
	req := bssopenapi.CreateQueryBillOverviewRequest()
	req.BillingCycle = fmt.Sprintf("%d-%d", 2020, 1)

	resp, err := s.client.QueryBillOverview(req)
	if err != nil {
		log.Printf("E! fail to get account info, %s", err)
		return
	}

	s.accountName = resp.Data.AccountName
	s.accountID = resp.Data.AccountID
}

func (s *RunningInstance) run() error {

	s.lmtr = limiter.NewRateLimiter(10, time.Minute)
	defer s.lmtr.Stop()

	var err error
	s.client, err = bssopenapi.NewClientWithAccessKey(s.cfg.RegionID, s.cfg.AccessKeyID, s.cfg.AccessKeySecret)
	if err != nil {
		return err
	}

	select {
	case <-s.ctx.Done():
		return context.Canceled
	default:
	}

	s.getAccountInfo()

	for _, boaModule := range s.modules {
		s.wg.Add(1)
		go func(m costModule, ctx context.Context) {
			defer s.wg.Done()

			m.run(ctx)

		}(boaModule, s.ctx)

	}

	s.wg.Wait()

	return nil
}

func init() {
	inputs.Add("aliyuncost", func() telegraf.Input {
		ac := &AliyunCost{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
