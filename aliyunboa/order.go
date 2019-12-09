package aliyunboa

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/utils"
)

type BoaOrder struct {
	interval time.Duration
	name     string
	boa      *RunningBoa
}

func (b *BoaOrder) run(ctx context.Context) error {

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		b.getLastyearData(ctx, b.boa.lmtr)
	}()

	b.getRealtimeData(ctx)

	wg.Wait()

	b.boa.logger.Info("aliyun order fetcher done")

	return nil
}

func (b *BoaOrder) getInterval() time.Duration {
	return b.interval
}

func (b *BoaOrder) getName() string {
	return b.name
}

func (b *BoaOrder) getRealtimeData(ctx context.Context) error {

	ticker := time.NewTicker(b.boa.cfg.BiilInterval.Duration)
	defer ticker.Stop()

	b.boa.suspendLastyearFetch()
	now := time.Now().Truncate(time.Minute)
	start := now.Add(-b.boa.cfg.OrdertInterval.Duration).Format(`2006-01-02T15:04:05Z`)
	end := now.Format(`2006-01-02T15:04:05Z`)
	if err := b.getOrders(ctx, start, end, b.boa.lmtr, false); err != nil && err != context.Canceled {
		b.boa.logger.Errorf("%s", err)
	}
	b.boa.resumeLastyearFetch()

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		case <-ticker.C:
			b.boa.suspendLastyearFetch()
			now := time.Now()
			start := now.Add(-b.boa.cfg.OrdertInterval.Duration).Format(`2006-01-02T15:04:05Z`)
			end := now.Format(`2006-01-02T15:04:05Z`)
			if err := b.getOrders(ctx, start, end, b.boa.lmtr, false); err != nil && err != context.Canceled {
				b.boa.logger.Errorf("%s", err)
			}
			b.boa.resumeLastyearFetch()
		}
	}
}

func (b *BoaOrder) getLastyearData(ctx context.Context, lmtr *utils.RateLimiter) error {

	if !b.boa.cfg.CollectHistoryData {
		return nil
	}

	b.boa.logger.Info("start get orders of last year")

	m := md5.New()
	m.Write([]byte(b.boa.cfg.AccessKeyID))
	m.Write([]byte(b.boa.cfg.AccessKeySecret))
	m.Write([]byte(b.boa.cfg.RegionID))
	m.Write([]byte(`orders`))
	k1 := hex.EncodeToString(m.Sum(nil))
	k1 = "." + k1

	orderFlag, _ := config.GetLastyearFlag(k1)

	if orderFlag == 1 {
		return nil
	}

	now := time.Now().Truncate(time.Minute).Add(-b.boa.cfg.OrdertInterval.Duration)
	start := now.Add(-time.Hour * 8760).Format(`2006-01-02T15:04:05Z`)
	end := now.Format(`2006-01-02T15:04:05Z`)
	if err := b.getOrders(ctx, start, end, lmtr, true); err != nil && err != context.Canceled {
		b.boa.logger.Errorf("fail to get orders of last year(%s-%s): %s", start, end, err)
		return err
	}

	config.SetLastyearFlag(k1, 1)

	return nil
}

func (b *BoaOrder) getOrders(ctx context.Context, start, end string, lmtr *utils.RateLimiter, fromLastyear bool) error {

	defer func() {
		recover()
	}()

	b.boa.logger.Infof("start get orders from %s to %s", start, end)

	var respOrder *bssopenapi.QueryOrdersResponse

	req := bssopenapi.CreateQueryOrdersRequest()
	req.CreateTimeStart = start
	req.CreateTimeEnd = end
	req.PageSize = requests.NewInteger(300)

	for {
		if fromLastyear {
			b.boa.wait()
		}

		select {
		case <-ctx.Done():
			return context.Canceled
		case <-lmtr.C:
		}

		resp, err := b.boa.client.QueryOrders(req)
		if err != nil {
			return fmt.Errorf("fail to get orders from %s to %s: %s", start, end, err)
		} else {
			b.boa.logger.Debugf("[Order] %s - %s: TotalCount=%d, PageNum=%d, PageSize=%d, count=%d", start, end, resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.Items.Item))

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
	}

	b.boa.logger.Debugf("finish getting Orders between %s-%s finish count=%d", start, end, len(respOrder.Data.OrderList.Order))

	return b.parseOrderResponse(ctx, respOrder)
}

func (b *BoaOrder) parseOrderResponse(ctx context.Context, resp *bssopenapi.QueryOrdersResponse) error {

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
			b.boa.logger.Errorf("fail to parse time %s, error:%s", item.PaymentTime, err)
		} else {
			addLine(b.getName(), tags, fields, t, b.boa.uploader, b.boa.logger)
		}

	}

	return nil
}
