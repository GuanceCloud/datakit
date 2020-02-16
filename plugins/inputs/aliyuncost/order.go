package aliyuncost

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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
		//b.getLastyearData(ctx, b.runningInstance.lmtr)
	}()

	co.getRealtimeData(ctx)

	wg.Wait()

	co.logger.Info("done")

	return nil
}

func (co *CostOrder) getRealtimeData(ctx context.Context) error {

	for {
		//co.runningInstance.suspendLastyearFetch()

		now := time.Now().Truncate(time.Minute)
		start := now.Add(-co.interval)

		from := strings.Replace(start.Format(time.RFC3339), "+", "Z", -1)
		to := strings.Replace(now.Format(time.RFC3339), "+", "Z", -1)

		if err := co.getOrders(ctx, from, to, co.runningInstance.lmtr, false); err != nil && err != context.Canceled {
			co.logger.Errorf("%s", err)
		}

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}
		//co.runningInstance.resumeLastyearFetch()
		internal.SleepContext(ctx, co.interval)

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}
	}
}

func (co *CostOrder) getLastyearData(ctx context.Context, lmtr *limiter.RateLimiter) error {

	if !co.runningInstance.cfg.CollectHistoryData {
		return nil
	}

	// log.Printf("I! [aliyunboa:order] start get orders of last year")

	// m := md5.New()
	// m.Write([]byte(b.runningInstance.cfg.AccessKeyID))
	// m.Write([]byte(b.runningInstance.cfg.AccessKeySecret))
	// m.Write([]byte(b.runningInstance.cfg.RegionID))
	// m.Write([]byte(`orders`))
	// k1 := hex.EncodeToString(m.Sum(nil))
	// k1 = "." + k1

	// orderFlag, _ := config.GetLastyearFlag(k1)

	// if orderFlag == 1 {
	// 	return nil
	// }

	// now := time.Now().Truncate(time.Minute).Add(-b.runningInstance.cfg.OrdertInterval.Duration)
	// start := now.Add(-time.Hour * 8760).Format(`2006-01-02T15:04:05Z`)
	// end := now.Format(`2006-01-02T15:04:05Z`)
	// if err := b.getOrders(ctx, start, end, lmtr, true); err != nil && err != context.Canceled {
	// 	log.Printf("E! [aliyunboa:order] fail to get orders of last year(%s-%s): %s", start, end, err)
	// 	return err
	// }

	// config.SetLastyearFlag(k1, 1)

	return nil
}

func (co *CostOrder) getOrders(ctx context.Context, start, end string, lmtr *limiter.RateLimiter, fromLastyear bool) error {

	defer func() {
		recover()
	}()

	co.logger.Infof("start getting Orders(%s - %s)", start, end)

	var respOrder *bssopenapi.QueryOrdersResponse

	req := bssopenapi.CreateQueryOrdersRequest()
	req.Scheme = "https"
	req.CreateTimeStart = start
	req.CreateTimeEnd = end
	req.PageSize = requests.NewInteger(300)

	for {
		if fromLastyear {
			//co.runningInstance.wait()
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

		co.logger.Debugf("Order(%s - %s): TotalCount=%d, PageNum=%d, PageSize=%d, count=%d", start, end, resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.Items.Item))

		if respOrder == nil {
			respOrder = resp
		} else {
			respOrder.Data.OrderList.Order = append(respOrder.Data.OrderList.Order, resp.Data.OrderList.Order...)
		}

		if resp.Data.TotalCount > 0 && resp.Data.PageNum*resp.Data.PageSize < resp.Data.TotalCount {
			req.PageNum = requests.NewInteger(resp.Data.PageNum + 1)
		} else {
			break
		}
	}

	co.logger.Debugf("finish getting Orders(%s - %s), count=%d", start, end, len(respOrder.Data.OrderList.Order))

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
