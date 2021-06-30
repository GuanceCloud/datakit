package aliyuncost

import (
	"context"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	apiRetryCount = 60
	apiRetryWait  = time.Second * 3
)

func (ag *agent) queryBillOverview(ctx context.Context) {

	for i := 0; i < apiRetryCount; i++ {
		ag.rateLimiter.Wait(ctx)
		req := bssopenapi.CreateQueryBillOverviewRequest()
		req.BillingCycle = fmt.Sprintf("%d-%d", time.Now().Year(), 1)

		response, err := ag.client.QueryBillOverview(req)
		if err == nil && !response.Success {
			err = fmt.Errorf("%s", response.String())
		}
		if err == nil {
			ag.accountName = response.Data.AccountName
			ag.accountID = response.Data.AccountID
			break
		}

		moduleLogger.Warnf("%s", err)
		datakit.SleepContext(ctx, apiRetryWait)
	}

}

func (ag *agent) queryAccountTransactions(ctx context.Context, request *bssopenapi.QueryAccountTransactionsRequest) (response *bssopenapi.QueryAccountTransactionsResponse, err error) {
	for i := 0; i < apiRetryCount; i++ {
		ag.rateLimiter.Wait(ctx)
		response, err = ag.client.QueryAccountTransactions(request)
		if err == nil && !response.Success {
			err = fmt.Errorf("%s", response.String())
		}
		if err == nil {
			return
		}
		datakit.SleepContext(ctx, apiRetryWait)
	}

	return
}

func (ag *agent) queryAccountBalance(ctx context.Context, request *bssopenapi.QueryAccountBalanceRequest) (response *bssopenapi.QueryAccountBalanceResponse, err error) {
	for i := 0; i < apiRetryCount; i++ {
		ag.rateLimiter.Wait(ctx)
		response, err = ag.client.QueryAccountBalance(request)
		if err == nil && !response.Success {
			err = fmt.Errorf("%s", response.String())
		}
		if err == nil {
			return
		}
		datakit.SleepContext(ctx, apiRetryWait)
	}

	return
}

func (ag *agent) queryBill(ctx context.Context, request *bssopenapi.QueryBillRequest) (response *bssopenapi.QueryBillResponse, err error) {
	for i := 0; i < apiRetryCount; i++ {
		ag.rateLimiter.Wait(ctx)
		response, err = ag.client.QueryBill(request)
		if err == nil && !response.Success {
			err = fmt.Errorf("%s", response.String())
		}
		if err == nil {
			return
		}
		datakit.SleepContext(ctx, apiRetryWait)
	}

	return
}

func (ag *agent) queryInstanceBill(ctx context.Context, request *bssopenapi.QueryInstanceBillRequest) (response *bssopenapi.QueryInstanceBillResponse, err error) {
	for i := 0; i < apiRetryCount; i++ {
		ag.rateLimiter.Wait(ctx)
		response, err = ag.client.QueryInstanceBill(request)
		if err == nil && !response.Success {
			err = fmt.Errorf("%s", response.String())
		}
		if err == nil {
			return
		}
		datakit.SleepContext(ctx, apiRetryWait)
	}

	return
}

func (ag *agent) queryOrders(ctx context.Context, request *bssopenapi.QueryOrdersRequest) (response *bssopenapi.QueryOrdersResponse, err error) {
	for i := 0; i < apiRetryCount; i++ {
		ag.rateLimiter.Wait(ctx)
		response, err = ag.client.QueryOrders(request)
		if err == nil && !response.Success {
			err = fmt.Errorf("%s", response.String())
		}
		if err == nil {
			return
		}
		datakit.SleepContext(ctx, apiRetryWait)
	}

	return
}
