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
	"golang.org/x/time/rate"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	//batchInterval = time.Duration(5) * time.Minute
	//metricPeriod  = time.Duration(5 * time.Minute)
	//rateLimit     = 20

	historyCacheDir = `/etc/datakit/aliyuncost`
)

type (
	AliyunCostAgent struct {
		Costs []*CostCfg `toml:"boa"`

		wg sync.WaitGroup

		ctx       context.Context
		cancelFun context.CancelFunc
		logger    *models.Logger
	}

	runningInstance struct {
		cfg *CostCfg

		wg sync.WaitGroup

		client *bssopenapi.Client

		modules []costModule

		wgSuspend sync.WaitGroup
		mutex     sync.Mutex

		rateLimiter *rate.Limiter

		agent *AliyunCostAgent

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

func (_ *AliyunCostAgent) Catalog() string {
	return "aliyun"
}

func (_ *AliyunCostAgent) SampleConfig() string {
	return aliyuncostConfigSample
}

// func (_ *AliyunCostAgent) Description() string {
// 	return ""
// }

func (ac *AliyunCostAgent) Run() {

	ac.logger = &models.Logger{
		Name: `aliyuncost`,
	}

	if len(ac.Costs) == 0 {
		ac.logger.Warnf("no configuration found")
		return
	}

	go func() {
		<-config.Exit.Wait()
		ac.cancelFun()
	}()

	for _, cfg := range ac.Costs {
		ac.wg.Add(1)

		go func(cfg *CostCfg) {
			defer ac.wg.Done()

			if cfg.AccountInterval.Duration == 0 {
				cfg.AccountInterval.Duration = 24 * time.Hour
			}

			if cfg.BiilInterval.Duration == 0 {
				cfg.BiilInterval.Duration = time.Hour
			}

			if cfg.OrdertInterval.Duration == 0 {
				cfg.OrdertInterval.Duration = time.Hour
			}

			ri := &runningInstance{
				cfg:   cfg,
				agent: ac,
				ctx:   ac.ctx,
			}

			limit := rate.Every(60 * time.Millisecond)
			ri.rateLimiter = rate.NewLimiter(limit, 1)

			if cfg.AccountInterval.Duration > 0 {
				ri.modules = append(ri.modules, NewCostAccount(cfg, ri))
			}

			if cfg.BiilInterval.Duration > 0 {
				ri.modules = append(ri.modules, NewCostBill(cfg, ri))
			}

			if cfg.OrdertInterval.Duration > 0 {
				ri.modules = append(ri.modules, NewCostOrder(cfg, ri))
			}

			if cfg.OrdertInterval.Duration > 0 {
				ri.modules = append(ri.modules, NewCostOrder(cfg, ri))
			}

			ri.run()

		}(cfg)
	}

	ac.wg.Wait()
}

func (s *runningInstance) suspendHistoryFetch() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.wgSuspend.Add(1)
}

func (s *runningInstance) resumeHistoryFetch() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.wgSuspend.Done()
}

func (s *runningInstance) wait() {
	s.wgSuspend.Wait()
}

func (s *runningInstance) cacheFileKey(subname string) string {
	m := md5.New()
	m.Write([]byte(s.cfg.AccessKeyID))
	m.Write([]byte(s.cfg.AccessKeySecret))
	m.Write([]byte(s.cfg.RegionID))
	m.Write([]byte(subname))
	return hex.EncodeToString(m.Sum(nil))
}

func (s *runningInstance) getAccountInfo() {
	req := bssopenapi.CreateQueryBillOverviewRequest()
	req.BillingCycle = fmt.Sprintf("%d-%d", time.Now().Year(), 1)

	resp, err := s.client.QueryBillOverview(req)
	if err != nil {
		log.Printf("E! fail to get account info, %s", err)
		return
	}

	s.accountName = resp.Data.AccountName
	s.accountID = resp.Data.AccountID
}

func (s *runningInstance) run() error {

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

	//先获取account name
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

func (r *runningInstance) QueryAccountTransactionsWrap(ctx context.Context, request *bssopenapi.QueryAccountTransactionsRequest) (response *bssopenapi.QueryAccountTransactionsResponse, err error) {
	for i := 0; i < 5; i++ {
		r.rateLimiter.Wait(ctx)
		response, err = r.client.QueryAccountTransactions(request)
		if err == nil {
			return
		}
		internal.SleepContext(ctx, time.Millisecond*200)
	}

	return
}

func (r *runningInstance) QueryAccountBalanceWrap(ctx context.Context, request *bssopenapi.QueryAccountBalanceRequest) (response *bssopenapi.QueryAccountBalanceResponse, err error) {
	for i := 0; i < 5; i++ {
		r.rateLimiter.Wait(ctx)
		response, err = r.client.QueryAccountBalance(request)
		if err == nil {
			return
		}
		internal.SleepContext(ctx, time.Millisecond*200)
	}

	return
}

func (r *runningInstance) QueryBillWrap(ctx context.Context, request *bssopenapi.QueryBillRequest) (response *bssopenapi.QueryBillResponse, err error) {
	for i := 0; i < 5; i++ {
		r.rateLimiter.Wait(ctx)
		response, err = r.client.QueryBill(request)
		if err == nil {
			return
		}
		internal.SleepContext(ctx, time.Millisecond*200)
	}

	return
}

func (r *runningInstance) QueryInstanceBillWrap(ctx context.Context, request *bssopenapi.QueryInstanceBillRequest) (response *bssopenapi.QueryInstanceBillResponse, err error) {
	for i := 0; i < 5; i++ {
		r.rateLimiter.Wait(ctx)
		response, err = r.client.QueryInstanceBill(request)
		if err == nil {
			return
		}
		internal.SleepContext(ctx, time.Millisecond*200)
	}

	return
}

func (r *runningInstance) QueryOrdersWrap(ctx context.Context, request *bssopenapi.QueryOrdersRequest) (response *bssopenapi.QueryOrdersResponse, err error) {
	for i := 0; i < 5; i++ {
		r.rateLimiter.Wait(ctx)
		response, err = r.client.QueryOrders(request)
		if err == nil {
			return
		}
		internal.SleepContext(ctx, time.Millisecond*200)
	}

	return
}

func init() {
	inputs.Add("aliyuncost", func() inputs.Input {
		ac := &AliyunCostAgent{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
