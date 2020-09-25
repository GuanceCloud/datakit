package aliyuncost

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"
	"golang.org/x/time/rate"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	apiRetryCount = 5
)

var (
	inputName    = `aliyuncost`
	moduleLogger *logger.Logger

	historyCacheDir = ""
)

func (*agent) Catalog() string {
	return "aliyun"
}

func (*agent) SampleConfig() string {
	return sampleConfig
}

func (ag *agent) Run() {

	moduleLogger = logger.SLogger(inputName)

	go func() {
		<-datakit.Exit.Wait()
		ag.cancelFun()
	}()

	if !ag.debugMode {
		historyCacheDir = filepath.Join(datakit.DataDir, inputName)
		os.MkdirAll(historyCacheDir, 0775)
	}

	limit := rate.Every(60 * time.Millisecond)
	ag.rateLimiter = rate.NewLimiter(limit, 1)

	if ag.AccountInterval.Duration > 0 {
		if ag.AccountInterval.Duration < time.Minute {
			ag.AccountInterval.Duration = time.Minute
		}
		ag.subModules = append(ag.subModules, newCostAccount(ag))
	}

	if ag.BiilInterval.Duration > 0 {
		if ag.BiilInterval.Duration < time.Minute {
			ag.BiilInterval.Duration = time.Minute
		}
		ag.subModules = append(ag.subModules, newCostBill(ag))
	}

	if ag.OrdertInterval.Duration > 0 {
		if ag.OrdertInterval.Duration < time.Minute {
			ag.OrdertInterval.Duration = time.Minute
		}
		ag.subModules = append(ag.subModules, newCostOrder(ag))
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			return
		default:
		}

		var err error
		ag.client, err = bssopenapi.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
		if err != nil {
			moduleLogger.Errorf("%s", err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	//先获取account name
	ag.getAccountInfo()

	var wg sync.WaitGroup
	for _, m := range ag.subModules {
		wg.Add(1)
		go func(m subModule) {
			defer wg.Done()
			m.run(ag.ctx)
		}(m)
	}

	wg.Wait()
}

func (ag *agent) cacheFileKey(subname string) string {
	m := md5.New()
	m.Write([]byte(ag.AccessKeyID))
	m.Write([]byte(ag.AccessKeySecret))
	m.Write([]byte(ag.RegionID))
	m.Write([]byte(subname))
	return hex.EncodeToString(m.Sum(nil))
}

func (ag *agent) getAccountInfo() {
	req := bssopenapi.CreateQueryBillOverviewRequest()
	req.BillingCycle = fmt.Sprintf("%d-%d", time.Now().Year(), 1)

	resp, err := ag.client.QueryBillOverview(req)
	if err != nil {
		moduleLogger.Errorf("fail to get account info, %s", err)
		return
	}

	ag.accountName = resp.Data.AccountName
	ag.accountID = resp.Data.AccountID
}

func (ag *agent) queryAccountTransactionsWrap(ctx context.Context, request *bssopenapi.QueryAccountTransactionsRequest) (response *bssopenapi.QueryAccountTransactionsResponse, err error) {
	for i := 0; i < 5; i++ {
		ag.rateLimiter.Wait(ctx)
		response, err = ag.client.QueryAccountTransactions(request)
		if err == nil && !response.Success {
			err = fmt.Errorf("%s", response.String())
		}

		if err == nil {
			return
		}
		datakit.SleepContext(ctx, time.Millisecond*200)
	}

	return
}

func (ag *agent) queryAccountBalanceWrap(ctx context.Context, request *bssopenapi.QueryAccountBalanceRequest) (response *bssopenapi.QueryAccountBalanceResponse, err error) {
	for i := 0; i < 5; i++ {
		ag.rateLimiter.Wait(ctx)
		response, err = ag.client.QueryAccountBalance(request)
		if err == nil && !response.Success {
			err = fmt.Errorf("%s", response.String())
		}
		if err == nil {
			return
		}
		datakit.SleepContext(ctx, time.Millisecond*200)
	}

	return
}

func (ag *agent) queryBillWrap(ctx context.Context, request *bssopenapi.QueryBillRequest) (response *bssopenapi.QueryBillResponse, err error) {
	for i := 0; i < apiRetryCount; i++ {
		ag.rateLimiter.Wait(ctx)
		response, err = ag.client.QueryBill(request)
		if err == nil && !response.Success {
			err = fmt.Errorf("%s", response.String())
		}
		if err == nil {
			return
		}
		datakit.SleepContext(ctx, time.Millisecond*200)
	}

	return
}

// func (r *runningInstance) QueryInstanceBillWrap(ctx context.Context, request *bssopenapi.QueryInstanceBillRequest) (response *bssopenapi.QueryInstanceBillResponse, err error) {
// 	for i := 0; i < 5; i++ {
// 		r.rateLimiter.Wait(ctx)
// 		response, err = r.client.QueryInstanceBill(request)
// 		if err == nil {
// 			return
// 		}
// 		datakit.SleepContext(ctx, time.Millisecond*200)
// 	}

// 	return
// }

func (ag *agent) queryOrdersWrap(ctx context.Context, request *bssopenapi.QueryOrdersRequest) (response *bssopenapi.QueryOrdersResponse, err error) {
	for i := 0; i < apiRetryCount; i++ {
		ag.rateLimiter.Wait(ctx)
		response, err = ag.client.QueryOrders(request)
		if err == nil && !response.Success {
			err = fmt.Errorf("%s", response.String())
		}
		if err == nil {
			return
		}
		datakit.SleepContext(ctx, time.Millisecond*200)
	}

	return
}

func newAgent() *agent {
	ag := &agent{
		debugMode: false,
	}
	ag.ctx, ag.cancelFun = context.WithCancel(context.Background())
	return ag
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newAgent()
	})
}
