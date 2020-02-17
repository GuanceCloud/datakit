package aliyuncost

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"
	"github.com/influxdata/telegraf/selfstat"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/limiter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
)

type CostOrder struct {
	interval        time.Duration
	name            string
	runningInstance *RunningInstance
	logger          *models.Logger
}

func NewCostOrder(cfg *CostCfg, ri *RunningInstance) *CostOrder {
	c := &CostOrder{
		name:            "aliyun_cost_order",
		interval:        cfg.OrdertInterval.Duration,
		runningInstance: ri,
	}
	c.logger = &models.Logger{
		Errs: selfstat.Register("gather", "errors", nil),
		Name: `aliyuncost:order`,
	}
	return c
}

func (co *CostOrder) run(ctx context.Context) error {

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		co.getHistoryData(ctx, co.runningInstance.lmtr)
	}()

	co.getRealtimeData(ctx)

	wg.Wait()

	co.logger.Info("done")

	return nil
}

func (co *CostOrder) getRealtimeData(ctx context.Context) error {

	for {
		co.runningInstance.suspendHistoryFetch()
		now := time.Now().Truncate(time.Minute)
		start := now.Add(-co.interval)

		from := unixTimeStr(start)
		to := unixTimeStr(now)

		if err := co.getOrders(ctx, from, to, co.runningInstance.lmtr, nil); err != nil && err != context.Canceled {
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

func (co *CostOrder) getHistoryData(ctx context.Context, lmtr *limiter.RateLimiter) error {

	key := "." + co.runningInstance.cacheFileKey(`orders`)

	if !co.runningInstance.cfg.CollectHistoryData {
		return nil
	}

	co.logger.Info("start getting history Orders")

	info, _ := GetAliyunCostHistory(key)

	if info == nil {
		info = &historyInfo{}
	} else if info.Statue == 1 {
		return nil
	}

	if info.Start == "" {
		now := time.Now().Truncate(time.Minute)
		start := now.Add(-time.Hour * 8760)
		info.Start = unixTimeStr(start)
		info.End = unixTimeStr(now)
		info.Statue = 0
		info.PageNum = 0
	}

	info.key = key

	if err := co.getOrders(ctx, info.Start, info.End, lmtr, info); err != nil && err != context.Canceled {
		co.logger.Errorf("fail to get orders of history(%s-%s): %s", info.Start, info.End, err)
		return err
	}

	return nil
}

func (co *CostOrder) getOrders(ctx context.Context, start, end string, lmtr *limiter.RateLimiter, info *historyInfo) error {

	defer func() {
		recover()
	}()

	if info != nil {
		co.logger.Infof("(history)start getting Orders(%s - %s)", start, end)
	} else {
		co.logger.Infof("start getting Orders(%s - %s)", start, end)
	}

	var respOrder *bssopenapi.QueryOrdersResponse

	req := bssopenapi.CreateQueryOrdersRequest()
	req.Scheme = "https"
	req.CreateTimeStart = start
	req.CreateTimeEnd = end
	req.PageSize = requests.NewInteger(300)
	if info != nil {
		req.PageNum = requests.NewInteger(info.PageNum)
	}

	for {
		if info != nil {
			co.runningInstance.wait()
		}

		select {
		case <-ctx.Done():
			return context.Canceled
		case <-lmtr.C:
		}

		resp, err := co.runningInstance.client.QueryOrders(req)
		if err != nil {
			return fmt.Errorf("fail to get orders from %s to %s: %s", start, end, err)
		}

		if info != nil {
			co.logger.Debugf("(history)Order(%s - %s): TotalCount=%d, PageNum=%d, PageSize=%d, count=%d", start, end, resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.Items.Item))
		} else {
			co.logger.Debugf("Order(%s - %s): TotalCount=%d, PageNum=%d, PageSize=%d, count=%d", start, end, resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.Items.Item))
		}

		if respOrder == nil {
			respOrder = resp
		} else {
			respOrder.Data.OrderList.Order = append(respOrder.Data.OrderList.Order, resp.Data.OrderList.Order...)
		}

		if info != nil {
			if err = co.parseOrderResponse(ctx, respOrder); err != nil {
				return err
			}
			respOrder.Data.OrderList.Order = respOrder.Data.OrderList.Order[:0]
		}

		if resp.Data.TotalCount > 0 && resp.Data.PageNum*resp.Data.PageSize < resp.Data.TotalCount {
			req.PageNum = requests.NewInteger(resp.Data.PageNum + 1)
			if info != nil {
				info.PageNum = resp.Data.PageNum + 1
				SetAliyunCostHistory(info.key, info)
			}
		} else {
			if info != nil {
				info.Statue = 1
				SetAliyunCostHistory(info.key, info)
			}
			break
		}
	}

	if info != nil {
		co.logger.Debugf("(history)finish getting Orders(%s - %s), count=%d", start, end, len(respOrder.Data.OrderList.Order))
	} else {
		co.logger.Debugf("finish getting Orders(%s - %s), count=%d", start, end, len(respOrder.Data.OrderList.Order))
	}

	return co.parseOrderResponse(ctx, respOrder)
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
		}

		fields := map[string]interface{}{
			"OrderID":        item.OrderId,
			"RelatedOrderId": item.RelatedOrderId,
		}

		fields["PretaxGrossAmount"], _ = strconv.ParseFloat(item.PretaxGrossAmount, 64)
		fields["PretaxAmount"], _ = strconv.ParseFloat(item.PretaxAmount, 64)

		t, err := time.Parse(time.RFC3339, item.PaymentTime)
		if err != nil {
			co.logger.Errorf("fail to parse time:%v, error:%s", item.PaymentTime, err)
		} else {
			_ = t
			if co.runningInstance.cost.accumulator != nil {
				co.runningInstance.cost.accumulator.AddFields(co.getName(), fields, tags)
			}
		}

	}

	return nil
}

func (co *CostOrder) getInterval() time.Duration {
	return co.interval
}

func (co *CostOrder) getName() string {
	return co.name
}
