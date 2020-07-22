package aliyuncost

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

type CostOrder struct {
	interval        time.Duration
	name            string
	runningInstance *runningInstance
	logger          *logger.Logger
}

func NewCostOrder(cfg *CostCfg, ri *runningInstance) *CostOrder {
	c := &CostOrder{
		name:            "aliyun_cost_order",
		interval:        cfg.OrdertInterval.Duration,
		runningInstance: ri,
		logger:          logger.SLogger(`aliyuncost:order`),
	}
	return c
}

func (co *CostOrder) run(ctx context.Context) error {

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(time.Millisecond * 10)
		co.getHistoryData(ctx)
	}()

	co.getNormalData(ctx)

	wg.Wait()

	co.logger.Info("done")

	return nil
}

func (co *CostOrder) getNormalData(ctx context.Context) error {

	for {
		co.runningInstance.suspendHistoryFetch()
		now := time.Now().Truncate(time.Minute)
		//start := now.Add(-co.interval)
		start := now.Add(-time.Hour * 24 * 30)

		from := unixTimeStr(start)
		to := unixTimeStr(now)

		if err := co.getOrders(ctx, from, to, nil); err != nil && err != context.Canceled {
			co.logger.Errorf("%s", err)
		}

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}
		co.runningInstance.resumeHistoryFetch()
		internal.SleepContext(ctx, co.interval)

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}
	}
}

func (co *CostOrder) getHistoryData(ctx context.Context) error {

	key := "." + co.runningInstance.cacheFileKey(`orders`)

	if !co.runningInstance.cfg.CollectHistoryData {
		return nil
	}

	co.logger.Info("start getting history Orders")

	info, _ := GetAliyunCostHistory(key)

	if info == nil {
		info = &historyInfo{}
	} else if info.Statue == 1 {
		co.logger.Infof("already fetched the history data")
		return nil
	}

	if info.Start == "" {
		now := time.Now().Truncate(time.Minute)
		start := now.Add(-time.Hour * 8760)
		info.Start = unixTimeStr(start)
		info.End = unixTimeStr(now)
		info.Statue = 0
		info.PageNum = 1
	}

	info.key = key

	if err := co.getOrders(ctx, info.Start, info.End, info); err != nil && err != context.Canceled {
		co.logger.Errorf("fail to get orders of history(%s-%s): %s", info.Start, info.End, err)
		return err
	}

	return nil
}

func (co *CostOrder) getOrders(ctx context.Context, start, end string, info *historyInfo) error {

	defer func() {
		recover()
	}()

	logPrefix := ""
	if info != nil {
		logPrefix = "(history) "
	}

	co.logger.Infof("%sstart getting Orders(%s - %s)", logPrefix, start, end)

	req := bssopenapi.CreateQueryOrdersRequest()
	req.Scheme = "https"
	req.CreateTimeStart = start
	req.CreateTimeEnd = end
	req.PageSize = requests.NewInteger(100)
	if info != nil {
		req.PageNum = requests.NewInteger(info.PageNum)
	} else {
		req.PageNum = requests.NewInteger(1)
	}

	for {
		if info != nil {
			co.runningInstance.wait()
		}

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		resp, err := co.runningInstance.QueryOrdersWrap(ctx, req)
		if err != nil {
			return fmt.Errorf("%sfail to get Orders(%s - %s), %s", logPrefix, start, end, err)
		}

		co.logger.Debugf("%sOrder(%s - %s): TotalCount=%d, PageNum=%d, PageSize=%d, Count=%d", logPrefix, start, end, resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.OrderList.Order))

		if err = co.parseOrderResponse(ctx, resp); err != nil {
			return err
		}

		if resp.Data.TotalCount > 0 && resp.Data.PageNum*resp.Data.PageSize < resp.Data.TotalCount {
			req.PageNum = requests.NewInteger(resp.Data.PageNum + 1)
			if info != nil {
				info.PageNum = resp.Data.PageNum + 1
				SetAliyunCostHistory(info.key, info)
			}
		} else {
			break
		}
	}

	co.logger.Debugf("%sfinish getting Orders(%s - %s)", logPrefix, start, end)

	if info != nil {
		info.Statue = 1
		SetAliyunCostHistory(info.key, info)
	}

	return nil
}

func (co *CostOrder) parseOrderResponse(ctx context.Context, resp *bssopenapi.QueryOrdersResponse) error {

	for _, item := range resp.Data.OrderList.Order {

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		tags := map[string]string{
			"ProductCode":      item.ProductCode,
			"ProductType":      item.ProductType,
			"SubscriptionType": item.SubscriptionType,
			"OrderType":        item.OrderType,
			"Currency":         item.Currency,
			"AccountName":      co.runningInstance.accountName,
			"AccountID":        co.runningInstance.accountID,
		}

		fields := map[string]interface{}{
			"OrderID":        item.OrderId,
			"RelatedOrderId": item.RelatedOrderId,
		}

		fields["PretaxGrossAmount"], _ = strconv.ParseFloat(internal.NumberFormat(item.PretaxGrossAmount), 64)
		fields["PretaxAmount"], _ = strconv.ParseFloat(internal.NumberFormat(item.PretaxAmount), 64)

		// reqDetail := bssopenapi.CreateGetOrderDetailRequest()
		// reqDetail.OrderId = item.OrderId

		// respDetail, err := co.runningInstance.client.GetOrderDetail(reqDetail)
		// if err != nil {
		// 	co.logger.Warnf("fail to get order detail of %s, %s", item.OrderId, err)
		// } else {
		// 	fields["OutstandingAmount"] = respDetail.Data.OutstandingAmount

		// }

		t, err := time.Parse(time.RFC3339, item.CreateTime)
		if err != nil {
			co.logger.Warnf("fail to parse time:%v, error:%s", item.CreateTime, err)
			continue
		}
		io.NamedFeedEx(inputName, io.Metric, co.getName(), tags, fields, t)
	}

	return nil
}

func (co *CostOrder) getInterval() time.Duration {
	return co.interval
}

func (co *CostOrder) getName() string {
	return co.name
}
