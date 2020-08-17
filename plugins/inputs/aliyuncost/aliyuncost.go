package aliyuncost

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"
	"golang.org/x/time/rate"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	moduleLogger *logger.Logger

	historyCacheDir = `/etc/datakit/aliyuncost`

	inputName = `aliyuncost`
)

type (
	runningInstance struct {
		cfg *CostCfg

		wg sync.WaitGroup

		client *bssopenapi.Client

		modules []costModule

		wgSuspend sync.WaitGroup
		mutex     sync.Mutex

		rateLimiter *rate.Limiter

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

func (_ *CostCfg) Catalog() string {
	return "aliyun"
}

func (_ *CostCfg) SampleConfig() string {
	return aliyuncostConfigSample
}

func (ac *CostCfg) Run() {

	moduleLogger = logger.SLogger(inputName)

	go func() {
		<-datakit.Exit.Wait()
		ac.cancelFun()
	}()

	if ac.AccountInterval.Duration == 0 {
		ac.AccountInterval.Duration = 24 * time.Hour
	}

	if ac.BiilInterval.Duration == 0 {
		ac.BiilInterval.Duration = time.Hour
	}

	if ac.OrdertInterval.Duration == 0 {
		ac.OrdertInterval.Duration = time.Hour
	}

	ri := &runningInstance{
		cfg: ac,
		ctx: ac.ctx,
	}

	limit := rate.Every(60 * time.Millisecond)
	ri.rateLimiter = rate.NewLimiter(limit, 1)

	if ac.AccountInterval.Duration > 0 {
		ri.modules = append(ri.modules, NewCostAccount(ac, ri))
	}

	if ac.BiilInterval.Duration > 0 {
		ri.modules = append(ri.modules, NewCostBill(ac, ri))
	}

	if ac.OrdertInterval.Duration > 0 {
		ri.modules = append(ri.modules, NewCostOrder(ac, ri))
	}

	if ac.OrdertInterval.Duration > 0 {
		ri.modules = append(ri.modules, NewCostOrder(ac, ri))
	}

	ri.run()
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
		moduleLogger.Errorf("fail to get account info, %s", err)
		return
	}

	s.accountName = resp.Data.AccountName
	s.accountID = resp.Data.AccountID
}

func (s *runningInstance) run() error {

	var err error

	for {
		select {
		case <-datakit.Exit.Wait():
			return nil
		default:
		}

		s.client, err = bssopenapi.NewClientWithAccessKey(s.cfg.RegionID, s.cfg.AccessKeyID, s.cfg.AccessKeySecret)
		if err != nil {
			moduleLogger.Errorf("%s", err)
			time.Sleep(time.Second)
		} else {
			break
		}
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
		datakit.SleepContext(ctx, time.Millisecond*200)
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
		datakit.SleepContext(ctx, time.Millisecond*200)
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
		datakit.SleepContext(ctx, time.Millisecond*200)
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
		datakit.SleepContext(ctx, time.Millisecond*200)
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
		datakit.SleepContext(ctx, time.Millisecond*200)
	}

	return
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		ac := &CostCfg{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
